package errors

import (
	"github.com/go-modulus/modulus/errlog"
)

func WrapCause(err error, cause error) error {
	return errlog.WrapCause(err, cause)
}

func Cause(err error) error {
	return errlog.Cause(err)
}
