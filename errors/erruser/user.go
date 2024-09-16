package erruser

import (
	"errors"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ErrorCode string

const (
	InternalErrorCode ErrorCode = "Internal"
)

type userError struct {
	code    ErrorCode
	message string
}

func (e userError) Error() string {
	return fmt.Sprintf("<%s> %s", e.code, e.message)
}

func (e userError) Code() ErrorCode {
	return e.code
}

func (e userError) Message(t *message.Printer) string {
	return t.Sprintf(e.message)
}

var p = message.NewPrinter(language.English)

func New(code ErrorCode, message string) error {
	// it is a hack to mark the error for extracting to the translation file
	_ = p.Sprintf(message)
	return userError{
		code:    code,
		message: message,
	}
}

func Code(err error) ErrorCode {
	type withCode interface {
		Code() ErrorCode
	}
	var wc withCode
	if errors.As(err, &wc) {
		return wc.Code()
	}
	return InternalErrorCode
}

func Message(t *message.Printer, err error) string {
	type withMessage interface {
		Message() string
	}
	var wm withMessage
	if errors.As(err, &wm) {
		return wm.Message()
	}

	type withMessagePrinter interface {
		Message(*message.Printer) string
	}
	var wmp withMessagePrinter
	if errors.As(err, &wmp) {
		return wmp.Message(t)
	}

	return t.Sprintf("Something went wrong on our side")
}

func Details(err error) map[string]any {
	type withDetails interface {
		Details() map[string]any
	}
	var we withDetails
	if errors.As(err, &we) {
		return we.Details()
	}
	return nil
}
