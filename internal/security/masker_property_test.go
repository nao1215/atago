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
			for _, v := range []string{tt.a, tt.b} {
				if strings.Contains(got, v) {
					t.Errorf("secret %q survived: %q", v, got)
				}
				// A value declared with one line ending must also be masked
				// where the program under test emitted the other, which is the
				// case a PEM key hits: the crlf-form row declares LF and the
				// text carries CRLF, so checking only the declared spelling
				// would pass while the real bytes leaked.
				crlf := strings.ReplaceAll(v, "\n", "\r\n")
				if crlf != v && strings.Contains(got, crlf) {
					t.Errorf("the CRLF form of secret %q survived: %q", v, got)
				}
			}
		})
	}
}

// The minimum length is a real part of the contract — masking two-character
// values would garble unrelated text — and the property tests above skip
// everything below it, so on their own they would still pass if NewMasker
// stopped masking entirely. This pins both sides of the boundary.
func TestMasker_MinimumLength(t *testing.T) {
	t.Parallel()
	const marker = "|"
	tests := map[string]struct {
		value      string
		wantMasked bool
	}{
		"empty":            {value: "", wantMasked: false},
		"one below":        {value: strings.Repeat("a", minSecretLen-1), wantMasked: false},
		"exactly minimum":  {value: strings.Repeat("b", minSecretLen), wantMasked: true},
		"one above":        {value: strings.Repeat("c", minSecretLen+1), wantMasked: true},
		"long":             {value: strings.Repeat("d", 200), wantMasked: true},
		"multibyte at min": {value: "日本語", wantMasked: true}, // 9 bytes, 3 runes
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tt.value == "" {
				return
			}
			m := NewMasker([]string{tt.value})
			text := marker + tt.value + marker
			masked := !strings.Contains(m.Mask(text), tt.value)
			if masked != tt.wantMasked {
				t.Errorf("value of %d bytes masked = %v, want %v (minSecretLen is %d)",
					len(tt.value), masked, tt.wantMasked, minSecretLen)
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
