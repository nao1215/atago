package ptyrun

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// fourByTwo is a 4x2 RGBA picture: red, green, blue, white on both rows.
func fourByTwo() []byte {
	px := []byte{255, 0, 0, 255, 0, 255, 0, 255, 0, 0, 255, 255, 255, 255, 255, 255}
	return append(append([]byte(nil), px...), px...)
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func pngOf(t *testing.T, raw []byte, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, raw)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func apc(body string) string { return "\x1b_G" + body + "\x1b\\" }

func decodeRecorded(t *testing.T, g *kittyGraphics) []image.Image {
	t.Helper()
	var out []image.Image
	for _, im := range g.snapshot() {
		img, err := png.Decode(bytes.NewReader(im.PNG))
		if err != nil {
			t.Fatalf("recorded image is not a PNG: %v", err)
		}
		if b := img.Bounds(); b.Dx() != im.Width || b.Dy() != im.Height {
			t.Fatalf("recorded size %dx%d disagrees with the PNG %dx%d", im.Width, im.Height, b.Dx(), b.Dy())
		}
		out = append(out, img)
	}
	return out
}

func TestKittyGraphics_DecodesEveryDirectFormat(t *testing.T) {
	t.Parallel()
	raw := fourByTwo()
	rgb := make([]byte, 0, 24)
	for i := 0; i < len(raw); i += 4 {
		rgb = append(rgb, raw[i:i+3]...)
	}
	var z bytes.Buffer
	zw := zlib.NewWriter(&z)
	_, _ = zw.Write(raw)
	_ = zw.Close()

	for name, cmd := range map[string]string{
		"png":        apc("a=T,f=100;" + b64(pngOf(t, raw, 4, 2))),
		"rgba":       apc("a=T,f=32,s=4,v=2;" + b64(raw)),
		"rgba-deflt": apc("a=T,s=4,v=2;" + b64(raw)),
		"rgb":        apc("a=T,f=24,s=4,v=2;" + b64(rgb)),
		"zlib":       apc("a=T,f=32,o=z,s=4,v=2;" + b64(z.Bytes())),
		"transmit":   apc("f=32,s=4,v=2;" + b64(raw)), // a defaults to t
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := newKittyGraphics()
			g.consume([]byte(cmd))
			imgs := decodeRecorded(t, g)
			if len(imgs) != 1 {
				t.Fatalf("recorded %d images, want 1", len(imgs))
			}
			got, ok := color.NRGBAModel.Convert(imgs[0].At(1, 1)).(color.NRGBA)
			if !ok || got != (color.NRGBA{G: 255, A: 255}) {
				t.Fatalf("pixel (1,1) = %v, want green", got)
			}
		})
	}
}

// TestKittyGraphics_ChunkedTransferAcrossReads sends one image in three
// protocol chunks, each padded on its own, delivered a byte at a time: the
// scanner must keep its state across reads and decode each chunk separately.
func TestKittyGraphics_ChunkedTransferAcrossReads(t *testing.T) {
	t.Parallel()
	raw := fourByTwo()
	stream := apc("i=7,a=T,f=32,s=4,v=2,m=1;"+b64(raw[:10])) +
		"plain text between chunks" +
		apc("m=1;"+b64(raw[10:20])) +
		apc("q=2,m=0;"+b64(raw[20:]))
	g := newKittyGraphics()
	for i := range len(stream) {
		g.step(stream[i])
	}
	if imgs := decodeRecorded(t, g); len(imgs) != 1 {
		t.Fatalf("recorded %d images, want 1", len(imgs))
	}
	// The reply is keyed by the first chunk's id and honors the last
	// chunk-independent keys of that first chunk (no q there), so it is sent.
	if got := string(bytes.Join(g.takeReplies(), nil)); got != "\x1b_Gi=7;OK\x1b\\" {
		t.Fatalf("replies = %q", got)
	}
}

func TestKittyGraphics_RepliesRespectQuietAndID(t *testing.T) {
	t.Parallel()
	raw := b64([]byte{1, 2, 3, 4})
	for _, tc := range []struct {
		name, cmd, want string
		recorded        int
	}{
		{"query with id", apc("i=31,a=q,s=1,v=1,f=24;AAAA"), "\x1b_Gi=31;OK\x1b\\", 0},
		{"no id, no reply", apc("a=T,f=32,s=1,v=1;" + raw), "", 1},
		{"q=1 drops OK", apc("i=2,q=1,a=T,f=32,s=1,v=1;" + raw), "", 1},
		{"q=1 keeps errors", apc("i=2,q=1,a=T,f=32,s=2,v=2;" + raw), "\x1b_Gi=2;EINVAL:2x2 pixels need 16 bytes, got 4\x1b\\", 0},
		{"q=2 drops errors", apc("i=2,q=2,a=T,f=32,s=2,v=2;" + raw), "", 0},
		{"file medium", apc("i=3,a=T,t=f,f=100;L3RtcC94"), "\x1b_Gi=3;EBADF:only direct transmission (t=d) is supported\x1b\\", 0},
		{"bad format", apc("i=4,a=T,f=99;" + raw), "\x1b_Gi=4;EINVAL:unsupported format f=99\x1b\\", 0},
		{"not base64", apc("i=5,a=T,f=100;@@@@"), "\x1b_Gi=5;EINVAL:payload is not base64\x1b\\", 0},
		{"not a png", apc("i=6,a=T,f=100;" + raw), "", 0},
		{"huge png header", apc("i=9,a=T,f=100;" + b64(hugePNGHeader())), "\x1b_Gi=9;EINVAL:a 100000x100000 PNG is larger than atago records\x1b\\", 0},
		{"placement ignored", apc("i=8,a=p"), "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := newKittyGraphics()
			g.consume([]byte(tc.cmd))
			got := string(bytes.Join(g.takeReplies(), nil))
			if tc.name == "not a png" {
				if !strings.HasPrefix(got, "\x1b_Gi=6;EINVAL:payload is not a PNG") {
					t.Fatalf("reply = %q", got)
				}
			} else if got != tc.want {
				t.Fatalf("reply = %q, want %q", got, tc.want)
			}
			if n := len(g.snapshot()); n != tc.recorded {
				t.Fatalf("recorded %d images, want %d", n, tc.recorded)
			}
		})
	}
}

// TestKittyGraphics_IgnoresOtherStrings pins that only APC G is read: another
// APC, an OSC, and an APC aborted by a bare ESC leave nothing recorded, and a
// complete command after them still is.
func TestKittyGraphics_IgnoresOtherStrings(t *testing.T) {
	t.Parallel()
	raw := b64([]byte{9, 9, 9, 255})
	stream := "\x1b_Xnot-graphics\x1b\\" +
		"\x1b]0;title\x07" +
		"\x1b_Ga=T,f=32,s=1,v=1;" + raw + "\x1bA" + // aborted: ESC not followed by '\'
		apc("a=T,f=32,s=1,v=1;"+raw)
	g := newKittyGraphics()
	g.consume([]byte(stream))
	if n := len(g.snapshot()); n != 1 {
		t.Fatalf("recorded %d images, want 1", n)
	}
}

// TestKittyGraphics_OversizedCommandIsSkipped bounds memory: a command larger
// than maxGraphicsCommand is dropped whole, and the scanner recovers for the
// next one.
func TestKittyGraphics_OversizedCommandIsSkipped(t *testing.T) {
	t.Parallel()
	g := newKittyGraphics()
	g.consume([]byte("\x1b_Ga=T,f=100;"))
	g.consume(bytes.Repeat([]byte("A"), maxGraphicsCommand+1))
	g.consume([]byte("\x1b\\"))
	if len(g.snapshot()) != 0 || len(g.takeReplies()) != 0 {
		t.Fatal("an oversized command was processed")
	}
	g.consume([]byte(apc("a=T,f=32,s=1,v=1;" + b64([]byte{1, 1, 1, 1}))))
	if len(g.snapshot()) != 1 {
		t.Fatal("the scanner did not recover after an oversized command")
	}
}

// shownWidths lists the widths of the images on screen, in arrival order. The
// deletion tests give every image a distinct width, so the list says which
// ones are left.
func shownWidths(g *kittyGraphics) []int {
	var out []int
	for _, im := range g.snapshot() {
		out = append(out, im.Width)
	}
	return out
}

// transmit is an a=T command for a 1-pixel-high picture of the given width,
// with the extra control keys in front (for example "i=1").
func transmit(keys string, width int) string {
	return apc(keys + ",a=T,f=32,s=" + strconv.Itoa(width) + ",v=1;" + b64(bytes.Repeat([]byte{1, 2, 3, 255}, width)))
}

// TestKittyGraphics_DeletionsTakeImagesOffScreen pins that the recorded set is
// what is on screen now: a delete command takes the images it names away, a
// lower-case target only hides them (the data stays and a=p shows them again),
// an upper-case one frees them, and a transmission that reuses an id replaces
// the image. Targets that name a screen position cannot be followed from the
// transcript and change nothing.
func TestKittyGraphics_DeletionsTakeImagesOffScreen(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		stream string
		want   []int
	}{
		{"nothing deleted", transmit("i=1", 1) + transmit("i=2", 2), []int{1, 2}},
		{"by id hides one", transmit("i=1", 1) + transmit("i=2", 2) + apc("a=d,d=i,i=1"), []int{2}},
		{"by id frees one", transmit("i=1", 1) + transmit("i=2", 2) + apc("a=d,d=I,i=2"), []int{1}},
		{"all", transmit("i=1", 1) + transmit("", 2) + apc("a=d,d=A"), nil},
		{"all is the default target", transmit("i=1", 1) + transmit("", 2) + apc("a=d"), nil},
		{"a hidden image placed again is back in order", transmit("i=1", 1) + transmit("i=2", 2) + apc("a=d,d=a") + apc("a=p,i=1"), []int{1}},
		{"a freed image cannot be placed again", transmit("i=1", 1) + apc("a=d,d=A") + apc("a=p,i=1"), nil},
		{"by number", transmit("I=7", 1) + transmit("I=8", 2) + apc("a=d,d=N,I=7"), []int{2}},
		{"by id range", transmit("i=1", 1) + transmit("i=5", 2) + transmit("i=9", 3) + apc("a=d,d=R,x=2,y=9"), []int{1}},
		{"same id replaces", transmit("i=1", 1) + transmit("i=2", 2) + transmit("i=1", 3), []int{3, 2}},
		{"a position target changes nothing", transmit("i=1", 1) + apc("a=d,d=p,x=1,y=1"), []int{1}},
		{"an unknown id changes nothing", transmit("i=1", 1) + apc("a=d,d=I,i=4"), []int{1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := newKittyGraphics()
			g.consume([]byte(tc.stream))
			if got := shownWidths(g); !slices.Equal(got, tc.want) {
				t.Fatalf("shown widths = %v, want %v", got, tc.want)
			}
		})
	}
}

func FuzzKittyGraphics(f *testing.F) {
	f.Add([]byte(apc("i=1,a=T,f=32,s=1,v=1,m=1;AAAA") + apc("m=0;AAAA")))
	f.Add([]byte(apc("a=T,f=24,o=z,s=100000,v=100000;eJw=")))
	f.Add([]byte("\x1b_G\x1b_G\x1b\\"))
	f.Fuzz(func(t *testing.T, data []byte) {
		g := newKittyGraphics()
		g.consume(data)
		for _, im := range g.snapshot() {
			if im.Width <= 0 || im.Height <= 0 || len(im.PNG) == 0 {
				t.Fatalf("recorded a degenerate image: %+v", im)
			}
		}
	})
}

// hugePNGHeader is a PNG signature and IHDR declaring 100000x100000 pixels and
// nothing else: decoding it naively would allocate 40 GB.
func hugePNGHeader() []byte {
	var buf bytes.Buffer
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	_ = png.Encode(&buf, img)
	b := buf.Bytes()
	// IHDR data starts at byte 16: width then height, big-endian.
	for _, off := range []int{16, 20} {
		b[off], b[off+1], b[off+2], b[off+3] = 0x00, 0x01, 0x86, 0xa0
	}
	binary.BigEndian.PutUint32(b[29:33], crc32.ChecksumIEEE(b[12:29]))
	return b
}
