package errors

import (
	syserrors "errors"
	"golang.org/x/text/message"
)

const (
	SystemErrorTag = "system-error"
	UserErrorTag   = "user-error"
)

// New creates a new handled system error with the given error code.
// Error hint equals to the error code.
// This error is tagged with SystemErrorTag.
//
// If the default error pipeline is used, this error will be logged and shown to the user as is with added request id to the message.
func New(code string) error {
	return WithAddedTags(WithHint(syserrors.New(code), code))
}

// NewSysError creates a new handled system error with the given error code and hint.
// It works as New, but allows to specify a custom hint.
func NewSysError(code string, hint string) error {
	return WithAddedTags(WithHint(syserrors.New(code), hint), SystemErrorTag)
}

// NewUserError creates a new user error with the given error code and message.
// This error is tagged with UserErrorTag.
//
// If the default error pipeline is used, this error will be shown to the user as is without added request id to the message.
// By default the error is not logged.
func NewUserError(code string, hint string) error {
	return WithAddedTags(WithHint(syserrors.New(code), hint), UserErrorTag)
}

func Message(t *message.Printer, err error) string {
	hint := Hint(err)
	if hint != "" {
		return t.Sprint(hint)
	}

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

	return t.Sprintf("Something went wrong on our side")
}

func Is(err, target error) bool {
	return syserrors.Is(err, target)
}

func As(err error, target any) bool {
	return syserrors.As(err, &target)
}

func IsSystemError(err error) bool {
	tags := Tags(err)
	isSystem := false
	for _, tag := range tags {
		switch tag {
		case SystemErrorTag:
			isSystem = true
			break
		case UserErrorTag:
			isSystem = false
		}
	}

	return isSystem
}
