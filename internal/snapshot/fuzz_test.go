package snapshot

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/security"
)

// FuzzNormalize checks the two contracts internal/snapshot documents for
// arbitrary captured output:
//   - Normalize is idempotent: Normalize(Normalize(x)) == Normalize(x). A
//     violation means a snapshot written with --update-snapshots would fail its
//     very next comparison.
//   - a declared secret "never reaches the golden file": with a Secrets masker
//     configured, the normalized output must not contain the secret value, even
//     when a normalization step (CRLF fold, CSI/OSC strip) would otherwise
//     reassemble a secret the raw bytes split apart.
//
// Secrets containing bytes the built-in normalizers can inject ('<', '>', '~',
// '*') are excluded from the leak check to avoid false positives where the
// placeholder text itself happens to spell the secret.
func FuzzNormalize(f *testing.F) {
	f.Add([]byte("hello\r\nworld"), "s3cret")
	f.Add([]byte("\x1b[31mred\x1b[0m /tmp/atago-fz-work/x"), "hunter2")
	f.Add([]byte("id=123e4567-e89b-12d3-a456-426614174000 at 2026-07-08T01:02:03Z"), "pass")
	f.Add([]byte("listen 127.0.0.1:8080"), "top\r\nsecret")
	f.Add([]byte("tok\x1b[0men"), "token")
	f.Add([]byte("\x1b]0;title\x07body\x1b[2J"), "abcd")
	f.Add([]byte("\r\r\n0"), "s3cr")
	f.Fuzz(func(t *testing.T, data []byte, secret string) {
		opt := Options{Workdir: "/tmp/atago-fz-work"}
		m := security.NewMasker([]string{secret})
		if !m.Empty() {
			opt.Secrets = m.MaskBytes
		}

		once := Normalize(append([]byte(nil), data...), opt) // must not panic/hang
		twice := Normalize(append([]byte(nil), once...), opt)

		// A secret value that embeds the placeholder characters ('*' from "***",
		// or '<' '>' '~' the built-ins inject) collides with normalization output
		// and makes masking inherently non-convergent — e.g. secret "***0" over
		// "***000". That is a masker/placeholder corner, not a snapshot bug, so it
		// is excluded from both contract checks below.
		collides := opt.Secrets != nil && strings.ContainsAny(secret, "<>~*")
		if !collides && !bytes.Equal(once, twice) {
			t.Fatalf("Normalize is not idempotent:\n in: %q\nonce: %q\ntwice: %q", data, once, twice)
		}
		if opt.Secrets != nil && !collides && bytes.Contains(once, []byte(secret)) {
			t.Fatalf("secret %q reached the golden: %q", secret, once)
		}
	})
}
