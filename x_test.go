package signals_test

import "testing"

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func Equal[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func EqualSlice[T comparable](t *testing.T, expected, actual []T) {
	t.Helper()
	l := len(expected)
	if l != len(actual) {
		t.Fatalf("Expected length to be %d, got %d", l, len(actual))
	}
	for i := range l {
		a := expected[i]
		b := actual[i]
		if a != b {
			t.Fatalf("Expected %v, got %v at %d", a, b, i)
		}
	}
}
