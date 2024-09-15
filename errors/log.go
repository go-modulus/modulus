package errors

import (
	"context"
	"github.com/go-modulus/modulus/errlog"
	"log/slog"
)

const LogAsError = "log_as_error"
const LogAsWarn = "log_as_warn"
const LogAsInfo = "log_as_info"
const LogAsDebug = "log_as_debug"
const DontLog = "dont_log"

func LogError(ctx context.Context, err error, logger *slog.Logger) (slog.Level, bool) {
	level, dontLog := getLevel(Tags(err))

	if dontLog {
		return level, false
	}

	logger.Log(
		ctx,
		level,
		err.Error(),
		errlog.Error(err),
	)

	return level, true
}

func getLevel(tags []string) (slog.Level, bool) {
	level := slog.LevelError
	dontLog := false
	for _, tag := range tags {
		switch tag {
		case LogAsError:
			level = slog.LevelError
		case LogAsWarn:
			level = slog.LevelWarn
		case LogAsInfo:
			level = slog.LevelInfo
		case LogAsDebug:
			level = slog.LevelDebug
		case DontLog:
			dontLog = true
			level = slog.LevelDebug
		}
	}
	return level, dontLog
}
