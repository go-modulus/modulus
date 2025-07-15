package errors

import (
	syserrors "errors"
)

const (
	SystemErrorTag     = "system-error"
	UserErrorTag       = "user-error"
	ValidationErrorTag = "validation-error"
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

func Is(err, target error) bool {
	return syserrors.Is(err, target)
}

func As(err error, target any) bool {
	return syserrors.As(err, target)
}

func Join(errs ...error) error {
	return syserrors.Join(errs...)
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
		case UserErrorTag:
			lastTag = UserErrorTag
		}
	}

	return lastTag
}
