package http

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errlog"
	"github.com/go-modulus/modulus/errors/errsys"
	context2 "github.com/go-modulus/modulus/http/context"
	"log/slog"
)

const InternalErrorCode = "unhandled internal error"

type ErrorProcessor func(ctx context.Context, err error) error

type ErrorPipeline struct {
	Processors []ErrorProcessor
}

type ErrorLoggerConfig struct {
	UserLogLevel   string `env:"HTTP_USER_ERROR_LOG_LEVEL, default=dont_log" comment:"Log level for the user errors: dont_log, error, warn, info, debug"`
	SystemLogLevel string `env:"HTTP_SYSTEM_ERROR_LOG_LEVEL, default=error" comment:"Log level for the system errors: dont_log, error, warn, info, debug"`
}

func NewDefaultErrorPipeline(
	logger *slog.Logger,
	loggerConfig ErrorLoggerConfig,
) *ErrorPipeline {
	return &ErrorPipeline{
		Processors: []ErrorProcessor{
			LogError(logger, loggerConfig),
			HideInternalError(),
			AddRequestID(),
		},
	}
}

func HideInternalError() ErrorProcessor {
	return func(ctx context.Context, err error) error {
		if err == nil {
			return nil
		}
		hint := errors.Hint(err)
		code := err.Error()
		if hint == "" || code == InternalErrorCode {
			resultErr := errsys.New(InternalErrorCode, "Something went wrong")
			resultErr = errors.WithCause(resultErr, err)

			return resultErr
		}
		return err
	}
}

func AddRequestID() ErrorProcessor {
	return func(ctx context.Context, err error) error {
		if err == nil {
			return nil
		}
		requestID := context2.GetRequestID(ctx)
		if requestID != "" {
			err = HideInternalError()(ctx, err)
			if errors.IsSystemError(err) {
				hint := errors.Hint(err)
				hint = hint + " (Code: " + requestID + ")"
				err = errors.WithAddedMeta(err, "requestId", requestID)
				return errors.WithHint(err, hint)
			}
			return errors.WithAddedMeta(err, "requestId", requestID)
		}
		return err
	}
}

func LogError(logger *slog.Logger, loggerConfig ErrorLoggerConfig) ErrorProcessor {
	return func(ctx context.Context, err error) error {
		if err == nil {
			return nil
		}
		defaultLogLevel := convertConfigLogLevelToSlogLevel(loggerConfig.SystemLogLevel)
		if errors.IsUserError(err) {
			defaultLogLevel = convertConfigLogLevelToSlogLevel(loggerConfig.UserLogLevel)
		}
		errlog.LogError(ctx, err, logger, defaultLogLevel)
		return err
	}
}

func convertConfigLogLevelToSlogLevel(logLevel string) slog.Level {
	switch logLevel {
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	case "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	case "dont_log":
		return slog.Level(-8)
	default:
		return slog.LevelDebug
	}
}
