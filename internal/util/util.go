package util

import (
	"log/slog"
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

func FormatRFC3339(rfc3339 string, layout string, lg *slog.Logger) (string, error) {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		lg.Error(err.Error())
		return "", err
	}

	return t.Format(layout), nil
}
