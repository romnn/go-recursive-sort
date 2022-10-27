package recursivesort

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// EqualJSON checks if two JSON strings are equal, ignoring the order of values.
//
// Strings `a` and `b` are first unmarshaled into a JSON `interface{}`.
// Then, they are recursively sorted and compared using `reflect.DeepEqual`
//
// If values differ or cannot be parsed,
// `false` is returned and `err != nil` describes the error.
// Otherwise, `true` and `nil` is returned.
//
// If you wish to compare for equality in a different way,
// e.g. using github.com/google/go-cmp/cmp, you can easily reimplement
// this method yourself.
func EqualJSON(a, b string) (bool, error) {
	var err error
	var ia, ib interface{}

	sort := RecursiveSort{}

	err = json.Unmarshal([]byte(a), &ia)
	if err != nil {
		return false, fmt.Errorf("unmarshal failed: %v", err)
	}
	sort.Sort(ia)

	err = json.Unmarshal([]byte(b), &ib)
	if err != nil {
		return false, fmt.Errorf("unmarshal failed: %v", err)
	}
	sort.Sort(ib)

	return reflect.DeepEqual(ia, ib), nil
}
