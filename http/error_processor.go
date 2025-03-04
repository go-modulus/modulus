package http

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errlog"
	context2 "github.com/go-modulus/modulus/http/context"
	"log/slog"
)

const InternalErrorCode = "unhandled internal error"

type ErrorProcessor func(ctx context.Context, err error) error

type ErrorPipeline struct {
	Processors []ErrorProcessor
}

func NewDefaultErrorPipeline(logger *slog.Logger) *ErrorPipeline {
	return &ErrorPipeline{
		Processors: []ErrorProcessor{
			LogError(logger),
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
			resultErr := errors.NewSysError(InternalErrorCode, "Something went wrong")
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

func LogError(logger *slog.Logger) ErrorProcessor {
	return func(ctx context.Context, err error) error {
		if err == nil {
			return nil
		}
		errlog.LogError(ctx, err, logger)
		return err
	}
}
