// Package yaml reads and writes the YAML atago works with: spec files,
// directory manifests, and the YAML a program under test prints for an
// assertion to inspect.
//
// It is atago's own, so that how a spec is read is decided here and not by a
// third-party decoder. The reader keeps what a spec needs from the text: the
// line and column of every node, and the text of a plain scalar exactly as it
// was written, so a text field reads `1.20` as the four characters an author
// typed rather than the float they happen to spell.
//
// Input may be the output of an arbitrary program, so no input makes the
// reader panic, nesting is bounded, and an alias cannot expand a small
// document into an unbounded value.
package yaml

import (
	"strconv"
	"strings"
)

// Kind is what a node holds.
type Kind uint8

// The kinds of node.
const (
	ScalarNode Kind = iota + 1
	SequenceNode
	MappingNode
)

// String names the kind the way an error message does.
func (k Kind) String() string {
	switch k {
	case ScalarNode:
		return "scalar"
	case SequenceNode:
		return "sequence"
	case MappingNode:
		return "mapping"
	}
	return "value"
}

// Style is how a scalar was written.
type Style uint8

// The styles of scalar.
const (
	// PlainStyle is a scalar written bare. Only a plain scalar can be null, a
	// bool or a number; any other style is the string it spells.
	PlainStyle Style = iota
	SingleQuotedStyle
	DoubleQuotedStyle
	LiteralStyle
	FoldedStyle
)

// Node is one value of a document.
//
// An alias is not a node of its own: the parser hands out the node the alias
// names, so the same *Node may appear more than once in a document.
type Node struct {
	Kind  Kind
	Style Style
	// Tag is the tag as written ("!!binary", "!Ref"), or "" when there is none.
	Tag string
	// Anchor is the name the node was anchored with, or "".
	Anchor string
	// Value is the text of a scalar: for a plain scalar, the characters as
	// written; for any other style, the string the scalar spells.
	Value string
	// Items are the elements of a sequence.
	Items []*Node
	// Pairs are the entries of a mapping in the order written, with the
	// entries a merge key (`<<`) brings in after them.
	Pairs []Pair
	// Line and Col are where the node starts, both counted from 1.
	Line, Col int
}

// Pair is one entry of a mapping.
type Pair struct {
	Key, Value *Node
}

// nullNode stands for a value that is absent, such as the value of a key with
// nothing after it.
var nullNode = &Node{Kind: ScalarNode, Style: PlainStyle}

// IsNull reports a node that holds null: nothing at all, or a plain scalar
// spelling null, or one tagged !!null.
func (n *Node) IsNull() bool {
	if n == nil {
		return true
	}
	if n.Kind != ScalarNode {
		return false
	}
	if n.Tag == "!!null" {
		return true
	}
	if n.Style != PlainStyle || n.Tag != "" {
		return false
	}
	return isNullWord(n.Value)
}

func isNullWord(s string) bool {
	switch s {
	case "", "~", "null", "Null", "NULL":
		return true
	}
	return false
}

// Get returns the value of the mapping entry whose key is key, or nil when n is
// not a mapping or has no such entry.
func (n *Node) Get(key string) *Node {
	if n == nil || n.Kind != MappingNode {
		return nil
	}
	for _, p := range n.Pairs {
		if p.Key.Value == key {
			return p.Value
		}
	}
	return nil
}

// Index returns the i-th element of a sequence, or nil when n is not a
// sequence or has no such element.
func (n *Node) Index(i int) *Node {
	if n == nil || n.Kind != SequenceNode || i < 0 || i >= len(n.Items) {
		return nil
	}
	return n.Items[i]
}

// Document is one parsed document. Root is nil for a document with no content.
type Document struct {
	Root *Node
}

// File is every document of a YAML text in order.
type File struct {
	Docs []*Document
	src  []byte
}

// First returns the first document that has content, the one a reader of a
// single document takes, or nil when there is none.
func (f *File) First() *Node {
	for _, d := range f.Docs {
		if d.Root != nil {
			return d.Root
		}
	}
	return nil
}

// Source returns the text the file was parsed from, after a leading byte-order
// mark was dropped and CRLF line breaks were read as LF.
func (f *File) Source() []byte { return f.src }

// quoteKey writes a mapping key for an error path, quoting it when it is not a
// plain word.
func quoteKey(k string) string {
	for _, r := range k {
		if !isWordRune(r) {
			return strconv.Quote(k)
		}
	}
	if k == "" {
		return `""`
	}
	return k
}

// joinPath extends an error path with a mapping key.
func joinPath(path, key string) string {
	if path == "" {
		return quoteKey(key)
	}
	return path + "." + quoteKey(key)
}

// indexPath extends an error path with a sequence index.
func indexPath(path string, i int) string {
	var b strings.Builder
	b.WriteString(path)
	b.WriteByte('[')
	b.WriteString(strconv.Itoa(i))
	b.WriteByte(']')
	return b.String()
}

// isWordRune reports a character a key can hold and still be written bare in
// an error path.
func isWordRune(r rune) bool {
	return r == '_' || r == '-' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9'
}
