package errlog

import "errors"

type withLoggable struct {
	loggable bool
	err      error
}

func (m withLoggable) IsLoggable() bool {
	return m.loggable
}

func (m withLoggable) Error() string {
	return m.err.Error()
}

func (m withLoggable) Unwrap() error {
	return m.err
}

func IsLoggable(err error) bool {
	type withLoggable interface {
		IsLoggable() bool
	}
	var we withLoggable
	if errors.As(err, &we) {
		return we.IsLoggable()
	}
	return false
}

func WrapLoggable(err error, isLoggable bool) error {
	if err == nil {
		return err
	}

	return withLoggable{loggable: isLoggable, err: err}
}
