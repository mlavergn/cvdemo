package main

import (
	"testing"
)

func TestLoad(t *testing.T) {
	x := NewTrie()
	if x.load("foo") != false {
		t.Errorf("failed to prevent startup with bad file")
	}
}
