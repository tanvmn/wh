package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator has a map[string]string to hold the err and its val
type Validator struct {
	FieldErrs    map[string]string
	NonFieldErrs []string
	Errs         string
}

func (v *Validator) Valid() bool {
	return len(v.FieldErrs) == 0 && len(v.NonFieldErrs) == 0 && v.Errs == ""
}

func (v *Validator) AddFieldErr(k, msg string) {
	if v.FieldErrs == nil {
		v.FieldErrs = make(map[string]string)
	}

	if _, exists := v.FieldErrs[k]; !exists {
		v.FieldErrs[k] += msg
	} else {
		v.FieldErrs[k] += "\n" + msg
	}
}

func (v *Validator) AddNonFieldErr(msg string) {
	v.NonFieldErrs = append(v.NonFieldErrs, msg)
}

// Check adds an err to v.Errs if the passed in expression is false
func (v *Validator) Check(ok bool, msg string) {
	if !ok {
		if v.Errs == "" {
			v.Errs += msg
		} else {
			v.Errs += "\n" + msg
		}
	}
}

// CheckField adds an err to v.FieldErrs if the passed in expression is false
func (v *Validator) CheckField(ok bool, k, msg string) {
	if !ok {
		v.AddFieldErr(k, msg)
	}
}

func NotBlank(v string) bool {
	return strings.TrimSpace(v) != ""
}

func MaxChars(v string, n int) bool {
	return utf8.RuneCountInString(v) <= n
}

func MinChars(v string, n int) bool {
	return utf8.RuneCountInString(v) >= n
}

// Permitted return true if v is in vs
func Permitted[T comparable](v T, vs ...T) bool {
	return slices.Contains(vs, v)
}

// Match return true if v matche rx (regular expression)
func Match(v string, rx *regexp.Regexp) bool {
	return rx.MatchString(v)
}

func (v *Validator) Message() string {
	var msg string

	if len(v.FieldErrs) > 0 {
		for k, v := range v.FieldErrs {
			if msg == "" {
				msg += k + ": " + v
			} else {
				msg += "\n" + k + ": " + v
			}
		}
	}

	if len(v.NonFieldErrs) > 0 {
		for _, err := range v.NonFieldErrs {
			if msg == "" {
				msg += err
			} else {
				msg += "\n" + err
			}
		}
	}

	if v.Errs != "" {
		if msg == "" {
			msg += v.Errs
		} else {
			msg += "\n" + v.Errs
		}
	}

	return msg
}
