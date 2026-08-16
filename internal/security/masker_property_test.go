package security

import (
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
)

// The masker's whole point: whatever a secret's value is, it must not survive
// into text the user sees. Generated rather than enumerated, because the
// interesting inputs are the ones nobody thinks to write down — a value that is
// a regex metacharacter, one that overlaps another, one that is a prefix of
// another. Values shorter than minSecretLen are deliberately not masked, so the
// property only holds above that bound.
func TestMasker_RemovesEverySecretValue(t *testing.T) {
	cfg := &quick.Config{Rand: rand.New(rand.NewSource(1)), MaxCount: 3000}

	property := func(secret, before, after string) bool {
		if len(secret) < minSecretLen {
			return true
		}
		m := NewMasker([]string{secret})
		return !strings.Contains(m.Mask(before+secret+after), secret)
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Errorf("a secret survived masking: %v", err)
	}
}

// Two secrets that interact: one containing the other, one equal to the other,
// values made of regex metacharacters. Masking one must not leave the other
// readable.
func TestMasker_OverlappingSecrets(t *testing.T) {
	tests := map[string]struct{ a, b, text string }{
		"prefix":       {a: "abcd", b: "abcdefgh", text: "value abcdefgh here"},
		"suffix":       {a: "efgh", b: "abcdefgh", text: "value abcdefgh here"},
		"identical":    {a: "same-value", b: "same-value", text: "value same-value here"},
		"regex-meta":   {a: `a.c.e`, b: `a+c+e`, text: "value a.c.e and a+c+e here"},
		"backslash":    {a: `a\b\c`, b: `c\d\e`, text: `value a\b\c and c\d\e here`},
		"dollar":       {a: `$1$2$3`, b: `${xyz}`, text: `value $1$2$3 and ${xyz} here`},
		"newline":      {a: "aaa\nbbb", b: "cccc", text: "value aaa\nbbb here cccc"},
		"whole-text":   {a: "everything", b: "xxxx", text: "everything"},
		"adjacent":     {a: "aaaa", b: "bbbb", text: "aaaabbbb"},
		"interleaved":  {a: "xxxx", b: "yyyy", text: "xxxxyyyyxxxx"},
		"unicode":      {a: "日本語です", b: "🎌🎌", text: "秘密 日本語です と 🎌🎌 です"},
		"case-differs": {a: "SecretValue", b: "secretvalue", text: "SecretValue and secretvalue"},
		"crlf-form":    {a: "line1\nline2", b: "zzzz", text: "got line1\r\nline2 and zzzz"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewMasker([]string{tt.a, tt.b})
			got := m.Mask(tt.text)
			if strings.Contains(got, tt.a) {
				t.Errorf("secret A %q survived: %q", tt.a, got)
			}
			if strings.Contains(got, tt.b) {
				t.Errorf("secret B %q survived: %q", tt.b, got)
			}
		})
	}
}

// MaskBytes must agree with Mask: a consumer handed bytes must not see more
// than one handed a string.
func TestMasker_MaskBytesAgreesWithMask(t *testing.T) {
	cfg := &quick.Config{Rand: rand.New(rand.NewSource(2)), MaxCount: 2000}
	property := func(secret, text string) bool {
		if len(secret) < minSecretLen {
			return true
		}
		m := NewMasker([]string{secret})
		full := text + secret + text
		return string(m.MaskBytes([]byte(full))) == m.Mask(full)
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Errorf("MaskBytes and Mask disagree: %v", err)
	}
}

// Masking is idempotent: masking already-masked text changes nothing, so a
// value that passes through two layers of reporting is not double-mangled.
func TestMasker_IsIdempotent(t *testing.T) {
	cfg := &quick.Config{Rand: rand.New(rand.NewSource(3)), MaxCount: 2000}
	property := func(secret, text string) bool {
		if len(secret) < minSecretLen {
			return true
		}
		m := NewMasker([]string{secret})
		once := m.Mask(text + secret + text)
		return m.Mask(once) == once
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Errorf("masking is not idempotent: %v", err)
	}
}
