// Package validator provides functionality for validating and sanitizing data.
package validator

import (
	"encoding/json"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Validator collects field and non-field validation errors.
type Validator struct {
	FieldErrors    map[string][]string `json:"fieldErrors"`
	NonFieldErrors []string            `json:"nonFieldErrors"`
}

// Valid returns true if there are no field or non-field validation errors.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// AddFieldError adds an error message to a specific field.
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string][]string)
	}
	v.FieldErrors[key] = append(v.FieldErrors[key], message)
}

// AddNonFieldError adds a general error message not associated with a specific field.
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// CheckField adds a field error if the provided condition is false.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// JSON returns the validation errors as a JSON byte slice.
func (v *Validator) JSON() []byte {
	b, _ := json.Marshal(v)
	return b
}

// Validatable defines an interface for structs that can validate themselves using a Validator.
type Validatable interface {
	Validate(v *Validator)
}

/////////////////////////
// String Validators
/////////////////////////

// NotBlank returns true if the string is not empty or whitespace.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// Blank returns true if the string is empty or contains only whitespace.
func Blank(value string) bool {
	return strings.TrimSpace(value) == ""
}

// MaxChars returns true if the string contains no more than n characters.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// MinChars returns true if the string contains at least n characters.
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// Matches returns true if the string matches the provided regular expression.
func Matches(value string, pattern *regexp.Regexp) bool {
	return pattern.MatchString(value)
}

/////////////////////////
// Numeric Validators
/////////////////////////

// IsNumber returns true if the string represents a valid integer.
func IsNumber(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

// MinInt returns true if value is greater than or equal to min.
func MinInt(value, min int) bool { return value >= min }

// MaxInt returns true if value is less than or equal to max.
func MaxInt(value, max int) bool { return value <= max }

// MinFloat returns true if value is greater than or equal to min.
func MinFloat(value, min float64) bool { return value >= min }

// MaxFloat returns true if value is less than or equal to max.
func MaxFloat(value, max float64) bool { return value <= max }

/////////////////////////
// Duration Validators
/////////////////////////

// MaxDuration returns true if the duration is less than or equal to maxDuration.
func MaxDuration(d, maxDuration time.Duration) bool { return d <= maxDuration }

// MinDuration returns true if the duration is greater than or equal to minDuration.
func MinDuration(d, minDuration time.Duration) bool { return d >= minDuration }

/////////////////////////
// Generic Helpers
/////////////////////////

// PermittedValue returns true if value is among the provided permittedValues.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}
