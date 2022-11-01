package recursivesort

import (
	"testing"

	"reflect"
)

func recSortWithOrder(order ...interface{}) *RecursiveSort {
	priorityLookup := TypePriorityLookup{}.FromValues(order...)
	return &RecursiveSort{
		TypePriorityLookupHelper: priorityLookup,
	}
}

func TestAssumptions(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		a, b  interface{}
		equal bool
	}{
		{map[string]int{"b": 1, "a": 2}, map[string]int{"a": 2, "b": 1}, true},
	} {
		if equal := reflect.DeepEqual(c.a, c.b); equal != c.equal {
			t.Errorf("expected reflect.DeepEqual(%s, %s) == %t", c.a, c.b, c.equal)
		}
	}
}

func TestSortSliceOfUnexportedStruct(t *testing.T) {
	t.Parallel()
	sort := &RecursiveSort{
		StructSortField: "Exported",
	}
	type TestStruct struct {
		unexported string
		// Exported   string
		Exported int64
	}
	value := []interface{}{
		TestStruct{
			unexported: "test",
			Exported:   0,
		},
		TestStruct{
			unexported: "test",
			Exported:   1,
		},
	}
	t.Logf("before: %v", value)
	sort.Sort(&value)
	t.Logf("after: %v", value)
}

func TestRecSort(t *testing.T) {
	t.Parallel()
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
		if equal := reflect.DeepEqual(c.input, c.expected); !equal {
			t.Errorf("expected %v but got %v", c.expected, c.input)
		}
	}
}
