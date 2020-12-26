package main

import (
	"fmt"

	gorecursivesort "github.com/romnnn/go-recursive-sort"
)

func run() string {
	return gorecursivesort.Shout("This is an example")
}

func main() {
	fmt.Println(run())
}
