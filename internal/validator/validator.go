package validator

import (
	"slices"
	"strings"
	"unicode/utf8"
)

// Validator has a map[string]string to hold the err and its val
type Validator struct {
	Errs map[string]string
}

func (v *Validator) Valid() bool {
	return len(v.Errs) == 0
}

func (v *Validator) AddErr(k, msg string) {
	if v.Errs == nil {
		v.Errs = make(map[string]string)
	}

	if _, exists := v.Errs[k]; !exists {
		v.Errs[k] = msg
	}
}

// Check adds an err to v.Errs if the passed in expression is false 
func (v *Validator) Check(ok bool, k, msg string) {
	if !ok {
		v.AddErr(k, msg)
	}
}

func NotBlank(v string) bool {
	return strings.TrimSpace(v) != ""
}

func MaxChars(v string, n int) bool {
	return utf8.RuneCountInString(v) <= n
}

func Permitted[T comparable](v T, permittedVs ...T) bool {
	return slices.Contains(permittedVs, v)
}
