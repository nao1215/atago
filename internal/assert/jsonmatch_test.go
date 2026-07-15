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

func TestValuesEqual_SharedBackingSliceDifferentLengths(t *testing.T) {
	t.Parallel()

	backing := []any{"a", "b"}
	left := backing[:1]
	right := backing[:2]

	if valuesEqual(left, right) {
		t.Fatal("valuesEqual reported shared backing slices with different lengths as equal")
	}
}
