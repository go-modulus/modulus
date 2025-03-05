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
// The error hint equals to the error code.
// This error is tagged with SystemErrorTag.
//
// If the default error pipeline is used, this error will be logged and shown to the user as is with added request id to the message.
func New(code string) error {
	return WithAddedTags(WithHint(syserrors.New(code), code), SystemErrorTag)
}

func NewWithCause(code string, cause error) error {
	return WithCause(New(code), cause)
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
	lastTag := getLastErrorTypeTag(err)
	hint := Hint(err)

	return hint == "" || lastTag == SystemErrorTag
}

func IsUserError(err error) bool {
	lastTag := getLastErrorTypeTag(err)

	return lastTag == UserErrorTag
}

func getLastErrorTypeTag(err error) string {
	tags := Tags(err)
	if len(tags) == 0 {
		return ""
	}
	lastTag := ""
	for _, tag := range tags {
		switch tag {
		case SystemErrorTag:
			lastTag = SystemErrorTag
			break
		case UserErrorTag:
			lastTag = UserErrorTag
		}
	}

	return lastTag
}
