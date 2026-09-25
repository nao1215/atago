package yaml

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// MaxDepth bounds how deeply collections may nest.
const MaxDepth = 512

// parser walks the text by byte offset. Every offset has a line and a column,
// which is what nodes and errors carry.
type parser struct {
	src     []byte
	lines   []int // offset of every line start
	depth   int
	anchors map[string]*Node
	// nodes is a block to hand nodes out from. A document is a few hundred of
	// them and they all live as long as the document, so they are taken from
	// blocks rather than allocated one at a time.
	nodes []Node
}

// Parse reads every document of a YAML text.
func Parse(data []byte) (*File, error) {
	data = bytes.TrimPrefix(data, []byte("\xEF\xBB\xBF"))
	if bytes.IndexByte(data, '\r') >= 0 {
		// A carriage return is a line break in YAML, alone or before a line feed.
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	}
	p := &parser{src: data, lines: make([]int, 1, 1+bytes.Count(data, []byte("\n")))}
	for i, b := range data {
		if b == '\n' {
			p.lines = append(p.lines, i+1)
		}
	}
	return p.stream()
}

// node hands out one node holding n.
func (p *parser) node(n Node) *Node {
	if len(p.nodes) == 0 {
		p.nodes = make([]Node, 32)
	}
	out := &p.nodes[0]
	p.nodes = p.nodes[1:]
	*out = n
	return out
}

// newNode hands out a node of kind k starting at off.
func (p *parser) newNode(k Kind, off int) *Node {
	line, col := p.lineCol(off)
	return p.node(Node{Kind: k, Line: line, Col: col})
}

// lineCol returns the 1-based line and column of an offset.
func (p *parser) lineCol(off int) (line, col int) {
	i := sort.Search(len(p.lines), func(i int) bool { return p.lines[i] > off }) - 1
	return i + 1, off - p.lines[i] + 1
}

// col returns the 0-based column of an offset.
func (p *parser) col(off int) int {
	_, c := p.lineCol(off)
	return c - 1
}

func (p *parser) errorf(off int, format string, args ...any) error {
	line, col := p.lineCol(min(off, len(p.src)))
	return &Error{Line: line, Col: col, Msg: fmt.Sprintf(format, args...)}
}

func (p *parser) at(off int) byte {
	if off >= 0 && off < len(p.src) {
		return p.src[off]
	}
	return '\n'
}

func (p *parser) eof(off int) bool { return off >= len(p.src) }

// eol returns the offset of the line break ending the line that holds off, or
// the end of the text.
func (p *parser) eol(off int) int {
	off = min(off, len(p.src))
	if i := bytes.IndexByte(p.src[off:], '\n'); i >= 0 {
		return off + i
	}
	return len(p.src)
}

// nextLine returns the offset of the line after the one holding off.
func (p *parser) nextLine(off int) int { return min(p.eol(off)+1, len(p.src)) }

// skipSpace moves past the spaces and tabs of the current line.
func (p *parser) skipSpace(off int) int {
	for off < len(p.src) && (p.src[off] == ' ' || p.src[off] == '\t') {
		off++
	}
	return off
}

// isBreak reports a line break or the end of the text.
func (p *parser) isBreak(off int) bool { return off >= len(p.src) || p.src[off] == '\n' }

// separates reports a line break, a blank or the end of the text, which is what
// has to follow "---", "...", a ":" that ends a key and a "-" that opens an
// item.
func (p *parser) separates(off int) bool {
	return p.isBreak(off) || p.src[off] == ' ' || p.src[off] == '\t'
}

// lineEnds reports whether the rest of the line from off is blank or a comment.
func (p *parser) lineEnds(off int) bool {
	s := p.skipSpace(off)
	return p.isBreak(s) || p.src[s] == '#'
}

// dash reports a "-" that opens a sequence item.
func (p *parser) dash(off int) bool { return p.at(off) == '-' && p.separates(off+1) }

func (p *parser) hasPrefix(off int, s string) bool {
	return bytes.HasPrefix(p.src[min(off, len(p.src)):], []byte(s))
}

// marker reports a document marker, "---" or "...", at the start of a line.
func (p *parser) marker(off int) bool {
	return p.col(off) == 0 && (p.hasPrefix(off, "---") || p.hasPrefix(off, "...")) && p.separates(off+3)
}

// nextContent returns the offset of the first content character on the first
// line after the one holding off that has any, skipping blank and comment
// lines, or the end of the text.
func (p *parser) nextContent(off int) (int, error) {
	return p.contentFrom(p.nextLine(off))
}

// contentFrom is nextContent starting at the line that begins at off.
func (p *parser) contentFrom(off int) (int, error) {
	for off < len(p.src) {
		c := off
		for c < len(p.src) && p.src[c] == ' ' {
			c++
		}
		switch p.at(c) {
		case '\n', '#':
			off = p.nextLine(c)
			continue
		case '\t':
			t := p.skipSpace(c)
			if p.isBreak(t) || p.src[t] == '#' {
				off = p.nextLine(t)
				continue
			}
			return 0, p.errorf(c, "a tab character cannot indent a line; indent with spaces")
		}
		return c, nil
	}
	return len(p.src), nil
}

// enter counts one collection the reading is inside of.
func (p *parser) enter(off int) error {
	p.depth++
	if p.depth > MaxDepth {
		return p.errorf(off, "the document nests deeper than %d levels", MaxDepth)
	}
	return nil
}

func (p *parser) leave() { p.depth-- }

// stream reads the documents of the text.
func (p *parser) stream() (*File, error) {
	f := &File{src: p.src}
	off, err := p.contentFrom(0)
	if err != nil {
		return nil, err
	}
	for !p.eof(off) {
		var doc *Document
		if doc, off, err = p.document(off); err != nil {
			return nil, err
		}
		if doc != nil {
			f.Docs = append(f.Docs, doc)
		}
	}
	return f, nil
}

// document reads one document starting at off, and returns it with the
// offset of what follows it. A stream position that holds only a "..."
// marker yields no document.
func (p *parser) document(off int) (*Document, int, error) {
	p.anchors = nil
	off, err := p.directives(off)
	if err != nil {
		return nil, 0, err
	}
	explicit := p.col(off) == 0 && p.hasPrefix(off, "---") && p.separates(off+3)
	if explicit {
		if off = p.skipSpace(off + 3); p.lineEnds(off) {
			if off, err = p.nextContent(off); err != nil {
				return nil, 0, err
			}
		}
	}
	if p.eof(off) || p.marker(off) {
		var doc *Document
		if explicit {
			doc = &Document{}
		}
		if p.hasPrefix(off, "...") && p.marker(off) {
			off, err = p.nextContent(off)
		}
		return doc, off, err
	}
	root, end, err := p.block(off, -1, false)
	if err != nil {
		return nil, 0, err
	}
	if off, err = p.nextContent(end); err != nil {
		return nil, 0, err
	}
	switch {
	case p.eof(off):
	case !p.marker(off):
		return nil, 0, p.errorf(off, "unexpected text after the document; check the indentation")
	case p.hasPrefix(off, "..."):
		off, err = p.nextContent(off)
	}
	return &Document{Root: root}, off, err
}

// directives passes over the directive lines a document may open with, which
// have to end at "---". A %YAML directive has to name version 1.
func (p *parser) directives(off int) (int, error) {
	seen := -1
	for !p.eof(off) && p.col(off) == 0 && p.at(off) == '%' {
		if seen < 0 {
			seen = off
		}
		line := string(p.src[off:p.eol(off)])
		if name, rest, _ := strings.Cut(line, " "); name == "%YAML" {
			version := strings.TrimSpace(rest)
			if i := strings.IndexByte(version, '#'); i >= 0 {
				version = strings.TrimSpace(version[:i])
			}
			if major, _, _ := strings.Cut(version, "."); major != "1" {
				return 0, p.errorf(off, "the document is YAML %s; only YAML 1 is read", version)
			}
		}
		var err error
		if off, err = p.nextContent(off); err != nil {
			return 0, err
		}
	}
	if seen >= 0 && (!p.hasPrefix(off, "---") || !p.separates(off+3)) {
		return 0, p.errorf(seen, `a directive has to be followed by "---", which starts the document`)
	}
	return off, nil
}

// props are the anchor and tag written before a node.
type props struct {
	anchor, tag string
	off         int // where the first of them starts
}

func (pr props) empty() bool { return pr.anchor == "" && pr.tag == "" }

// properties reads the anchor and tag that may stand before a node at off, in
// either order, and returns them with the offset of what follows them.
func (p *parser) properties(off int, flow bool) (props, int, error) {
	pr := props{off: off}
	for {
		switch p.at(off) {
		case '&':
			if pr.anchor != "" {
				return pr, 0, p.errorf(off, "a node has two anchors")
			}
			name, end := p.name(off+1, flow)
			if name == "" {
				return pr, 0, p.errorf(off, "an anchor needs a name after \"&\"")
			}
			pr.anchor = name
			off = p.skipSpace(end)
		case '!':
			if pr.tag != "" {
				return pr, 0, p.errorf(off, "a node has two tags")
			}
			end := off + 1
			if p.at(end) == '<' {
				gt := bytes.IndexByte(p.src[end:p.eol(end)], '>')
				if gt < 0 {
					return pr, 0, p.errorf(off, "unterminated verbatim tag")
				}
				end += gt + 1
			} else {
				_, end = p.name(end, flow)
			}
			pr.tag = string(p.src[off:end])
			off = p.skipSpace(end)
		default:
			return pr, off, nil
		}
	}
}

// name reads an anchor, alias or tag name at off: everything up to a blank, a
// line break, or, inside a flow collection, a flow indicator.
func (p *parser) name(off int, flow bool) (string, int) {
	end := off
	for end < len(p.src) {
		c := p.src[end]
		if c == ' ' || c == '\t' || c == '\n' || flow && (c == ',' || c == '[' || c == ']' || c == '{' || c == '}') {
			break
		}
		end++
	}
	return string(p.src[off:end]), end
}

// alias resolves the alias at off to the node it names.
func (p *parser) alias(off int, flow bool) (*Node, int, error) {
	name, end := p.name(off+1, flow)
	if name == "" {
		return nil, 0, p.errorf(off, "an alias needs a name after \"*\"")
	}
	n, ok := p.anchors[name]
	if !ok {
		return nil, 0, p.errorf(off, "alias *%s names no anchor defined before it", name)
	}
	return n, end, nil
}

// withProps applies properties to a node that has just been read, and records
// its anchor.
func (p *parser) withProps(n *Node, pr props) *Node {
	if pr.empty() {
		return n
	}
	// A node with properties starts where they do: that is what an error
	// about its tag, or about the node, points at.
	line, col := p.lineCol(pr.off)
	if n == nil {
		n = p.node(Node{Kind: ScalarNode})
	}
	n.Line, n.Col = line, col
	if pr.tag != "" {
		n.Tag = pr.tag
	}
	if pr.anchor != "" {
		n.Anchor = pr.anchor
		if p.anchors == nil {
			p.anchors = map[string]*Node{}
		}
		p.anchors[pr.anchor] = n
	}
	return n
}

// block reads the node that starts at off, a content character on a line whose
// parent node stands at column parent. With seqAtParent set, a sequence whose
// dashes stand at column parent itself also belongs to the node, which is how a
// mapping value may be written. It returns the node and the offset where it
// ended; what is left of that line is blank or a comment.
func (p *parser) block(off, parent int, seqAtParent bool) (*Node, int, error) {
	pr, rest, err := p.properties(off, false)
	if err != nil {
		return nil, 0, err
	}
	switch {
	case pr.empty():
	case p.lineEnds(rest):
		n, end, err := p.below(rest, parent, seqAtParent)
		if err != nil {
			return nil, 0, err
		}
		return p.withProps(n, pr), max(end, rest), nil
	case p.at(rest) != '?' && p.keyAhead(rest):
		// Properties on the line of an implicit key belong to the key.
		return p.mapping(off, p.col(off))
	default:
		off = rest
	}
	if p.at(off) == '*' {
		if !pr.empty() {
			return nil, 0, p.errorf(pr.off, "an alias cannot have an anchor or a tag")
		}
		n, end, err := p.alias(off, false)
		if err != nil {
			return nil, 0, err
		}
		if p.keyAhead(off) {
			return p.mapping(off, p.col(off))
		}
		if !p.lineEnds(end) {
			return nil, 0, p.errorf(p.skipSpace(end), "unexpected text after the alias")
		}
		return n, end, nil
	}
	c := p.col(off)
	if p.dash(off) {
		n, end, err := p.sequence(off, c)
		if err != nil {
			return nil, 0, err
		}
		return p.withProps(n, pr), end, nil
	}
	if p.keyAhead(off) {
		n, end, err := p.mapping(off, c)
		if err != nil {
			return nil, 0, err
		}
		return p.withProps(n, pr), end, nil
	}
	n, end, err := p.scalar(off, parent)
	if err != nil {
		return nil, 0, err
	}
	return p.withProps(n, pr), end, nil
}

// below reads the node on the lines after the one holding off, whose content
// ended there: a node indented deeper than parent, a sequence at column parent
// when seqAtParent is set, or nothing, which is null.
func (p *parser) below(off, parent int, seqAtParent bool) (*Node, int, error) {
	next, err := p.nextContent(off)
	if err != nil {
		return nil, 0, err
	}
	if p.eof(next) || p.marker(next) {
		return nil, off, nil
	}
	switch c := p.col(next); {
	case c > parent:
		return p.block(next, parent, false)
	case c == parent && seqAtParent && p.dash(next):
		return p.sequence(next, c)
	}
	return nil, off, nil
}

// keyAhead reports whether the text at off opens a block mapping entry: a
// scalar key and then ":" followed by a blank or the end of the line, or "? ".
func (p *parser) keyAhead(off int) bool {
	switch p.at(off) {
	case '"', '\'':
		end, ok := p.quotedEnd(off)
		if !ok {
			return false
		}
		end = p.skipSpace(end)
		return p.at(end) == ':' && p.separates(end+1)
	case '&', '!':
		_, rest, err := p.properties(off, false)
		if err != nil || rest == off || p.lineEnds(rest) || p.at(rest) == '&' || p.at(rest) == '!' {
			// Not a key; the value reader reports what is wrong with it.
			return false
		}
		return p.keyAhead(rest)
	case '[', '{', '|', '>', '#', '%', '@', '`':
		return false
	case '?':
		if p.separates(off + 1) {
			return true
		}
	case '-':
		if p.separates(off + 1) {
			return false
		}
	case '*':
		_, end := p.name(off+1, false)
		end = p.skipSpace(end)
		return p.at(end) == ':' && p.separates(end+1)
	}
	_, ok := p.plainKeyEnd(off)
	return ok
}

// plainKeyEnd returns the offset of the ":" that ends a plain key starting at
// off, when there is one on the line.
func (p *parser) plainKeyEnd(off int) (int, bool) {
	for i := off; i < len(p.src) && p.src[i] != '\n'; i++ {
		switch p.src[i] {
		case ':':
			if p.separates(i + 1) {
				return i, true
			}
		case '#':
			if i > off && (p.src[i-1] == ' ' || p.src[i-1] == '\t') {
				return 0, false
			}
		}
	}
	return 0, false
}

// quotedEnd returns the offset just past a quoted scalar that starts at off and
// ends on the same line.
func (p *parser) quotedEnd(off int) (int, bool) {
	q := p.src[off]
	for i := off + 1; i < len(p.src) && p.src[i] != '\n'; i++ {
		switch {
		case q == '"' && p.src[i] == '\\':
			i++
		case q == '\'' && p.src[i] == '\'' && p.at(i+1) == '\'':
			i++
		case p.src[i] == q:
			return i + 1, true
		}
	}
	return 0, false
}

// mapping reads a block mapping whose keys stand at column c.
func (p *parser) mapping(off, c int) (*Node, int, error) {
	if err := p.enter(off); err != nil {
		return nil, 0, err
	}
	defer p.leave()
	n := p.newNode(MappingNode, off)
	keys := map[string]*Node{}
	var end int
	for {
		key, after, explicit, err := p.key(off)
		if err != nil {
			return nil, 0, err
		}
		var value *Node
		v := p.skipSpace(after)
		switch {
		case p.lineEnds(v):
			value, end, err = p.below(v, c, true)
			end = max(end, v)
		case explicit:
			// The ":" of an explicit key starts its line, so the value after
			// it may open a collection there, as an item after a dash does.
			value, end, err = p.block(v, c, false)
		default:
			value, end, err = p.inline(v, c)
		}
		if err != nil {
			return nil, 0, err
		}
		if err := p.addPair(n, key, value, off, keys); err != nil {
			return nil, 0, err
		}
		if off, err = p.nextContent(end); err != nil {
			return nil, 0, err
		}
		if p.eof(off) || p.col(off) < c || p.marker(off) {
			break
		}
		if p.col(off) > c {
			return nil, 0, p.errorf(off, "unexpected indentation; a mapping value that continues on the next line cannot hold a key")
		}
		if !p.keyAhead(off) {
			if p.dash(off) {
				return nil, 0, p.errorf(off, "a sequence item stands where a key was expected; indent the items under their key or dedent the key")
			}
			return nil, 0, p.errorf(off, "expected a key (\"name: value\"); a mapping value that continues on the next line has to be indented deeper than its key")
		}
	}
	if err := p.merge(n); err != nil {
		return nil, 0, err
	}
	return n, end, nil
}

// addPair appends an entry to a mapping, refusing a key written twice. A
// mapping of many entries keeps its keys in keys, so the check stays linear.
func (p *parser) addPair(n, key, value *Node, off int, keys map[string]*Node) error {
	if key.Kind == ScalarNode && !isMergeKey(key) {
		if first, dup := keys[key.Value]; dup {
			return p.errorf(off, "mapping key %q is written twice (first at line %d)", key.Value, first.Line)
		}
		keys[key.Value] = key
	}
	n.Pairs = append(n.Pairs, Pair{Key: key, Value: value})
	return nil
}

// merge replaces the merge keys (`<<`) of a mapping with the entries of the
// mapping, or mappings, they name. An entry written in the mapping itself wins
// over a merged one, and an earlier merged mapping over a later one.
func (p *parser) merge(n *Node) error {
	has := false
	for _, pair := range n.Pairs {
		if isMergeKey(pair.Key) {
			has = true
			break
		}
	}
	if !has {
		return nil
	}
	own := make([]Pair, 0, len(n.Pairs))
	var sources []*Node
	for _, pair := range n.Pairs {
		if !isMergeKey(pair.Key) {
			own = append(own, pair)
			continue
		}
		switch v := pair.Value; {
		case v != nil && v.Kind == MappingNode:
			sources = append(sources, v)
		case v != nil && v.Kind == SequenceNode:
			for _, item := range v.Items {
				if item == nil || item.Kind != MappingNode {
					return &Error{Line: pair.Key.Line, Col: pair.Key.Col, Msg: "a merge key (<<) takes a mapping or a list of mappings"}
				}
				sources = append(sources, item)
			}
		default:
			return &Error{Line: pair.Key.Line, Col: pair.Key.Col, Msg: "a merge key (<<) takes a mapping or a list of mappings"}
		}
	}
	seen := make(map[string]bool, len(own))
	for _, pair := range own {
		seen[pair.Key.Value] = true
	}
	for _, src := range sources {
		for _, pair := range src.Pairs {
			if seen[pair.Key.Value] {
				continue
			}
			seen[pair.Key.Value] = true
			own = append(own, pair)
		}
	}
	n.Pairs = own
	return nil
}

func isMergeKey(k *Node) bool {
	return k != nil && k.Kind == ScalarNode && k.Style == PlainStyle && k.Tag == "" && k.Value == "<<"
}

// key reads a mapping key at off and returns it with the offset just past its
// ":".
func (p *parser) key(off int) (*Node, int, bool, error) {
	if p.at(off) == '?' && p.separates(off+1) {
		k, after, err := p.explicitKey(off)
		return k, after, true, err
	}
	pr, rest, err := p.properties(off, false)
	if err != nil {
		return nil, 0, false, err
	}
	off = rest
	if p.at(off) == '*' {
		if !pr.empty() {
			return nil, 0, false, p.errorf(pr.off, "an alias cannot have an anchor or a tag")
		}
		k, end, err := p.alias(off, false)
		if err != nil {
			return nil, 0, false, err
		}
		if k.Kind != ScalarNode {
			return nil, 0, false, p.errorf(off, "a mapping key has to be a scalar; *%s is a %s", k.Anchor, k.Kind)
		}
		end = p.skipSpace(end)
		if p.at(end) != ':' || !p.separates(end+1) {
			return nil, 0, false, p.errorf(end, "expected \":\" after the key")
		}
		return k, end + 1, false, nil
	}
	n := p.newNode(ScalarNode, off)
	defer p.withProps(n, pr)
	if q := p.at(off); q == '"' || q == '\'' {
		text, end, err := p.quoted(off)
		if err != nil {
			return nil, 0, false, err
		}
		end = p.skipSpace(end)
		if p.at(end) != ':' || !p.separates(end+1) {
			return nil, 0, false, p.errorf(end, "expected \":\" after the key")
		}
		n.Value = text
		n.Style = quoteStyle(q)
		return n, end + 1, false, nil
	}
	end, ok := p.plainKeyEnd(off)
	if !ok {
		return nil, 0, false, p.errorf(off, "expected a key")
	}
	n.Value = string(bytes.TrimRight(p.src[off:end], " \t"))
	return n, end + 1, false, nil
}

func quoteStyle(q byte) Style {
	if q == '"' {
		return DoubleQuotedStyle
	}
	return SingleQuotedStyle
}

// explicitKey reads a key written after "? ", with its ":" at the start of the
// next line. The key is one scalar: a collection as a key has no field to name.
func (p *parser) explicitKey(off int) (*Node, int, error) {
	c := p.col(off)
	start := p.skipSpace(off + 1)
	n := p.newNode(ScalarNode, start)
	var end int
	switch q := p.at(start); q {
	case '"', '\'':
		text, e, err := p.quoted(start)
		if err != nil {
			return nil, 0, err
		}
		if !p.lineEnds(e) {
			return nil, 0, p.errorf(e, "unexpected text after the quoted key")
		}
		n.Value, n.Style, end = text, quoteStyle(q), e
	case '[', '{', '\n', '#', '&', '*', '!', '|', '>':
		return nil, 0, p.errorf(start, "a key written after \"? \" has to be one plain or quoted scalar on that line")
	default:
		n.Value, end = p.plainLine(start)
	}
	next, err := p.nextContent(end)
	if err != nil {
		return nil, 0, err
	}
	if p.eof(next) || p.col(next) != c || p.at(next) != ':' || !p.separates(next+1) {
		return nil, 0, p.errorf(end, "a key written after \"? \" needs its \":\" at the start of the next line")
	}
	return n, next + 1, nil
}

// sequence reads a block sequence whose dashes stand at column c.
func (p *parser) sequence(off, c int) (*Node, int, error) {
	if err := p.enter(off); err != nil {
		return nil, 0, err
	}
	defer p.leave()
	n := p.newNode(SequenceNode, off)
	var end int
	for {
		var (
			item *Node
			err  error
		)
		after := p.skipSpace(off + 1)
		if p.lineEnds(after) {
			item, end, err = p.below(after, c, false)
			end = max(end, after)
		} else {
			item, end, err = p.block(after, c, false)
		}
		if err != nil {
			return nil, 0, err
		}
		n.Items = append(n.Items, item)
		if off, err = p.nextContent(end); err != nil {
			return nil, 0, err
		}
		if p.eof(off) || p.col(off) < c || p.marker(off) {
			return n, end, nil
		}
		if p.col(off) > c {
			return nil, 0, p.errorf(off, "unexpected indentation; an item that continues on the next line cannot hold a key or an item")
		}
		if !p.dash(off) {
			return n, end, nil
		}
	}
}

// inline reads a value that starts at off on the line of the key it belongs
// to, whose continuation lines have to be indented deeper than parent. It may
// carry an anchor or a tag, and be an alias, a scalar or a flow collection, but
// not a block collection: that starts on the next line.
func (p *parser) inline(off, parent int) (*Node, int, error) {
	pr, rest, err := p.properties(off, false)
	if err != nil {
		return nil, 0, err
	}
	if !pr.empty() && p.lineEnds(rest) {
		n, end, err := p.below(rest, parent, true)
		if err != nil {
			return nil, 0, err
		}
		return p.withProps(n, pr), max(end, rest), nil
	}
	off = rest
	switch {
	case p.at(off) == '*':
		if !pr.empty() {
			return nil, 0, p.errorf(pr.off, "an alias cannot have an anchor or a tag")
		}
		n, end, err := p.alias(off, false)
		if err != nil {
			return nil, 0, err
		}
		if !p.lineEnds(end) {
			return nil, 0, p.errorf(p.skipSpace(end), "unexpected text after the alias")
		}
		return n, end, nil
	case p.dash(off):
		return nil, 0, p.errorf(off, "a sequence cannot start on the line of its key; start the items on the next line")
	}
	if p.at(off) != '?' && p.keyAhead(off) {
		return nil, 0, p.errorf(off, "a mapping value is not allowed here; quote the value if it contains \": \", or start a nested mapping on the next line")
	}
	n, end, err := p.scalar(off, parent)
	if err != nil {
		return nil, 0, err
	}
	return p.withProps(n, pr), end, nil
}

// scalar reads a scalar or a flow collection at off in block context.
func (p *parser) scalar(off, parent int) (*Node, int, error) {
	switch p.at(off) {
	case '|', '>':
		n := p.newNode(ScalarNode, off)
		text, end, err := p.blockScalar(off, parent)
		if err != nil {
			return nil, 0, err
		}
		n.Value = text
		n.Style = LiteralStyle
		if p.at(off) == '>' {
			n.Style = FoldedStyle
		}
		return n, end, nil
	case '[', '{':
		n, end, err := p.flow(off)
		if err != nil {
			return nil, 0, err
		}
		if !p.lineEnds(end) {
			return nil, 0, p.errorf(p.skipSpace(end), "unexpected text after the flow collection")
		}
		return n, end, nil
	case '"', '\'':
		n := p.newNode(ScalarNode, off)
		text, end, err := p.quoted(off)
		if err != nil {
			return nil, 0, err
		}
		if !p.lineEnds(end) {
			return nil, 0, p.errorf(p.skipSpace(end), "unexpected text after the quoted value; a quoted scalar ends at its closing quote")
		}
		n.Value, n.Style = text, quoteStyle(p.at(off))
		return n, end, nil
	case '@', '`':
		return nil, 0, p.errorf(off, "%q cannot start a plain value; quote the value", p.src[off])
	case '?':
		if p.separates(off + 1) {
			return nil, 0, p.errorf(off, "an explicit key (\"? \") cannot stand where a value is")
		}
	}
	n := p.newNode(ScalarNode, off)
	text, end, err := p.plain(off, parent)
	if err != nil {
		return nil, 0, err
	}
	n.Value = text
	return n, end, nil
}

// plainLine returns the text of a plain scalar's line from off to the comment
// or the line break, trimmed, with the offset of that end.
func (p *parser) plainLine(off int) (string, int) {
	end := off
	for end < len(p.src) && p.src[end] != '\n' {
		if p.src[end] == '#' && end > off && (p.src[end-1] == ' ' || p.src[end-1] == '\t') {
			break
		}
		end++
	}
	return string(bytes.TrimRight(p.src[off:end], " \t")), end
}

// hasMappingIndicator reports ": " or a trailing ":" in the text of a plain
// scalar's line, which would make it a key.
func hasMappingIndicator(s string) bool {
	return strings.Contains(s, ": ") || strings.Contains(s, ":\t") || strings.HasSuffix(s, ":")
}

// plain reads a plain scalar in block context. A line indented deeper than
// parent continues it, a line break between two lines folds to a space, and an
// empty line is a line break kept.
func (p *parser) plain(off, parent int) (string, int, error) {
	first, next := p.plainLine(off)
	if hasMappingIndicator(first) {
		return "", 0, p.errorf(off, "a mapping value is not allowed here; quote the value if it contains \": \"")
	}
	if p.at(next) == '#' {
		return first, next, nil
	}
	c, empty, err := p.continuation(next, parent)
	if err != nil {
		return "", 0, err
	}
	if c < 0 {
		return first, next, nil
	}
	var b strings.Builder
	b.WriteString(first)
	for {
		text, end := p.plainLine(c)
		if hasMappingIndicator(text) {
			return "", 0, p.errorf(c, "a mapping value is not allowed here; a line that continues a value cannot hold a key")
		}
		fold(&b, empty)
		b.WriteString(text)
		next = end
		if p.at(next) == '#' {
			return b.String(), next, nil
		}
		if c, empty, err = p.continuation(next, parent); err != nil {
			return "", 0, err
		}
		if c < 0 {
			return b.String(), next, nil
		}
	}
}

// continuation returns the offset of the content on the next line when it
// continues a plain scalar whose parent stands at column parent: not a comment,
// indented deeper than parent, and not a document marker. It is -1 when the
// scalar ends. empty counts the blank lines skipped on the way.
func (p *parser) continuation(off, parent int) (c, empty int, err error) {
	for off = p.nextLine(off); off < len(p.src); off = p.nextLine(off) {
		c = off
		for c < len(p.src) && p.src[c] == ' ' {
			c++
		}
		if p.isBreak(c) {
			empty++
			continue
		}
		if p.at(c) == '\t' {
			t := p.skipSpace(c)
			if p.isBreak(t) {
				empty++
				continue
			}
			c = t
		}
		if p.src[c] == '#' || p.col(c) <= parent || p.marker(c) {
			return -1, 0, nil
		}
		return c, empty, nil
	}
	return -1, 0, nil
}

// fold writes what a line break stands for: a space, or a line break for each
// empty line that followed it.
func fold(b *strings.Builder, empty int) {
	if empty == 0 {
		b.WriteByte(' ')
		return
	}
	for range empty {
		b.WriteByte('\n')
	}
}

// quoted reads a single- or double-quoted scalar, which may run over several
// lines, and returns its text with the offset just past the closing quote.
func (p *parser) quoted(off int) (string, int, error) {
	q := p.src[off]
	var b []byte
	// kept is how much of b a fold may not trim: the blanks an escape wrote are
	// content, not the spacing around a line break.
	kept := 0
	i := off + 1
	for {
		if i >= len(p.src) {
			return "", 0, p.errorf(off, "the quoted value is never closed; add the closing %c", q)
		}
		ch := p.src[i]
		switch {
		case ch == q && q == '\'' && p.at(i+1) == '\'':
			b = append(b, '\'')
			i += 2
		case ch == q:
			return string(b), i + 1, nil
		case ch == '\\' && q == '"':
			var err error
			if b, i, err = p.escape(b, i); err != nil {
				return "", 0, err
			}
			kept = len(b)
		case ch == '\n':
			// The break folds: the blanks around it are not part of the value,
			// and an empty line is a line break kept.
			b = trimBlanks(b, kept)
			var empty int
			i, empty = p.foldBreak(i)
			b = appendFold(b, empty)
		default:
			b = append(b, ch)
			i++
		}
	}
}

// trimBlanks drops the spaces and tabs at the end of b, but not before kept.
func trimBlanks(b []byte, kept int) []byte {
	for len(b) > kept && (b[len(b)-1] == ' ' || b[len(b)-1] == '\t') {
		b = b[:len(b)-1]
	}
	return b
}

// foldBreak passes over the line break at off and the empty lines after it,
// and returns the offset of the next content with the number of empty lines.
func (p *parser) foldBreak(off int) (int, int) {
	empty := 0
	j := off + 1
	for {
		k := p.skipSpace(j)
		if p.isBreak(k) && !p.eof(k) {
			empty++
			j = k + 1
			continue
		}
		return k, empty
	}
}

// appendFold appends what a line break stands for: a space, or a line break
// for each empty line that followed it.
func appendFold(b []byte, empty int) []byte {
	if empty == 0 {
		return append(b, ' ')
	}
	for range empty {
		b = append(b, '\n')
	}
	return b
}

var simpleEscapes = [256]string{
	'0': "\x00", 'a': "\a", 'b': "\b", 't': "\t", '\t': "\t", 'n': "\n", 'v': "\v",
	'f': "\f", 'r': "\r", 'e': "\x1b", ' ': " ", '"': "\"", '/': "/", '\\': "\\",
	'N': "\u0085", '_': " ", 'L': " ", 'P': " ",
}

// escape appends what the escape at off stands for and returns the offset past
// it.
func (p *parser) escape(b []byte, off int) ([]byte, int, error) {
	if off+1 >= len(p.src) {
		return nil, 0, p.errorf(off, "unterminated escape")
	}
	c := p.src[off+1]
	if s := simpleEscapes[c]; s != "" {
		return append(b, s...), off + 2, nil
	}
	digits := 0
	switch c {
	case 'x':
		digits = 2
	case 'u':
		digits = 4
	case 'U':
		digits = 8
	}
	if digits > 0 {
		hex := off + 2
		if hex+digits > len(p.src) {
			return nil, 0, p.errorf(off, "unterminated escape")
		}
		v, err := strconv.ParseInt(string(p.src[hex:hex+digits]), 16, 32)
		r := rune(v)
		if err != nil || !utf8.ValidRune(r) {
			return nil, 0, p.errorf(off, "invalid escape %q", p.src[off:hex+digits])
		}
		return utf8.AppendRune(b, r), hex + digits, nil
	}
	if c == '\n' {
		// An escaped line break joins the lines with nothing between them.
		return b, p.skipSpace(off + 2), nil
	}
	return nil, 0, p.errorf(off, "unknown escape \\%c", c)
}

// blockScalar reads a literal (|) or folded (>) scalar whose indicator stands
// at off and whose parent node stands at column parent.
func (p *parser) blockScalar(off, parent int) (string, int, error) {
	literal := p.src[off] == '|'
	chomp, explicit, i := p.blockHeader(off + 1)
	if !p.lineEnds(i) {
		return "", 0, p.errorf(i, "unexpected text after the block scalar indicator")
	}
	indent := -1
	if explicit > 0 {
		indent = max(parent, 0) + explicit
	}
	lines, end := p.blockLines(p.eol(i), parent, indent)
	trailing := 0
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
		trailing++
	}
	var b strings.Builder
	if literal {
		for k, l := range lines {
			if k > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(l)
		}
	} else {
		foldLines(&b, lines)
	}
	// The line break a scalar ends with is the one after its last line, which a
	// text that stops right there does not have.
	if len(lines) > 0 && chomp != '-' && (trailing > 0 || !p.eof(end)) {
		b.WriteByte('\n')
	}
	if chomp == '+' {
		for range trailing {
			b.WriteByte('\n')
		}
	}
	return b.String(), max(end, i), nil
}

// blockHeader reads the chomping sign and indentation digit that may follow a
// block scalar indicator, in either order.
func (p *parser) blockHeader(i int) (chomp byte, explicit, next int) {
	for ; i < len(p.src); i++ {
		switch c := p.src[i]; {
		case (c == '-' || c == '+') && chomp == 0:
			chomp = c
		case c >= '1' && c <= '9' && explicit == 0:
			explicit = int(c - '0')
		default:
			return chomp, explicit, i
		}
	}
	return chomp, explicit, i
}

// blockLines collects the lines of a block scalar after the line ending at
// next, stripped of their indentation: the given one, or the first content
// line's. It returns them with the offset of the end of the last line taken.
func (p *parser) blockLines(next, parent, indent int) ([]string, int) {
	var lines []string
	for {
		start := p.nextLine(next)
		if start >= len(p.src) || start == next {
			return lines, next
		}
		end := p.eol(start)
		text := string(p.src[start:end])
		spaces := len(text) - len(strings.TrimLeft(text, " "))
		if spaces == len(text) {
			// A blank line belongs to the scalar whatever its width, and keeps
			// what it has beyond the indentation.
			if indent >= 0 && spaces > indent {
				lines = append(lines, text[indent:])
			} else {
				lines = append(lines, "")
			}
			next = end
			continue
		}
		if indent < 0 {
			if spaces <= parent {
				return lines, next
			}
			indent = spaces
		}
		if spaces < indent || (spaces == 0 && p.marker(start)) {
			return lines, next
		}
		lines = append(lines, text[indent:])
		next = end
	}
}

// foldLines joins the lines of a folded scalar: a break between two lines of
// text is a space, an empty line is a break kept, and a line indented deeper
// than the others keeps the breaks around it.
func foldLines(b *strings.Builder, lines []string) {
	prevText := false
	empty := 0
	for _, l := range lines {
		if l == "" {
			empty++
			continue
		}
		more := l[0] == ' ' || l[0] == '\t'
		switch {
		case b.Len() == 0 && !prevText:
			for range empty {
				b.WriteByte('\n')
			}
		case prevText && !more:
			fold(b, empty)
		default:
			b.WriteByte('\n')
			for range empty {
				b.WriteByte('\n')
			}
		}
		b.WriteString(l)
		prevText = !more
		empty = 0
	}
}

// flow reads a flow sequence or mapping starting at off.
func (p *parser) flow(off int) (*Node, int, error) {
	if err := p.enter(off); err != nil {
		return nil, 0, err
	}
	defer p.leave()
	if p.src[off] == '[' {
		return p.flowSequence(off)
	}
	return p.flowMapping(off)
}

// flowSequence reads "[a, b]". An entry written "k: v" is a mapping of one
// entry.
func (p *parser) flowSequence(off int) (*Node, int, error) {
	n := p.newNode(SequenceNode, off)
	i := off + 1
	for {
		var err error
		if i, err = p.flowSpace(i); err != nil {
			return nil, 0, err
		}
		if p.at(i) == ']' {
			return n, i + 1, nil
		}
		start := i
		item, next, err := p.flowValue(i)
		if err != nil {
			return nil, 0, err
		}
		if i, err = p.flowSpace(next); err != nil {
			return nil, 0, err
		}
		if p.at(i) == ':' {
			value, after, err := p.flowPairValue(i, ']')
			if err != nil {
				return nil, 0, err
			}
			m := p.newNode(MappingNode, start)
			m.Pairs = []Pair{{Key: item, Value: value}}
			item, i = m, after
		}
		n.Items = append(n.Items, item)
		switch p.at(i) {
		case ',':
			i++
		case ']':
			return n, i + 1, nil
		default:
			return nil, 0, p.errorf(i, "expected \",\" or \"]\" in the flow sequence")
		}
	}
}

// flowMapping reads "{a: 1, b: 2}". A key with no ":" holds null.
func (p *parser) flowMapping(off int) (*Node, int, error) {
	n := p.newNode(MappingNode, off)
	keys := map[string]*Node{}
	i := off + 1
	for {
		var err error
		if i, err = p.flowSpace(i); err != nil {
			return nil, 0, err
		}
		if p.at(i) == '}' {
			if err := p.merge(n); err != nil {
				return nil, 0, err
			}
			return n, i + 1, nil
		}
		keyOff := i
		key, next, err := p.flowValue(i)
		if err != nil {
			return nil, 0, err
		}
		if key == nil || key.Kind != ScalarNode {
			return nil, 0, p.errorf(keyOff, "a mapping key has to be a scalar")
		}
		if next, err = p.flowSpace(next); err != nil {
			return nil, 0, err
		}
		var value *Node
		if p.at(next) == ':' {
			if value, next, err = p.flowPairValue(next, '}'); err != nil {
				return nil, 0, err
			}
		}
		if err := p.addPair(n, key, value, keyOff, keys); err != nil {
			return nil, 0, err
		}
		switch p.at(next) {
		case ',':
			i = next + 1
		case '}':
			if err := p.merge(n); err != nil {
				return nil, 0, err
			}
			return n, next + 1, nil
		default:
			return nil, 0, p.errorf(next, "expected \",\" or \"}\" in the flow mapping")
		}
	}
}

// flowPairValue reads what follows the ":" at off inside a flow collection
// closed by closer: nothing, which is null, or a value. It returns the offset
// of the "," or closer after it.
func (p *parser) flowPairValue(off int, closer byte) (*Node, int, error) {
	i, err := p.flowSpace(off + 1)
	if err != nil {
		return nil, 0, err
	}
	if p.at(i) == ',' || p.at(i) == closer {
		return nil, i, nil
	}
	value, next, err := p.flowValue(i)
	if err != nil {
		return nil, 0, err
	}
	if next, err = p.flowSpace(next); err != nil {
		return nil, 0, err
	}
	return value, next, nil
}

// flowSpace moves past spacing inside a flow collection: blanks, line breaks
// and comments.
func (p *parser) flowSpace(off int) (int, error) {
	for {
		off = p.skipSpace(off)
		if p.eof(off) {
			return 0, p.errorf(off, "the flow collection is never closed")
		}
		switch p.src[off] {
		case '\n':
			off++
		case '#':
			off = p.eol(off)
		default:
			return off, nil
		}
	}
}

// flowValue reads one value inside a flow collection.
func (p *parser) flowValue(off int) (*Node, int, error) {
	pr, rest, err := p.properties(off, true)
	if err != nil {
		return nil, 0, err
	}
	if !pr.empty() {
		if rest, err = p.flowSpace(rest); err != nil {
			return nil, 0, err
		}
		switch p.at(rest) {
		case ',', ']', '}', ':':
			return p.withProps(nil, pr), rest, nil
		}
	}
	off = rest
	switch p.src[off] {
	case '*':
		if !pr.empty() {
			return nil, 0, p.errorf(pr.off, "an alias cannot have an anchor or a tag")
		}
		return p.alias(off, true)
	case '[', '{':
		n, end, err := p.flow(off)
		if err != nil {
			return nil, 0, err
		}
		return p.withProps(n, pr), end, nil
	case '|', '>':
		return nil, 0, p.errorf(off, "a block scalar cannot be written inside a flow collection")
	}
	n, end, err := p.flowScalar(off)
	if err != nil {
		return nil, 0, err
	}
	return p.withProps(n, pr), end, nil
}

// flowScalar reads a plain or quoted scalar inside a flow collection, where
// "," "[" "]" "{" "}" and ": " end a plain one.
func (p *parser) flowScalar(off int) (*Node, int, error) {
	n := p.newNode(ScalarNode, off)
	if q := p.src[off]; q == '"' || q == '\'' {
		text, next, err := p.quoted(off)
		if err != nil {
			return nil, 0, err
		}
		n.Value, n.Style = text, quoteStyle(q)
		return n, next, nil
	}
	switch p.src[off] {
	case '@', '`':
		return nil, 0, p.errorf(off, "%q cannot start a plain value; quote the value", p.src[off])
	case ',', ']', '}':
		return nil, 0, p.errorf(off, "expected a value before %q", p.src[off])
	case '?':
		if p.separates(off + 1) {
			return nil, 0, p.errorf(off, "an explicit key (\"? \") inside a flow collection is not supported")
		}
	case '-':
		if p.separates(off + 1) {
			return nil, 0, p.errorf(off, "a block sequence item (\"- \") cannot stand inside a flow collection")
		}
	}
	var b []byte
	i := off
	lineStart := off
	for !p.eof(i) {
		c := p.src[i]
		if c == ',' || c == '[' || c == ']' || c == '{' || c == '}' {
			break
		}
		if c == ':' {
			if nx := p.at(i + 1); p.separates(i+1) || nx == ',' || nx == '[' || nx == ']' || nx == '{' || nx == '}' {
				break
			}
		}
		if c == '#' && i > lineStart && (p.src[i-1] == ' ' || p.src[i-1] == '\t') {
			break
		}
		if c == '\n' {
			b = trimBlanks(b, 0)
			j, empty := p.foldBreak(i)
			if p.at(j) == '#' {
				// A comment line inside the collection: skip to its end, and
				// fold the break after it as well.
				i = p.eol(j)
				continue
			}
			b = appendFold(b, empty)
			i, lineStart = j, j
			continue
		}
		b = append(b, c)
		i++
	}
	if p.eof(i) {
		return nil, 0, p.errorf(off, "the flow collection is never closed")
	}
	n.Value = string(trimBlanks(b, 0))
	return n, i, nil
}
