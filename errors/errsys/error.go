package errsys

import "github.com/go-modulus/modulus/errors"

// New creates a new system error with the given code and hint
// it is an alias for errors.NewSysError
func New(code string, hint string) error {
	return errors.WithHint(errors.New(code), hint)
}

func NewWithCause(code, hint string, cause error) error {
	return errors.WithCause(New(code, hint), cause)
}

func WithCause(err error, cause error) error {
	return errors.WithAddedTags(errors.WithCause(err, cause), errors.SystemErrorTag)
}
