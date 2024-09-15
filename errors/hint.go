package errors

import (
	"errors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var ht = message.NewPrinter(language.English)

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

func Hint(t *message.Printer, err error) string {
	if err == nil {
		return ""
	}
	type withHint interface {
		Hint() string
	}
	var we withHint
	if errors.As(err, &we) {
		hint := we.Hint()
		if hint != "" && t != nil {
			return t.Sprintf(hint)
		}
	}
	return ""
}

func WrapHint(err error, hint string) error {
	if err == nil {
		return err
	}
	// it is a hack to mark the error for extracting to the translation file
	_ = ht.Sprintf(hint)

	return withHint{hint: hint, err: err}
}
