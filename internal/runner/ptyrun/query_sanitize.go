package ptyrun

// The live device-query terminal (terminal_queries.go) is fed the program's raw
// output chunk by chunk so it can answer DA1/DSR/CPR probes against the current
// screen state. Feeding those chunks straight to the emulator has the same two
// hazards the screen render defends against — an oversized loop-per-count CSI
// spins the parser, and a malformed CSI can abort the rest of a chunk before a
// later CPR request is seen — plus one the render does not have: a chunk is an
// arbitrary read boundary, so a CSI can be SPLIT across two chunks. Sanitizing
// each chunk on its own cannot bound a split sequence, because the two halves
// reassemble inside the emulator's stateful parser (issue #438).
//
// streamSanitizer closes that by carrying the incomplete trailing escape of one
// chunk into the next, so a sequence is only handed to the emulator once it is
// whole and has been through sanitizeTranscriptMarks (which clamps counts and
// drops malformed sequences).

// maxQueryCarry bounds the incomplete trailing sequence held between chunks. An
// escape that never terminates — a bare ESC[ with a growing parameter run, an
// OSC with no ST — would otherwise buffer without limit. A real sequence is tiny;
// anything past this is adversarial, so it is dropped rather than held or fed on.
const maxQueryCarry = 4096

// streamSanitizer sanitizes a byte stream delivered in arbitrary chunks, holding
// an incomplete trailing escape sequence until the chunk that completes it.
type streamSanitizer struct {
	carry []byte
}

// feed returns the sanitized bytes ready to hand to the emulator for this chunk:
// every complete unit in carry+chunk, sanitized, with any incomplete trailing
// escape kept back in carry for next time. A trailing escape that grows past
// maxQueryCarry is dropped, since no legitimate sequence is that long.
func (s *streamSanitizer) feed(chunk []byte) []byte {
	buf := chunk
	if len(s.carry) > 0 {
		buf = append(s.carry, chunk...)
		s.carry = nil
	}
	split := incompleteEscapeStart(buf)
	ready, _ := sanitizeTranscriptMarks(buf[:split], nil)
	switch {
	case split >= len(buf):
		s.carry = nil
	case len(buf)-split > maxQueryCarry:
		s.carry = nil // drop an over-long, never-terminating escape
	default:
		s.carry = append([]byte(nil), buf[split:]...)
	}
	return ready
}

// incompleteEscapeStart returns the index in b where a trailing escape sequence
// that is not yet complete begins, or len(b) if b ends on a unit boundary. Bytes
// from that index on must be held until the next chunk completes the sequence, so
// it is sanitized whole instead of reassembling inside the emulator.
func incompleteEscapeStart(b []byte) int {
	i := 0
	for i < len(b) {
		if b[i] != 0x1b {
			i++
			continue
		}
		end, ok := escapeEnd(b, i)
		if !ok {
			return i
		}
		i = end
	}
	return len(b)
}

// escapeEnd reports where the escape sequence starting at b[i] (an ESC) ends and
// whether it is complete within b. It is deliberately conservative: when the kind
// of sequence needs a terminator that has not arrived, it reports incomplete, so
// the caller holds the bytes rather than risk splitting a sequence.
func escapeEnd(b []byte, i int) (end int, complete bool) {
	if i+1 >= len(b) {
		return i, false // a lone trailing ESC: wait for what follows
	}
	switch b[i+1] {
	case 0x1b:
		// A second ESC restarts parsing; the first is an inert one-byte unit.
		return i + 1, true
	case '[':
		return csiEnd(b, i+2)
	case ']', 'P', 'X', '^', '_':
		// OSC/DCS/SOS/PM/APC string sequences run until ST (ESC \) or BEL.
		return stringEnd(b, i+2)
	case '(', ')', '*', '+':
		// A charset designator takes exactly one more byte.
		if i+2 >= len(b) {
			return i, false
		}
		return i + 3, true
	default:
		// ESC + a single final byte (ESC M, ESC D, ESC c, ...).
		return i + 2, true
	}
}

// csiEnd finds the end of a CSI sequence whose parameter/intermediate bytes start
// at b[j], i.e. after "ESC [". A final byte in 0x40..0x7e completes it; an ESC
// aborts it (and starts a new sequence at that ESC); running out of bytes leaves
// it incomplete.
func csiEnd(b []byte, j int) (end int, complete bool) {
	for k := j; k < len(b); k++ {
		c := b[k]
		if c == 0x1b {
			return k, true // aborted CSI; the boundary is the new ESC
		}
		if c >= 0x40 && c <= 0x7e {
			return k + 1, true
		}
	}
	return 0, false
}

// stringEnd finds the end of a string sequence (OSC/DCS/…) whose data starts at
// b[j]. BEL or ST (ESC \) terminates it; running out of bytes leaves it
// incomplete.
func stringEnd(b []byte, j int) (end int, complete bool) {
	for k := j; k < len(b); k++ {
		switch b[k] {
		case 0x07:
			return k + 1, true
		case 0x1b:
			if k+1 < len(b) && b[k+1] == '\\' {
				return k + 2, true
			}
			return 0, false // ESC with no following byte yet: wait for it
		}
	}
	return 0, false
}
