package spec

import (
	"strconv"
	"strings"
)

// DefaultScreenColor is the value ParseScreenColor returns for `default` — the
// terminal's own foreground or background, i.e. no color was set. It matches
// vt10x's DefaultFG; its DefaultBG is one higher, and both compare equal to
// `default` so a spec never has to know which one a given cell carries.
const DefaultScreenColor uint32 = 1 << 24

// ansiScreenColors is the ANSI palette by name. The bright- prefixed variants
// are the same eight at indices 8-15, which is how a terminal numbers them.
var ansiScreenColors = map[string]uint32{
	"black": 0, "red": 1, "green": 2, "yellow": 3,
	"blue": 4, "magenta": 5, "cyan": 6, "white": 7,
	"bright-black": 8, "bright-red": 9, "bright-green": 10, "bright-yellow": 11,
	"bright-blue": 12, "bright-magenta": 13, "bright-cyan": 14, "bright-white": 15,
}

// ParseScreenColor resolves a spec color to the emulator's value (#382):
// an ANSI name, a 256-palette index, or `default`.
func ParseScreenColor(value string) (uint32, bool) {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "default" {
		return DefaultScreenColor, true
	}
	if c, ok := ansiScreenColors[v]; ok {
		return c, true
	}
	if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 255 {
		return uint32(n), true
	}
	return 0, false
}

// ValidScreenColor reports whether value names a color.
func ValidScreenColor(value string) bool {
	_, ok := ParseScreenColor(value)
	return ok
}

// ScreenColorNames spells the vocabulary for error messages, compactly.
func ScreenColorNames() string {
	return "default, black, red, green, yellow, blue, magenta, cyan, white, their bright-* variants, or a 256-palette index 0-255"
}

// DescribeScreenColor renders an emulator color value the way a spec would
// write it, so a failure message can say `fg=red` rather than `fg=1` — and say
// something honest for a value outside the palette atago names.
func DescribeScreenColor(c uint32) string {
	// vt10x uses DefaultFG and DefaultFG+1 for the terminal's own colors.
	if c == DefaultScreenColor || c == DefaultScreenColor+1 {
		return "default"
	}
	for name, v := range ansiScreenColors {
		if v == c && !strings.HasPrefix(name, "bright-") {
			return name
		}
	}
	for name, v := range ansiScreenColors {
		if v == c {
			return name
		}
	}
	if c <= 255 {
		return strconv.FormatUint(uint64(c), 10)
	}
	// Anything else is a color atago has no name for — a truecolor cell, say.
	// Saying so beats printing a number a spec could not have written.
	return "an unnamed color (" + strconv.FormatUint(uint64(c), 10) + ")"
}
