package main

import (
	"log"

	"github.com/romnn/go-recursive-sort"
)

func main() {
	a := `{"test": ["a", "c", "b"]}`
	b := `{"test": ["c", "a", "b"]}`
	equal, err := recursivesort.EqualJSON(a, b)
	if err != nil {
		log.Fatalf("failed to compare JSON values: %v", err)
	}
	if !equal {
		log.Fatalf("expected %s == %s", a, b)
	}
}
