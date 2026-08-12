package errhttp

import (
	"context"
	"log/slog"

	"github.com/go-modulus/modulus/errors/errpipe"
)

const InternalErrorCode = "unhandled internal error"

type ErrorProcessor func(ctx context.Context, err error) error

// ErrorPipelineFactory
type ErrorPipelineFactory interface {
	New() *ErrorPipeline
}

type ErrorPipeline struct {
	pipe *errpipe.ErrorPipeline
}

func (p *ErrorPipeline) Process(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if p.pipe == nil {
		return err
	}
	return p.pipe.Process(ctx, err)
}

// SetProcessor adds an error processor.
func (p *ErrorPipeline) SetProcessor(rank int, processor ErrorProcessor) {
	if p.pipe == nil {
		p.pipe = &errpipe.ErrorPipeline{}
	}
	p.pipe.SetProcessor(
		rank, func(ctx context.Context, err error) error {
			return processor(ctx, err)
		},
	)
}

type ErrorLoggerConfig struct {
	UserLogLevel   string `env:"HTTP_USER_ERROR_LOG_LEVEL, default=dont_log" comment:"Log level for the user errors: dont_log, error, warn, info, debug"`
	SystemLogLevel string `env:"HTTP_SYSTEM_ERROR_LOG_LEVEL, default=error" comment:"Log level for the system errors: dont_log, error, warn, info, debug"`
}

// NewDefaultErrorPipeline creates a new default error pipeline.
func NewDefaultErrorPipeline(
	logger *slog.Logger,
	loggerConfig ErrorLoggerConfig,
) *ErrorPipeline {
	return &ErrorPipeline{
		pipe: errpipe.NewDefaultErrorPipeline(
			logger, errpipe.ErrorLoggerConfig{
				UserLogLevel:   loggerConfig.UserLogLevel,
				SystemLogLevel: loggerConfig.SystemLogLevel,
			},
		),
	}
}

func WrapErrProcessor(processor errpipe.ErrorProcessor) ErrorProcessor {
	return func(ctx context.Context, err error) error {
		return processor(ctx, err)
	}
}

func HideInternalError() ErrorProcessor {
	return WrapErrProcessor(errpipe.HideInternalError())
}

func AddRequestID() ErrorProcessor {
	return WrapErrProcessor(errpipe.AddRequestID())
}

func LogError(logger *slog.Logger, loggerConfig ErrorLoggerConfig) ErrorProcessor {
	return WrapErrProcessor(
		errpipe.LogError(
			logger, errpipe.ErrorLoggerConfig{
				UserLogLevel:   loggerConfig.UserLogLevel,
				SystemLogLevel: loggerConfig.SystemLogLevel,
			},
		),
	)
}

func SaveMultiErrorsToMeta() ErrorProcessor {
	return WrapErrProcessor(errpipe.SaveMultiErrorsToMeta())
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
