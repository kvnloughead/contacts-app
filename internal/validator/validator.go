package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// Email pattern recommended by W3C.
// https://html.spec.whatwg.org/multipage/input.html#valid-e-mail-address
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

var PhoneNumberRX = regexp.MustCompile(`^\+?[0-9]{1,3}\s?\(?[0-9]{1,4}\)?[-.\s]?[0-9]{1,4}[-.\s]?[0-9]{1,9}$`)

type Validator struct {
	// For errors that are associated with specific form fields.
	FieldErrors map[string]string

	// For errors that aren't associated with specific form fields.
	NonFieldErrors []string
}

// Returns true if there are no validation errors.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

/*
Adds an error to the validator's FieldErrors struct, unless the field in question already has an error.

The struct will be initialized if it hasn't been already.
*/
func (v *Validator) AddFieldError(field, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[field]; !exists {
		v.FieldErrors[field] = message
	}
}

// Adds an error message to the NonFieldErrors slice.
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

/*
Adds an error to the validator's FieldErrors struct if the field isn't valid.

'ok' should be true if the field is valid, otherwise false. 'field' is the name of the input field. 'message' is the associated error message.
*/
func (v *Validator) CheckField(ok bool, field, message string) {
	if !ok {
		v.AddFieldError(field, message)
	}
}

// Returns true if the string is not an empty string.
func NotBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}

// Returns true if a value contains no more than n characters.
func MaxChars(s string, n int) bool {
	return utf8.RuneCountInString(s) <= n
}

// Returns true if a value contains at least n characters.
func MinChars(s string, n int) bool {
	return utf8.RuneCountInString(s) >= n
}

// Returns true if the string matches the regex.
func Matches(s string, rx *regexp.Regexp) bool {
	return rx.MatchString(s)
}

// Returns true if the value matches one of the permittedValues.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}
