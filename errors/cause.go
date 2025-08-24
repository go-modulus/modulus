package errors

import "errors"

func Cause(err error) error {
	if err == nil {
		return err
	}
	var e mError
	if errors.As(err, &e) {
		return e.cause
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
	var e mError
	if errors.As(err, &e) {
		e.cause = cause
		return e
	}
	e = new(err.Error())
	e.cause = cause
	return e
}
