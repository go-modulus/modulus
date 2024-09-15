package errors

import (
	syserrors "errors"
)
import (
	"golang.org/x/text/message"
)

func New(err string) error {
	return WrapHint(syserrors.New(err), err)
}

func Message(t *message.Printer, err error) string {
	type withMessage interface {
		Message() string
	}
	var wm withMessage
	if syserrors.As(err, &wm) {
		return wm.Message()
	}

	type withMessagePrinter interface {
		Message(*message.Printer) string
	}
	var wmp withMessagePrinter
	if syserrors.As(err, &wmp) {
		return wmp.Message(t)
	}

	hint := Hint(t, err)
	if hint != "" {
		return hint
	}

	return t.Sprintf("Something went wrong on our side")
}

func Is(err, target error) bool {
	return syserrors.Is(err, target)
}

func As(err error, target any) bool {
	return syserrors.As(err, &target)
}
