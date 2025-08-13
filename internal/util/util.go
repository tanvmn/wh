package util

import (
	"slices"
	"time"
)

const (
	ErrLine   = "<--LOOK to the LEFT"
	DateTTime = "2006-01-02T15:04:05"
)

var ()

type MalformedRequest struct {
	Status int
	Msg    string
}

func (mr *MalformedRequest) Error() string {
	return mr.Msg
}

func (mr *MalformedRequest) HasErr() bool {
	return mr.Msg != ""
}

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

func FormatDateTTime(dateTTime string, layout string) (string, error) {
	t, err := time.Parse(DateTTime, dateTTime)
	if err != nil {
		return "", err
	}

	return t.Format(layout)[:16], nil
}

// func Format

// AnySlice return an slice of any that contains the vs passed in
func AnySlice[T comparable](vs ...T) []any {
	var as []any
	for _, v := range vs {
		var a any = v
		as = append(as, a)
	}

	return as
}
