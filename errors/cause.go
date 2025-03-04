package errors

import "errors"

type withCause struct {
	cause error
	err   error
}

func (m withCause) Cause() error {
	return m.cause
}

func (m withCause) Error() string {
	return m.err.Error()
}

func (m withCause) Unwrap() error {
	return m.err
}

func Cause(err error) error {
	if err == nil {
		return err
	}
	type withCause interface {
		Cause() error
	}
	var we withCause
	if errors.As(err, &we) {
		return we.Cause()
	}
	return nil
}

func CauseString(err error) string {
	if err == nil {
		return ""
	}
	cause := Cause(err)
	if cause == nil {
		return ""
	}
	return cause.Error()
}

func WithCause(err error, cause error) error {
	if err == nil {
		return err
	}

	return withCause{cause: cause, err: err}
}

func WithCauseHint(hint string, cause error) error {
	if cause == nil {
		return nil
	}
	err := New(hint)
	return withCause{cause: cause, err: err}
}
