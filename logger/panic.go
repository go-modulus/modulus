package logger

import (
	"fmt"
	"github.com/go-modulus/modulus/errors/errlog"
	"log/slog"
)

func Recover(logger *slog.Logger) {
	if err := recover(); err != nil {
		logger.Error(
			"panic occurred",
			errlog.Error(fmt.Errorf("%v", err)),
		)
	}
}
