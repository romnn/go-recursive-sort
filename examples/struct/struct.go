package main

import (
	"log"
	"reflect"

	recursivesort "github.com/romnn/go-recursive-sort"
)

// Common fields must be exported to be sorted
type Common struct {
	Values          map[string][]string
	willNotBeSorted []string
}

// Struct fields must be exported to be sorted
type Struct struct {
	A      string
	B      string
	C      []string
	Common Common
}

func main() {
	a := Struct{
		A: "a",
		B: "b",
		C: []string{"a", "b", "c"},
		Common: Common{
			Values: map[string][]string{
				"a": []string{"a", "b", "c"},
				"b": []string{"a", "b", "c"},
			},
			willNotBeSorted: []string{"a", "b"},
		},
	}

	b := Struct{
		A: "a",
		B: "b",
		C: []string{"c", "b", "a"},
		Common: Common{
			Values: map[string][]string{
				"b": []string{"c", "b", "a"},
				"a": []string{"c", "b", "a"},
			},
			willNotBeSorted: []string{"a", "b"},
		},
	}

	sort := recursivesort.RecursiveSort{}
	sort.Sort(&a)
	sort.Sort(&b)

	if equal := reflect.DeepEqual(a, b); !equal {
		log.Fatalf("expected %v == %v", a, b)
	}
}
