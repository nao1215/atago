//go:build windows

package ptyrun

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"time"
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
// bracketed paste so atago lets a test paste, and reports every key press as
// it arrives and all of them again at F12: Escape and Enter by name, any other
// key as the UTF-16 code unit it carries.
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
		token := fmt.Sprintf("%04X", k.char)
		switch k.vk {
		case vkEscape:
			token = "esc"
		case vkReturn:
			token = "enter"
		}
		got = append(got, token)
		childWrite(out, "K:"+token+"\r\n")
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
func runInputChild(t *testing.T, mode string, openConsole bool, actions ...spec.PTYAction) string {
	t.Helper()
	got, problem := inputChildReport(t, mode, openConsole, actions...)
	if problem != "" {
		t.Fatal(problem)
	}
	return got
}

// inputChildReport is runInputChild reporting a failed run instead of failing.
func inputChildReport(t *testing.T, mode string, openConsole bool, actions ...spec.PTYAction) (string, string) {
	t.Helper()
	command, env := graphicsChild(t, mode)
	session := append([]spec.PTYAction{{Expect: "READY"}}, actions...)
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

func key(name string) spec.PTYAction  { return spec.PTYAction{Send: &spec.PTYSend{Key: name}} }
func text(s string) spec.PTYAction    { return spec.PTYAction{Send: &spec.PTYSend{Text: &s}} }
func paste(s string) spec.PTYAction   { return spec.PTYAction{Send: &spec.PTYSend{Paste: &s}} }
func expect(re string) spec.PTYAction { return spec.PTYAction{Expect: re} }

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
		name    string
		actions []spec.PTYAction
		want    string
	}{
		{
			// Each key must arrive before anything else is sent, the way a
			// TUI closes an overlay on Esc and only then gets the next key.
			"Esc and then q are two keys, each delivered on its own",
			[]spec.PTYAction{key("esc"), expect("K:esc"), text("q"), expect("K:0071")},
			"esc 0071",
		},
		{
			"Esc pressed three times",
			[]spec.PTYAction{{Send: &spec.PTYSend{Key: "esc", Times: 3}}},
			"esc esc esc",
		},
		{
			"an emoji string after an Esc",
			[]spec.PTYAction{key("esc"), expect("K:esc"), text(familyEmoji + " tui"), expect("K:0069")},
			"esc " + familyEmojiKeys + " 0020 0074 0075 0069",
		},
		{"Enter", []spec.PTYAction{text("a"), key("enter")}, "0061 enter"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := runInputChild(t, "keys", true, c.actions...); got != c.want {
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
		got := runInputChild(t, "keys", true, key("esc"), expect("K:esc"), paste("rust\n"+familyEmoji+" tui"), expect("K:0069"))
		if want := "esc 0072 0075 0073 0074 000A " + familyEmojiKeys + " 0020 0074 0075 0069"; got != want {
			t.Errorf("keys = %q, want %q", got, want)
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

// TestRun_Windows_KeysAfterATerminalProbe is the path a ratatui-image TUI
// takes: it probes the terminal on stdin, then reads keys. The probe must end
// on its status report, or its reader stays behind and takes the keys.
func TestRun_Windows_KeysAfterATerminalProbe(t *testing.T) {
	t.Parallel()
	got := runInputChild(t, "stdinprobe", true, expect("ANSWER:.*\n"), key("esc"), expect("K:esc"), text("q"), expect("K:0071"))
	if got != "esc 0071" {
		t.Errorf("keys = %q, want %q", got, "esc 0071")
	}
}

// TestRun_Windows_InputThroughTheInboxHost records, for comparison, what the
// same sends look like behind the console host in Windows. It asserts nothing
// that host does not already do.
func TestRun_Windows_InputThroughTheInboxHost(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, mode string
		actions    []spec.PTYAction
	}{
		{"esc q", "keys", []spec.PTYAction{key("esc"), expect("K:esc"), text("q")}},
		{"emoji", "keys", []spec.PTYAction{text(familyEmoji + " tui")}},
		{"paste keys", "keys", []spec.PTYAction{paste("a" + thumbsUp + "b")}},
		{"paste vt", "vtinput", []spec.PTYAction{paste("a" + thumbsUp + "b")}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, problem := inputChildReport(t, c.mode, false, c.actions...)
			t.Logf("in-box %s: %q %s", c.name, got, problem)
		})
	}
}

// stdinProbeChild asks what ratatui-image asks before a TUI starts reading
// keys: the kitty graphics query, DA1, the cell size, and a status report
// (DSR) that marks the end of the answers. It reads the answers the way Rust's
// stdin does on a console, ReadConsoleW with echo, line and processed input
// turned off but no VT input, and then carries on as keysChild.
func stdinProbeChild() int {
	in, out, err := childConsole()
	if err != nil {
		fmt.Printf("console: %v\r\n", err)
		return 2
	}
	if err := windows.SetConsoleMode(in, windows.ENABLE_EXTENDED_FLAGS|windows.ENABLE_INSERT_MODE|windows.ENABLE_QUICK_EDIT_MODE); err != nil {
		fmt.Printf("console mode: %v\r\n", err)
		return 2
	}
	childWrite(out, "\x1b_Gi=31,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\\x1b[c\x1b[16t\x1b[5n")
	done := make(chan string, 1)
	go func() {
		var units []uint16
		buf := make([]uint16, 50)
		for !strings.HasSuffix(utf16Decode(units), "\x1b[0n") {
			var n uint32
			if err := windows.ReadConsole(in, &buf[0], uint32(len(buf)), &n, nil); err != nil {
				break
			}
			units = append(units, buf[:n]...)
		}
		done <- utf16Decode(units)
	}()
	select {
	case got := <-done:
		childWrite(out, "ANSWER:"+strings.ReplaceAll(got, "\x1b", "E")+"\r\n")
	case <-time.After(5 * time.Second):
		childWrite(out, "ANSWER:TIMEOUT\r\n")
		return 0
	}
	return keysChild()
}
