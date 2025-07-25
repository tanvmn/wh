package main

import (
	"slices"
	"testing"
)

func testUnique(t *testing.T) {
	want := []int{1, 2, 3}
	got := unique([]int{1, 1, 2, 3}...)

	if !slices.Equal(want, got) {
		t.Errorf("want: %v, got: %v", want, got)
	}
}
