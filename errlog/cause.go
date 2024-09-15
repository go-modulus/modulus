package errlog

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

func WrapCause(err error, cause error) error {
	if err == nil {
		return err
	}

	return withCause{cause: cause, err: err}
}
