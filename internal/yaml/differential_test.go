package yaml

import (
	"errors"
	"reflect"
	"testing"
	"unicode/utf8"

	yamlv3 "gopkg.in/yaml.v3"
)

// These inputs exercise the YAML forms that specs and YAML assertions can read.
// The comparison checks decoded values, not just whether parsing succeeds.
func TestValueAgainstYAMLv3(t *testing.T) {
	tests := map[string]string{
		"single quoted":      "value: 'it''s text'\n",
		"double quoted":      "value: \"line\\n\\u263a\"\n",
		"literal":            "value: |\n  first\n  second\n",
		"literal strip":      "value: |-\n  first\n  second\n",
		"literal keep":       "value: |+\n  first\n\n",
		"folded":             "value: >\n  first\n  second\n",
		"folded blank":       "value: >\n  first\n\n  second\n",
		"flow sequence":      "value: [first, 'second', \"third\"]\n",
		"flow mapping":       "value: {first: one, second: two}\n",
		"anchor":             "base: &base [one, two]\nvalue: *base\n",
		"merge":              "base: &base {first: one}\nvalue: {<<: *base, second: two}\n",
		"multiline plain":    "value: first\n  second\n",
		"document marker":    "---\nvalue: text\n...\n",
		"carriage linebreak": "value: first\r\nother: second\r\n",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			var want any
			if err := yamlv3.Unmarshal([]byte(source), &want); err != nil {
				t.Fatal(err)
			}
			file, err := Parse([]byte(source))
			if err != nil {
				t.Fatal(err)
			}
			got, err := Value(file.First())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("decoded value differs: got %#v, want %#v", got, want)
			}
		})
	}
}

func TestInvalidUTF8IsRejectedLikeYAMLv3(t *testing.T) {
	for _, data := range [][]byte{
		{'v', ':', ' ', 0xff, '\n'},
		{'v', ':', ' ', '"', 0xff, '"', '\n'},
		{'v', ':', ' ', '|', '\n', ' ', ' ', 0xff, '\n'},
	} {
		var decoded any
		if err := yamlv3.Unmarshal(data, &decoded); err == nil {
			t.Fatalf("reference parser accepted invalid UTF-8 %q", data)
		}
		if _, err := Parse(data); err == nil {
			t.Errorf("Parse accepted invalid UTF-8 %q", data)
		} else {
			var parseErr *Error
			if !errors.As(err, &parseErr) || parseErr.Line == 0 || parseErr.Col == 0 {
				t.Errorf("Parse(%q) error = %v; want a located YAML error", data, err)
			}
		}
	}
}

func TestRawControlBytesAreRejectedLikeYAMLv3(t *testing.T) {
	for _, invalid := range []byte{0, 1, 7, 0x1b, 0x7f} {
		data := []byte{'v', ':', ' ', invalid, '\n'}
		var decoded any
		if err := yamlv3.Unmarshal(data, &decoded); err == nil {
			t.Fatalf("reference parser accepted raw control byte %#x", invalid)
		}
		if _, err := Parse(data); err == nil {
			t.Errorf("Parse accepted raw control byte %#x", invalid)
		}
	}
}

func TestRawC1ControlsAgreeWithYAMLv3(t *testing.T) {
	for r := rune(0x80); r <= 0x9f; r++ {
		data := append([]byte("v: "), utf8.AppendRune(nil, r)...)
		data = append(data, '\n')
		var decoded any
		wantErr := yamlv3.Unmarshal(data, &decoded)
		if _, err := Parse(data); (err == nil) != (wantErr == nil) {
			t.Errorf("U+%04X: Parse error = %v; yaml.v3 disagrees", r, err)
		} else if err != nil {
			var parseErr *Error
			if !errors.As(err, &parseErr) || parseErr.Line != 1 || parseErr.Col != 4 {
				t.Errorf("U+%04X: error = %v; want line 1, column 4", r, err)
			}
		}
	}
}
