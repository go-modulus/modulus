package errwrap

import (
	"github.com/go-modulus/modulus/errlog"
)

type Wrapper func(err error) error

func With(w ...Wrapper) Wrapper {
	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, fn := range w {
			err = fn(err)
		}

		return err
	}
}

func Wrap(err error, w ...Wrapper) error {
	if err == nil {
		return nil
	}

	for _, fn := range w {
		err = fn(err)
	}

	return err
}

func WrapCause(err error, cause error) error {
	return errlog.WrapLoggable(errlog.WrapCause(err, cause), true)
}
