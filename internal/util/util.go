package util

import (
	"slices"
	"time"
)

const (
	ErrLine = "<--LOOK to the LEFT"
)

var ()

// Set returns a slice of unique T values
func Set[T comparable](vs ...T) []T {
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

func FormatRFC3339(rfc3339 string, layout string) (string, error) {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return "", err
	}

	return t.Format(layout), nil
}
