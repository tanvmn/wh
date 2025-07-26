package main

import "slices"

const errLine = "<==LOOK to the LEFT"

// unique returns a slice of unique T values
func unique[T comparable](vs ...T) []T {
	if len(vs) == 0 {
		return nil
	}

	s := []T{}
	for _, v := range vs {
		if !slices.Contains(s, v) {
			s = append(s, v)
		}
	}

	return s
}