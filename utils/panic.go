package utils

import (
	"fmt"

	"go.uber.org/zap"
)

func RecoverPanic(logger *zap.Logger) {
	if err := recover(); err != nil {
		logger.Error(
			"panic occurred",
			zap.Error(fmt.Errorf("panic: %v", err)),
		)
	}
}
