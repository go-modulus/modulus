package errors

import (
	"github.com/go-modulus/modulus/errlog"
)

func WrapMeta(err error, kv ...string) error {
	return errlog.Wrap(err, kv...)
}

func Meta(err error) map[string]string {
	return errlog.Meta(err)
}
