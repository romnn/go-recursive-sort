package recursivesort

import (
	"testing"
)

func TestEqualJSON(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		a, b  string
		equal bool
	}{
		{`{"test": ["a", "c", "b"]}`, `{"test": ["c", "a", "b"]}`, true},
		{`{"test": ["a", "c", "b"]}`, `{"testOther": ["c", "a", "b"]}`, false},
		{`{"test": ["a", "c", "b"]}`, `{"test": ["c", "A", "b"]}`, false},
		{`{"test": ["a", "c", "b"]}`, `{"test": ["A", "c", "b"]}`, false},
		{`{"test": ["a", "c", "b"]}`, `{"test": ["a", "c", "b", "d"]}`, false},
		{`{"test": ["a", "c", "b"], "other": "test"}`, `{"test": ["a", "c", "b"]}`, false},
	} {
		equal, err := EqualJSON(c.a, c.b)
		if c.equal != equal {
			t.Errorf("expected EqualJSON(%s, %s) = %v but got %v: %v", c.a, c.a, c.equal, equal, err)
		}
	}
}
