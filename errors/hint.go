package errors

import (
	"errors"
	"github.com/vorlif/spreak/localize"
)

type withHint struct {
	hint string
	err  error
}

func (m withHint) Hint() string {
	return m.hint
}

func (m withHint) Error() string {
	return m.err.Error()
}

func (m withHint) Unwrap() error {
	return m.err
}

func Hint(err error) string {
	if err == nil {
		return ""
	}
	type withHint interface {
		Hint() string
	}
	var we withHint
	if errors.As(err, &we) {
		return we.Hint()
	}
	return ""
}

func WithHint(err error, hint localize.Singular) error {
	if err == nil {
		return err
	}

	return withHint{hint: hint, err: err}
}
