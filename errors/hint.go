package errors

import (
	"errors"

	"github.com/vorlif/spreak/localize"
)

func Hint(err error) string {
	if err == nil {
		return ""
	}
	var e mError
	if errors.As(err, &e) {
		return e.hint
	}
	return ""
}

func WithHint(err error, hint localize.Singular) error {
	if err == nil {
		return err
	}

	e := copyErr(err)
	e.hint = hint

	return e
}
