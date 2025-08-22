package erruser

import (
	"github.com/go-modulus/modulus/errors"
	"github.com/vorlif/spreak/localize"
)

// New creates a new user error with the given code and hint
func New(code string, hint localize.Singular) error {
	return errors.WithAddedTags(errors.WithHint(errors.New(code), hint), errors.UserErrorTag)
}

// NewWithCause creates a new user error with the given code, hint and cause
func NewWithCause(code, hint localize.Singular, cause error) error {
	return errors.WithCause(New(code, hint), cause)
}

// WithCause adds a cause to the error
// Also marks the error as a user error, even if it was not before
func WithCause(err error, cause error) error {
	return errors.WithAddedTags(errors.WithCause(err, cause), errors.UserErrorTag)
}

func NewValidationError(validationErrors ...error) error {
	if len(validationErrors) == 0 {
		return nil
	}
	hint := errors.Hint(validationErrors[0])
	mainErr := New("invalid input", hint)
	res := errors.Join(append([]error{mainErr}, validationErrors...)...)
	meta := make(map[string]string, len(validationErrors))
	for _, err := range validationErrors {
		meta[err.Error()] = errors.Hint(err)
	}

	metaList := make([]string, 0, len(meta)*2)
	for k, v := range meta {
		metaList = append(metaList, k, v)
	}

	res = errors.WithAddedMeta(res, metaList...)
	res = errors.WithAddedTags(res, errors.ValidationErrorTag, errors.UserErrorTag)

	return res
}
