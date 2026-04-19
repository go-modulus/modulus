package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errlog"
)

type ErrorHandler interface {
	HandleError(err error)
}

type NoopErrorHandler struct{}

func NewNoopErrorHandler() ErrorHandler {
	return &NoopErrorHandler{}
}
func (e NoopErrorHandler) HandleError(err error) {}

type ExitErrorHandler struct{}

func NewExitErrorHandler() ErrorHandler {
	return &ExitErrorHandler{}
}
func (e ExitErrorHandler) HandleError(err error) {
	os.Exit(1)
}

type LogErrorHandler struct {
	logger *slog.Logger
}

func NewLogErrorHandler(logger *slog.Logger) ErrorHandler {
	return &LogErrorHandler{}
}
func (e LogErrorHandler) HandleError(err error) {
	e.logger.Error("application run error occurred", errlog.Error(err))
}

type PrintErrorHandler struct{}

func NewPrintErrorHandler() ErrorHandler {
	return &PrintErrorHandler{}
}
func (e PrintErrorHandler) HandleError(err error) {
	hint := errors.Hint(err)

	fmt.Println("Error occurred:", hint)
}
