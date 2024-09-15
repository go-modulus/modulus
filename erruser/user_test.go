package erruser_test

import (
	"errors"
	"github.com/go-modulus/modulus/erruser"
	"github.com/go-modulus/modulus/translation"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"testing"
)

var matcher = language.NewMatcher(
	[]language.Tag{
		language.MustParse("en-GB"),
	},
)

func TestCode(t *testing.T) {
	err := erruser.New("InvalidInput", "Invalid input")
	assert.Equal(t, "InvalidInput", string(erruser.Code(err)))
}

func TestCode_StdError(t *testing.T) {
	err := errors.New("invalid input")
	assert.Equal(t, erruser.InternalErrorCode, erruser.Code(err))
}

type customCodeError struct{}

func (customCodeError) Code() erruser.ErrorCode {
	return "CustomCode"
}

func (customCodeError) Error() string {
	return "custom error"
}

func TestCode_CustomError(t *testing.T) {
	err := customCodeError{}
	assert.Equal(t, "CustomCode", string(erruser.Code(err)))
}

func TestMessage(t *testing.T) {
	translator := translation.NewTranslator(matcher)
	p := translator.NewPrinter("en-GB")
	err := erruser.New("InvalidInput", "Invalid input")
	assert.Equal(t, "Invalid input", erruser.Message(p, err))
}

func TestMessage_StdError(t *testing.T) {
	translator := translation.NewTranslator(matcher)
	p := translator.NewPrinter("en-GB")
	err := errors.New("invalid input")
	assert.Equal(t, "Something went wrong on our side", erruser.Message(p, err))
}

type customMessageError struct {
}

func (customMessageError) Message(t *message.Printer) string {
	return t.Sprintf("Custom message error")
}

func (customMessageError) Error() string {
	return "custom message error"
}

func TestMessage_CustomError(t *testing.T) {
	translator := translation.NewTranslator(matcher)
	p := translator.NewPrinter("en-GB")
	err := customMessageError{}
	assert.Equal(t, "Custom message error", erruser.Message(p, err))
}

func TestDetails(t *testing.T) {
	err := erruser.New("InvalidInput", "Invalid input")
	assert.Nil(t, erruser.Details(err))
}

func TestDetails_StdError(t *testing.T) {
	err := errors.New("invalid input")
	assert.Nil(t, erruser.Details(err))
}

type customDetailsError struct {
	postID uuid.UUID
}

func (e customDetailsError) Details() map[string]any {
	return map[string]any{
		"postID": e.postID,
	}
}

func (customDetailsError) Error() string {
	return "custom details error"
}

func TestDetails_CustomError(t *testing.T) {
	postID := uuid.Must(uuid.NewV6())
	err := customDetailsError{postID: postID}
	assert.Equal(t, map[string]any{"postID": postID}, erruser.Details(err))
}
