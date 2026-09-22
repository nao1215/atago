//go:build windows

package ptyrun

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/nao1215/atago/internal/spec"
)

var procReadConsoleInputW = windows.NewLazySystemDLL("kernel32.dll").NewProc("ReadConsoleInputW")

const (
	keyEventType = 0x0001
	vkEscape     = 0x1B
	vkReturn     = 0x0D
	vkF12        = 0x7B
)

// inputRecord is the Win32 INPUT_RECORD: an event type and a 16-byte union.
type inputRecord struct {
	eventType uint16
	_         uint16
	event     [16]byte
}

// keyPress is the part of a KEY_EVENT_RECORD a TUI acts on.
type keyPress struct {
	down  bool
	vk    uint16
	char  uint16
	state uint32
}

func (r *inputRecord) key() keyPress {
	e := r.event[:]
	return keyPress{
		down:  binary.LittleEndian.Uint32(e[0:4]) != 0,
		vk:    binary.LittleEndian.Uint16(e[6:8]),
		char:  binary.LittleEndian.Uint16(e[10:12]),
		state: binary.LittleEndian.Uint32(e[12:16]),
	}
}

// keysChild reads its input the way crossterm does on Windows: raw console
// mode and ReadConsoleInputW, one KEY_EVENT record per key. It turns on
// bracketed paste so atago lets a test paste, and reports every key press
// until F12: Escape and Enter by name, any other key as the UTF-16 code unit
// it carries.
func keysChild() int {
	in, out, err := childConsole()
	if err != nil {
		fmt.Printf("console: %v\r\n", err)
		return 2
	}
	if err := windows.SetConsoleMode(in, 0); err != nil {
		fmt.Printf("console mode: %v\r\n", err)
		return 2
	}
	childWrite(out, "\x1b[?2004hREADY\r\n")
	var got, raw []string
	for {
		var rec inputRecord
		var n uint32
		r, _, callErr := procReadConsoleInputW.Call(uintptr(in), uintptr(unsafe.Pointer(&rec)), 1, uintptr(unsafe.Pointer(&n)))
		if r == 0 {
			fmt.Printf("ReadConsoleInputW: %v\r\n", callErr)
			return 2
		}
		if n == 0 || rec.eventType != keyEventType {
			continue
		}
		k := rec.key()
		raw = append(raw, fmt.Sprintf("%t/%d/%d/%d", k.down, k.vk, k.char, k.state))
		if !k.down {
			continue
		}
		if k.vk == vkF12 {
			break
		}
		switch k.vk {
		case vkEscape:
			got = append(got, "esc")
		case vkReturn:
			got = append(got, "enter")
		default:
			got = append(got, fmt.Sprintf("%04X", k.char))
		}
	}
	childWrite(out, "RAW:"+strings.Join(raw, " ")+"\r\n")
	childWrite(out, "GOT:"+strings.Join(got, " ")+" END\r\n")
	return 0
}

// vtInputChild reads its input the way a program that asks for VT input does:
// ENABLE_VIRTUAL_TERMINAL_INPUT and ReadConsoleW, so keys arrive as the
// characters a terminal sends. It reports every UTF-16 code unit, ESC as
// "E", until the F12 sequence.
func vtInputChild() int {
	in, out, err := childConsole()
	if err != nil {
		fmt.Printf("console: %v\r\n", err)
		return 2
	}
	childWrite(out, "\x1b[?2004hREADY\r\n")
	var units []uint16
	buf := make([]uint16, 256)
	for !strings.HasSuffix(string(utf16Decode(units)), "\x1b[24~") {
		var n uint32
		if err := windows.ReadConsole(in, &buf[0], uint32(len(buf)), &n, nil); err != nil {
			fmt.Printf("ReadConsoleW: %v\r\n", err)
			return 2
		}
		units = append(units, buf[:n]...)
	}
	var got []string
	for _, u := range units {
		if u == 0x1b {
			got = append(got, "E")
			continue
		}
		got = append(got, fmt.Sprintf("%04X", u))
	}
	childWrite(out, "GOT:"+strings.Join(got, " ")+" END\r\n")
	return 0
}

func utf16Decode(u []uint16) string {
	return windows.UTF16ToString(append(append([]uint16(nil), u...), 0))
}

// runInputChild runs a helper under the host a `graphics: kitty` step uses
// (openConsole) or the one in Windows, performs sends after it is ready, ends
// them with F12, and returns what the helper reported after "GOT:".
func runInputChild(t *testing.T, mode string, openConsole bool, sends ...spec.PTYSend) string {
	t.Helper()
	got, problem := inputChildReport(t, mode, openConsole, sends...)
	if problem != "" {
		t.Fatal(problem)
	}
	return got
}

// inputChildReport is runInputChild reporting a failed run instead of failing.
func inputChildReport(t *testing.T, mode string, openConsole bool, sends ...spec.PTYSend) (string, string) {
	t.Helper()
	command, env := graphicsChild(t, mode)
	session := []spec.PTYAction{{Expect: "READY"}}
	for i := range sends {
		session = append(session, spec.PTYAction{Send: &sends[i]})
	}
	session = append(session, spec.PTYAction{Send: &spec.PTYSend{Key: "f12"}}, spec.PTYAction{Expect: "GOT:.* END"})
	p := &spec.PTY{
		Command: command,
		Timeout: "30s",
		Rows:    10,
		Cols:    400,
		Session: session,
	}
	if openConsole {
		p.Graphics = spec.PTYGraphicsKitty
	}
	res, ef, err := Run(context.Background(), p, t.TempDir(), env)
	if err != nil {
		return "", fmt.Sprintf("Run: %v", err)
	}
	t.Logf("transcript %q", res.Stdout)
	if ef != nil {
		return "", fmt.Sprintf("expect failure: %+v (screen %q)", ef, res.Screen)
	}
	screen := string(res.Screen)
	start := strings.Index(screen, "GOT:")
	end := strings.Index(screen, " END")
	if start < 0 || end < start {
		return "", fmt.Sprintf("screen %q has no report", screen)
	}
	return screen[start+len("GOT:") : end], ""
}

func text(s string) spec.PTYSend  { return spec.PTYSend{Text: &s} }
func paste(s string) spec.PTYSend { return spec.PTYSend{Paste: &s} }

// The family emoji is four people joined by ZWJ; each person is outside the
// BMP, so it is a surrogate pair.
const (
	familyEmoji     = "\U0001F468‍\U0001F469‍\U0001F467‍\U0001F466"
	familyEmojiKeys = "D83D DC68 200D D83D DC69 200D D83D DC67 200D D83D DC66"
	thumbsUp        = "\U0001F44D\U0001F3FD"
	thumbsUpKeys    = "D83D DC4D D83C DFFD"
)

// TestRun_Windows_InputReachesAKeyReader is #678: behind the console host a
// `graphics: kitty` step runs under, a lone Esc never arrived and swallowed the
// key after it, and characters outside the BMP were dropped. A program reading
// console input records must get the keys a terminal user would have typed.
func TestRun_Windows_InputReachesAKeyReader(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		sends []spec.PTYSend
		want  string
	}{
		{"Esc and then q are two keys", []spec.PTYSend{{Key: "esc"}, text("q")}, "esc 0071"},
		{"Esc pressed three times", []spec.PTYSend{{Key: "esc", Times: 3}}, "esc esc esc"},
		{"an emoji string", []spec.PTYSend{text(familyEmoji + " tui")}, familyEmojiKeys + " 0020 0074 0075 0069"},
		{"Enter", []spec.PTYSend{text("a"), {Key: "enter"}}, "0061 enter"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := runInputChild(t, "keys", true, c.sends...); got != c.want {
				t.Errorf("keys = %q, want %q", got, c.want)
			}
		})
	}
}

// TestRun_Windows_PasteKeepsAnEmoji checks the other half of #678: a
// bracketed paste behind that host keeps its content, emoji included, for a
// program reading key records and for one reading VT input.
func TestRun_Windows_PasteKeepsAnEmoji(t *testing.T) {
	t.Parallel()
	t.Run("key records", func(t *testing.T) {
		t.Parallel()
		got := runInputChild(t, "keys", true, paste("a"+thumbsUp+"b"))
		t.Logf("keys = %q", got)
		if want := "0061 " + thumbsUpKeys + " 0062"; !strings.Contains(got, want) {
			t.Errorf("keys = %q, want them to contain %q", got, want)
		}
	})
	t.Run("VT input", func(t *testing.T) {
		t.Parallel()
		got := runInputChild(t, "vtinput", true, paste("a"+thumbsUp+"b"))
		if want := "E 005B 0032 0030 0030 007E 0061 " + thumbsUpKeys + " 0062 E 005B 0032 0030 0031 007E"; !strings.HasPrefix(got, want) {
			t.Errorf("units = %q, want them to start with %q", got, want)
		}
	})
}

// TestRun_Windows_InputThroughTheInboxHost records, for comparison, what the
// same sends look like behind the console host in Windows. It asserts nothing
// that host does not already do.
func TestRun_Windows_InputThroughTheInboxHost(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, mode string
		sends      []spec.PTYSend
	}{
		{"esc q", "keys", []spec.PTYSend{{Key: "esc"}, text("q")}},
		{"emoji", "keys", []spec.PTYSend{text(familyEmoji + " tui")}},
		{"paste keys", "keys", []spec.PTYSend{paste("a" + thumbsUp + "b")}},
		{"paste vt", "vtinput", []spec.PTYSend{paste("a" + thumbsUp + "b")}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, problem := inputChildReport(t, c.mode, false, c.sends...)
			t.Logf("in-box %s: %q %s", c.name, got, problem)
		})
	}
}
