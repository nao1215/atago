package ptyrun

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/nao1215/atago/internal/runner"
)

// The part of the kitty graphics protocol a program needs to learn that the
// terminal can draw images, and then to draw them.
//
// A program that shows images inline first asks whether the terminal supports
// them, and a program that REQUIRES them refuses to start when the answer is
// no. The plain pty answers nothing, so such a program can only be tested
// refusing. With `graphics: kitty` the pty answers the way a kitty-protocol
// terminal does: the capability query gets its OK, the cell-size query
// (CSI 16 t) gets a fixed 10x20-pixel cell, and every image the program
// transmits is decoded and kept for the `screen.images` assertion.
//
// Only direct transmission (t=d) is decoded; file and shared-memory media
// cannot be read from a transcript. The recorded set is what is on screen
// now. An image joins it when it is transmitted (a=t or a=T): whether the
// program displays it at once (a=T) or later (a=p, or a unicode placeholder
// with U=1), the image it sent is the same. A transmission that reuses an id
// replaces that image. A delete command (a=d) takes the images it names off
// the screen: a lower-case target only hides them, so a later a=p shows them
// again, and an upper-case target frees them. The targets followed are all
// (a, the default), image id (i, with or without a placement id), image
// number (n), and id range (r). The others name a screen position, a
// z-index, or animation frames, which a transcript does not track, so they
// change nothing.
const (
	// graphicsCellWidth / graphicsCellHeight are the pixel size of one cell the
	// emulated terminal reports.
	graphicsCellWidth  = 10
	graphicsCellHeight = 20

	// maxGraphicsCommand bounds one APC sequence. kitty caps a chunk at 4096
	// payload bytes; this leaves generous room for a program that does not
	// chunk while keeping a runaway sequence from exhausting memory.
	maxGraphicsCommand = 16 << 20
	// maxGraphicsPayload bounds one image's accumulated (chunked) payload.
	maxGraphicsPayload = 64 << 20
	// maxGraphicsImages bounds how many images a session keeps; images past it
	// are not recorded.
	maxGraphicsImages = 1024
)

// kittyGraphics scans the program's output for kitty graphics commands,
// answers queries, and records transmitted images. It is fed the raw pty
// stream, which reads may split anywhere, so it keeps its parse state across
// reads.
type kittyGraphics struct {
	state   apcState
	buf     []byte
	pending *kittyTransfer
	// replies are left for the caller to write, so they go out in order with
	// the rest of the terminal's answers.
	replies [][]byte

	mu     sync.Mutex
	images []kittyImage
}

// kittyImage is one recorded image with the names a program can later delete
// or place it by.
type kittyImage struct {
	id     string // i=, "" when the program named none
	number string // I=, "" when the program named none
	img    runner.TerminalImage
	hidden bool // deleted with a lower-case target: off screen, data kept
}

type apcState uint8

const (
	apcNormal apcState = iota
	apcESC
	apcBody
	apcBodyESC
	apcSkip
	apcSkipESC
)

// kittyTransfer is an image whose payload arrives over several chunks.
type kittyTransfer struct {
	keys map[string]string
	// data is the decoded payload so far. Each chunk is base64-decoded on
	// arrival: the protocol pads chunks independently, so the concatenated
	// text is not one valid base64 string.
	data   []byte
	tooBig bool
	badB64 bool
}

// add appends one chunk's payload.
func (t *kittyTransfer) add(payload []byte) {
	if t.tooBig || t.badB64 {
		return
	}
	chunk, err := decodeBase64Lenient(payload)
	if err != nil {
		t.badB64 = true
		return
	}
	if len(t.data)+len(chunk) > maxGraphicsPayload {
		t.tooBig = true
		return
	}
	t.data = append(t.data, chunk...)
}

func newKittyGraphics() *kittyGraphics {
	return &kittyGraphics{}
}

// step advances the scanner by one byte of program output and reports whether
// that byte completed an APC string. If the string was a graphics command, its
// replies (if any) are then waiting in takeReplies.
func (g *kittyGraphics) step(b byte) bool {
	switch g.state {
	case apcNormal:
		if b == 0x1b {
			g.state = apcESC
		}
	case apcESC:
		switch b {
		case '_':
			g.state = apcBody
			g.buf = g.buf[:0]
		case 0x1b:
		default:
			g.state = apcNormal
		}
	case apcBody, apcSkip:
		if b == 0x1b {
			g.state++ // apcBodyESC / apcSkipESC
			return false
		}
		if g.state == apcBody {
			if len(g.buf) >= maxGraphicsCommand {
				g.state = apcSkip
				return false
			}
			g.buf = append(g.buf, b)
		}
	case apcBodyESC, apcSkipESC:
		if b == '\\' {
			complete := g.state == apcBodyESC
			g.state = apcNormal
			if complete {
				g.command(g.buf)
			}
			return complete
		}
		// ESC not followed by '\' aborts the string, as in a real terminal,
		// and may open the next sequence.
		if b == '_' {
			g.state = apcBody
			g.buf = g.buf[:0]
			return false
		}
		g.state = apcNormal
	}
	return false
}

// consume scans a whole chunk.
func (g *kittyGraphics) consume(chunk []byte) {
	for _, b := range chunk {
		g.step(b)
	}
}

// takeReplies returns and clears the replies queued since the last call.
func (g *kittyGraphics) takeReplies() [][]byte {
	r := g.replies
	g.replies = nil
	return r
}

// command handles one complete APC body.
func (g *kittyGraphics) command(body []byte) {
	if len(body) == 0 || body[0] != 'G' {
		return // some other APC; not ours
	}
	control, payload, _ := bytes.Cut(body[1:], []byte{';'})
	keys := parseKittyKeys(string(control))

	// A continuation chunk carries only m= (and q=); it belongs to the transfer
	// in flight regardless of what the default action would be.
	if g.pending != nil {
		t := g.pending
		t.add(payload)
		if keys["m"] != "1" {
			g.pending = nil
			g.finish(t)
		}
		return
	}

	switch action(keys) {
	case "q":
		g.reply(keys, "OK")
	case "d":
		g.delete(keys)
	case "p":
		g.place(keys)
	case "t", "T":
		t := &kittyTransfer{keys: keys}
		t.add(payload)
		if keys["m"] == "1" {
			g.pending = t
			return
		}
		g.finish(t)
	}
}

func action(keys map[string]string) string {
	if a, ok := keys["a"]; ok {
		return a
	}
	return "t" // the protocol's default action is transmit
}

// finish decodes a complete transfer, records it, and acknowledges it.
func (g *kittyGraphics) finish(t *kittyTransfer) {
	if t.tooBig {
		g.reply(t.keys, "EFBIG:image payload too large")
		return
	}
	if t.badB64 {
		g.reply(t.keys, "EINVAL:payload is not base64")
		return
	}
	if m := t.keys["t"]; m != "" && m != "d" {
		// A file, temp-file, or shared-memory transfer names data atago cannot
		// read from the transcript.
		g.reply(t.keys, "EBADF:only direct transmission (t=d) is supported")
		return
	}
	img, err := decodeKittyImage(t.keys, t.data)
	if err != nil {
		g.reply(t.keys, "EINVAL:"+err.Error())
		return
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		g.reply(t.keys, "EINVAL:"+err.Error())
		return
	}
	b := img.Bounds()
	rec := kittyImage{
		id:     namedID(t.keys, "i"),
		number: namedID(t.keys, "I"),
		img:    runner.TerminalImage{Width: b.Dx(), Height: b.Dy(), PNG: encoded.Bytes()},
	}
	g.mu.Lock()
	if i := g.indexByID(rec.id); i >= 0 {
		g.images[i] = rec
	} else if len(g.images) < maxGraphicsImages {
		g.images = append(g.images, rec)
	}
	g.mu.Unlock()
	g.reply(t.keys, "OK")
}

// namedID returns the id-like key k (i= or I=). 0 is the protocol's "none".
func namedID(keys map[string]string, k string) string {
	if v := keys[k]; v != "0" {
		return v
	}
	return ""
}

// indexByID returns the index of the image with the given id, or -1. The
// caller holds g.mu.
func (g *kittyGraphics) indexByID(id string) int {
	if id == "" {
		return -1
	}
	return slices.IndexFunc(g.images, func(im kittyImage) bool { return im.id == id })
}

// delete handles a=d: the images the target names leave the screen.
func (g *kittyGraphics) delete(keys map[string]string) {
	target := keys["d"]
	if target == "" {
		target = "a"
	}
	match := deleteMatcher(strings.ToLower(target), keys)
	if match == nil {
		return
	}
	free := target != strings.ToLower(target)
	g.mu.Lock()
	defer g.mu.Unlock()
	if free {
		g.images = slices.DeleteFunc(g.images, match)
		return
	}
	for i := range g.images {
		if match(g.images[i]) {
			g.images[i].hidden = true
		}
	}
}

// deleteMatcher returns which images a lower-cased delete target names, or nil
// for a target a transcript cannot follow.
func deleteMatcher(target string, keys map[string]string) func(kittyImage) bool {
	switch target {
	case "a":
		return func(kittyImage) bool { return true }
	case "i":
		id := namedID(keys, "i")
		return func(im kittyImage) bool { return id != "" && im.id == id }
	case "n":
		number := namedID(keys, "I")
		return func(im kittyImage) bool { return number != "" && im.number == number }
	case "r":
		lo, loErr := strconv.Atoi(keys["x"])
		hi, hiErr := strconv.Atoi(keys["y"])
		if loErr != nil || hiErr != nil {
			return nil
		}
		return func(im kittyImage) bool {
			id, err := strconv.Atoi(im.id)
			return err == nil && id >= lo && id <= hi
		}
	}
	return nil
}

// place handles a=p: an image hidden by a lower-case delete is shown again.
// It sends no reply, as before placements were followed: an image atago could
// not record (a file medium, say) is one a real terminal would place, so an
// ENOENT here would change what the program under test sees.
func (g *kittyGraphics) place(keys map[string]string) {
	g.mu.Lock()
	i := g.indexByID(namedID(keys, "i"))
	if i < 0 {
		if number := namedID(keys, "I"); number != "" {
			i = slices.IndexFunc(g.images, func(im kittyImage) bool { return im.number == number })
		}
	}
	if i >= 0 {
		g.images[i].hidden = false
	}
	g.mu.Unlock()
}

// reply answers a command the way kitty does: only when the program named an
// image id (i=), and not when it asked for quiet (q=1 drops OK replies, q=2 drops
// all of them).
func (g *kittyGraphics) reply(keys map[string]string, msg string) {
	id := keys["i"]
	if id == "" {
		return
	}
	quiet := keys["q"]
	if quiet == "2" || (quiet == "1" && msg == "OK") {
		return
	}
	g.replies = append(g.replies, fmt.Appendf(nil, "\x1b_Gi=%s;%s\x1b\\", id, msg))
}

// snapshot returns the images on screen now, in the order they were sent.
func (g *kittyGraphics) snapshot() []runner.TerminalImage {
	if g == nil {
		return nil
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []runner.TerminalImage
	for _, im := range g.images {
		if !im.hidden {
			out = append(out, im.img)
		}
	}
	return out
}

func parseKittyKeys(control string) map[string]string {
	keys := map[string]string{}
	for _, kv := range strings.Split(control, ",") {
		k, v, ok := strings.Cut(kv, "=")
		if ok && k != "" {
			keys[k] = v
		}
	}
	return keys
}

// decodeKittyImage turns a decoded payload into an image: PNG (f=100), or raw
// RGB (f=24) / RGBA (f=32, the default) of s x v pixels, optionally zlib
// compressed (o=z).
func decodeKittyImage(keys map[string]string, data []byte) (image.Image, error) {
	if keys["o"] == "z" {
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("payload is not zlib: %w", err)
		}
		inflated, err := io.ReadAll(io.LimitReader(r, maxGraphicsPayload+1))
		_ = r.Close()
		if err != nil {
			return nil, fmt.Errorf("payload is not zlib: %w", err)
		}
		if len(inflated) > maxGraphicsPayload {
			return nil, fmt.Errorf("inflated payload is larger than %d bytes", maxGraphicsPayload)
		}
		data = inflated
	}
	format := keys["f"]
	if format == "" {
		format = "32"
	}
	if format == "100" {
		// A PNG header can declare any size; refuse one whose pixels would
		// outgrow the payload bound before png.Decode allocates for it.
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("payload is not a PNG: %w", err)
		}
		if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width*cfg.Height > maxGraphicsPayload/4 {
			return nil, fmt.Errorf("a %dx%d PNG is larger than atago records", cfg.Width, cfg.Height)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("payload is not a PNG: %w", err)
		}
		return img, nil
	}
	channels := map[string]int{"24": 3, "32": 4}[format]
	if channels == 0 {
		return nil, fmt.Errorf("unsupported format f=%s", format)
	}
	w, werr := strconv.Atoi(keys["s"])
	h, herr := strconv.Atoi(keys["v"])
	if werr != nil || herr != nil || w <= 0 || h <= 0 {
		return nil, fmt.Errorf("raw pixel data needs positive s= and v=")
	}
	if w*h*channels != len(data) {
		return nil, fmt.Errorf("%dx%d pixels need %d bytes, got %d", w, h, w*h*channels, len(data))
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < w*h; i++ {
		px := data[i*channels:]
		a := uint8(255)
		if channels == 4 {
			a = px[3]
		}
		img.SetNRGBA(i%w, i/w, color.NRGBA{R: px[0], G: px[1], B: px[2], A: a})
	}
	return img, nil
}

// decodeBase64Lenient accepts padded and unpadded base64, as kitty does.
func decodeBase64Lenient(b []byte) ([]byte, error) {
	s := strings.TrimRight(string(bytes.TrimSpace(b)), "=")
	return base64.RawStdEncoding.DecodeString(s)
}
