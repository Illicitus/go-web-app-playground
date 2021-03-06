package utils

import "strings"

type PublicError interface {
	error
	Public() string
}

type coreError string

func (ce coreError) Error() string {
	return string(ce)
}

func (ce coreError) Public() string {
	s := strings.Replace(string(ce), "orm: ", "", 1)
	s = strings.Replace(string(ce), "validator: ", "", 1)

	return s
}

type privateError string

func (e privateError) Error() string {
	return string(e)
}

type ValidationError struct {

	// Common
	InvalidID privateError

	// User
	InvalidEmail     coreError
	EmailRequired    coreError
	InvalidPassword  coreError
	PasswordTooShort coreError
	EmailTaken       coreError
	PasswordRequired coreError
	RememberTooShort privateError
	RememberRequired privateError

	// Gallery
	UserIDRequired privateError
	TitleRequired  coreError
}

func newValidationError() ValidationError {
	return ValidationError{
		InvalidID: "validator: ID must be > 0",

		InvalidEmail:     "validator: Invalid email provided",
		EmailRequired:    "validator: Email address is required",
		InvalidPassword:  "validator: Incorrect password provided",
		PasswordTooShort: "validator: Password must be at least 8 characters",
		EmailTaken:       "validator: Email address already taken",
		PasswordRequired: "validator: Password is required",
		RememberTooShort: "validator: Remember token must be at least 32 bytes",
		RememberRequired: "validator: Remember is required",

		UserIDRequired: "validator: user ID is required",
		TitleRequired:  "validator: title required",
	}
}

var ValErr = newValidationError()

type GormError struct {
	// Common
	InvalidID privateError
	NotFound  privateError

	// User
	InvalidEmail    privateError
	InvalidPassword privateError
}

func newGormError() GormError {
	return GormError{
		InvalidID:       "orm: ID must be > 0",
		InvalidEmail:    "orm: Invalid email provided",
		NotFound:        "orm: Resource not found",
		InvalidPassword: "orm: Incorrect password provided",
	}
}

var GormErr = newGormError()
