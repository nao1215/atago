package spec

import "strings"

// FoldCRLF collapses Windows CRLF line endings to LF so text comparison treats
// line endings as an OS artifact rather than observable CLI behavior. A lone CR
// (an old-Mac line ending) stays observable — only the CR that precedes an LF is
// dropped.
//
// It lives here because two layers have to agree about it: the engine folds a
// matcher's text before comparing it against a stream, and the loader folds the
// same text before deciding whether two matchers of one assert contradict each
// other. A spec authored in a CRLF editor would otherwise get one answer from
// the loader and a different one from the engine.
func FoldCRLF(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
