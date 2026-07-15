package assert

import "testing"

func TestValuesEqual_SameDeepTree(t *testing.T) {
	t.Parallel()

	var v any = "leaf"
	for i := 0; i < 100000; i++ {
		v = []any{v}
	}

	if !valuesEqual(v, v) {
		t.Fatal("valuesEqual returned false for the same nested tree")
	}
}
