package spec

import (
	"fmt"
	"strconv"

	"github.com/nao1215/atago/internal/yaml"
)

// failf is the error of a custom decoder. The reader places an error a decoder
// returns at the node it was decoding, so the loader renders it with the same
// [line:col] and excerpt as any other decode error.
func failf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// eachEntry calls fn with every entry of a mapping in the order written, the
// value read as a plain Go value, so a decoder that rejects a key reports the
// first one the author wrote.
func eachEntry(node *yaml.Node, fn func(key string, value any) error) error {
	for _, pair := range node.Pairs {
		v, err := yaml.Value(pair.Value)
		if err != nil {
			return err
		}
		if err := fn(pair.Key.Value, v); err != nil {
			return err
		}
	}
	return nil
}

// nodeText renders a node for an error: a scalar as the text it holds, quoted,
// and a collection by its kind.
func nodeText(n *yaml.Node) string {
	switch n.Kind {
	case yaml.SequenceNode:
		return "a list"
	case yaml.MappingNode:
		return "a mapping"
	case yaml.ScalarNode:
	}
	return strconv.Quote(n.Value)
}
