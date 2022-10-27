## go-recursive-sort

[![Build Status](https://github.com/romnn/go-recursive-sort/workflows/test/badge.svg)](https://github.com/romnn/go-recursive-sort/actions)
[![GitHub](https://img.shields.io/github/license/romnn/go-recursive-sort)](https://github.com/romnn/go-recursive-sort)
[![GoDoc](https://godoc.org/github.com/romnn/go-recursive-sort?status.svg)](https://godoc.org/github.com/romnn/go-recursive-sort)
[![Test Coverage](https://codecov.io/gh/romnn/go-recursive-sort/branch/master/graph/badge.svg)](https://codecov.io/gh/romnn/go-recursive-sort)

Recursively sort any golang `interface{}` for comparisons in your unit tests.

#### Installation

```bash
$ go get github.com/romnn/go-recursive-sort
```

#### Example

```go
// examples/example1/example1.go

package main

import (
	"log"

	recursivesort "github.com/romnn/go-recursive-sort"
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

```

For more examples, see `examples/`.

#### Development

###### Prerequisites

Before you get started, make sure you have installed the following tools

```bash
$ python3 -m pip install pre-commit bump2version invoke
$ go install golang.org/x/tools/cmd/goimports@latest
$ go install golang.org/x/lint/golint@latest
$ go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
```

Please always make sure all code checks pass:

```bash
invoke pre-commit
```
