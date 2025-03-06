package erruser

import "github.com/go-modulus/modulus/errors"

// New creates a new user error with the given code and hint
func New(code string, hint string) error {
	return errors.WithAddedTags(errors.WithHint(errors.New(code), hint), errors.UserErrorTag)
}

// NewWithCause creates a new user error with the given code, hint and cause
func NewWithCause(code, hint string, cause error) error {
	return errors.WithCause(New(code, hint), cause)
}

// WithCause adds a cause to the error
// Also marks the error as a user error, even if it was not before
func WithCause(err error, cause error) error {
	return errors.WithAddedTags(errors.WithCause(err, cause), errors.UserErrorTag)
}

func NewValidationError(validationErrors ...error) error {
	mainErr := New("invalid input", "Invalid input provided")
	mainErr = errors.Join(validationErrors...)
	meta := make(map[string]string, len(validationErrors))
	for _, err := range validationErrors {
		meta[err.Error()] = errors.Hint(err)
	}

	return mainErr
}
