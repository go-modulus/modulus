package errlog

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"log/slog"
)

const LogAsError = "log_as_error"
const LogAsWarn = "log_as_warn"
const LogAsInfo = "log_as_info"
const LogAsDebug = "log_as_debug"
const DontLog = "dont_log"

func LogError(
	ctx context.Context,
	err error,
	logger *slog.Logger,
	defaultLogLevel slog.Level,
) (slog.Level, bool) {
	if err == nil {
		return defaultLogLevel, false
	}
	if defaultLogLevel < slog.LevelDebug {
		return defaultLogLevel, false
	}
	level, dontLog := getLevel(errors.Tags(err), defaultLogLevel)

	if dontLog {
		return level, false
	}

	logger.Log(
		ctx,
		level,
		err.Error(),
		Error(err),
	)

	return level, true
}

func getLevel(tags []string, defaultLogLevel slog.Level) (slog.Level, bool) {
	level := defaultLogLevel
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

func WithLoggingAsError(err error) error {
	return errors.WithAddedTags(err, LogAsError)
}

func WithLoggingAsWarn(err error) error {
	return errors.WithAddedTags(err, LogAsWarn)
}

func WithLoggingAsInfo(err error) error {
	return errors.WithAddedTags(err, LogAsInfo)
}

func WithLoggingAsDebug(err error) error {
	return errors.WithAddedTags(err, LogAsDebug)
}

func WithoutLogging(err error) error {
	return errors.WithAddedTags(err, DontLog)
}
