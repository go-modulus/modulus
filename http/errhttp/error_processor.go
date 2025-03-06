package errhttp

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

func (p *ErrorPipeline) Process(ctx context.Context, err error) error {
	for _, processor := range p.Processors {
		err = processor(ctx, err)
	}
	return err
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
			SaveMultiErrorsToMeta(),
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

func SaveMultiErrorsToMeta() ErrorProcessor {
	return func(ctx context.Context, err error) error {
		if err == nil {
			return nil
		}
		allErrors := extractErrors(err)
		if len(allErrors) == 0 {
			return err
		}
		err = allErrors[0]
		additionalErrors := allErrors[1:]

		if len(additionalErrors) > 0 {
			meta := make([]string, 0, len(additionalErrors)*2)
			for _, err2 := range additionalErrors {
				meta = append(meta, err2.Error(), errors.Hint(err2))
			}
			err = errors.WithAddedMeta(err, meta...)
		}

		return err
	}
}

func extractErrors(err error) []error {
	var allErrors []error

	for err != nil {
		if uw, ok := err.(interface{ Unwrap() []error }); ok {
			allErrors = append(allErrors, uw.Unwrap()...)
			break
		}

		if uw, ok := err.(interface{ Unwrap() error }); ok {
			err = uw.Unwrap()
		} else {
			break
		}
	}

	return allErrors
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
