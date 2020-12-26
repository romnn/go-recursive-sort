package gorecursivesort

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	parallel = true
)

func recSortWithOrder(order ...interface{}) *RecursiveSort {
	priorityLookup := TypePriorityLookup{}.FromValues(order...)
	return &RecursiveSort{
		TypePriorityLookupHelper: priorityLookup,
	}
}

func TestAssumptions(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	for _, c := range []struct {
		a, b  interface{}
		equal bool
	}{
		{map[string]int{"b": 1, "a": 2}, map[string]int{"a": 2, "b": 1}, true},
	} {
		if diff := cmp.Diff(c.a, c.b); (diff == "") != c.equal {
			t.Errorf("got unexpected result: %s", diff)
		}
	}
}

func TestRecSort(t *testing.T) {
	if parallel {
		t.Parallel()
	}
	defaultSort := &RecursiveSort{}
	for _, c := range []struct {
		sort     *RecursiveSort
		input    interface{}
		expected interface{}
	}{
		{defaultSort, []string{"a", "c", "b"}, []string{"a", "b", "c"}},
		{
			defaultSort,
			struct{ Nested []string }{Nested: []string{"a", "c", "b"}},
			struct{ Nested []string }{Nested: []string{"a", "b", "c"}},
		},
		{defaultSort, []interface{}{"a", "c", "b"}, []interface{}{"a", "b", "c"}},
		{
			recSortWithOrder("", int(0), uint32(0)),
			[]interface{}{"a", 1, uint32(2)},
			[]interface{}{"a", 1, uint32(2)},
		},
		{
			recSortWithOrder(int(0), uint32(0), ""),
			[]interface{}{"a", 1, uint32(2)},
			[]interface{}{1, uint32(2), "a"},
		},
		{
			defaultSort, // int < string < uint32
			[]interface{}{"a", 1, uint32(2)},
			[]interface{}{1, "a", uint32(2)},
		},
		{
			recSortWithOrder("", map[string]int{}),
			[]interface{}{"c", "a", map[string]int{"b": 1, "a": 2}},
			[]interface{}{"a", "c", map[string]int{"a": 2, "b": 1}},
		},
		{
			recSortWithOrder(map[string]int{}, ""),
			[]interface{}{"c", "a", map[string]int{"b": 1, "a": 2}},
			[]interface{}{map[string]int{"a": 2, "b": 1}, "a", "c"},
		},
		{
			defaultSort, // does not change
			[]interface{}{map[string]int{"a": 2}, map[string]int{"a": 1}},
			[]interface{}{map[string]int{"a": 2}, map[string]int{"a": 1}},
		},
		{
			&RecursiveSort{MapSortKey: "a"},
			[]interface{}{map[string]int{"a": 2}, map[string]int{"a": 1}},
			[]interface{}{map[string]int{"a": 1}, map[string]int{"a": 2}},
		},
	} {
		t.Logf("before: %v", c.input)
		c.sort.Sort(c.input)
		t.Logf("after: %v", c.input)
		if diff := cmp.Diff(c.input, c.expected); diff != "" {
			t.Errorf("got unexpected result after recursive sort: %s", diff)
		}
	}
}

func TestAreEqualJSON(t *testing.T) {
	if parallel {
		t.Parallel()
	}
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
		equal, err := AreEqualJSON(c.a, c.b)
		if c.equal != equal {
			t.Errorf("Expected AreEqualJSON(%s, %s) = %v but got %v: %v", c.a, c.a, c.equal, equal, err)
		}
	}
}
