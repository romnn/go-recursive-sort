package main

import (
	"fmt"

	gorecursivesort "github.com/romnn/go-recursive-sort"
)

func run() bool {
	equal, err := gorecursivesort.AreEqualJSON(`{"test": ["a", "c", "b"]}`, `{"test": ["c", "a", "b"]}`)
	if err != nil {
		panic(err)
	}
	return equal
}

func main() {
	fmt.Println(run())
}
