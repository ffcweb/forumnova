package validator

import (
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Define a new Validator struct which contains a map of validation error messages
// for our form fields.
type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

// Valid checks if there are no field or non-field errors.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// AddFieldError adds a field error message for a specific key if it
// doesn't already exist.
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// AddNonFieldError adds an error message to the NonFieldErrors slice.
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// NotBlank returns true if the value is not an empty string.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// CheckField adds a field error message if the validation check is not 'ok'.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// MaxChars returns true if the value contains no more than n characters.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// CheckPassword validates that the password contains an uppercase letter,
// a digit, and at least 8 characters.
func CheckPassword(value string) bool {
	var hasUpper bool
	var hasDigit bool
	var minChars bool
	for _, char := range value {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}
	minChars = utf8.RuneCountInString(value) >= 8
	return hasUpper && hasDigit && minChars
}

// ValidateEmail checks if the provided email is valid.
func ValidateEmail(value string) bool {
	_, err := mail.ParseAddress(value)
	return err == nil
}
