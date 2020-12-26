package main

import (
	"testing"
)

func TestExample(t *testing.T) {
	out := run()
	expected := true
	if out != expected {
		t.Errorf("Got %t but expected %t", out, expected)
	}
}
