package errors

import (
	"context"
	"slices"

	"github.com/go-modulus/modulus/errors"
	temporalInterceptor "go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/temporal"
)

const errTagNonRetriable = "temporal_non_retriable_error"

func NonRetriable(err error) error {
	return errors.WithAddedTags(err, errTagNonRetriable)
}

type ErrDetails struct {
	Stacktrace []string          `json:"stacktrace"`
	Meta       map[string]string `json:"meta"`
}

// Наш воркер-интерсептор
type AppErrWrapWorkerInterceptor struct {
	temporalInterceptor.WorkerInterceptorBase
}

func (w *AppErrWrapWorkerInterceptor) InterceptActivity(
	ctx context.Context,
	next temporalInterceptor.ActivityInboundInterceptor,
) temporalInterceptor.ActivityInboundInterceptor {
	base := temporalInterceptor.ActivityInboundInterceptorBase{
		Next: next,
	}
	i := &appErrWrapActivityInbound{
		ActivityInboundInterceptorBase: base,
	}
	return i
}

type appErrWrapActivityInbound struct {
	temporalInterceptor.ActivityInboundInterceptorBase
}

func (a *appErrWrapActivityInbound) Init(outbound temporalInterceptor.ActivityOutboundInterceptor) error {
	return a.ActivityInboundInterceptorBase.Init(outbound)
}

func (a *appErrWrapActivityInbound) ExecuteActivity(
	ctx context.Context, in *temporalInterceptor.ExecuteActivityInput,
) (any, error) {
	res, err := a.ActivityInboundInterceptorBase.Next.ExecuteActivity(ctx, in)
	if err == nil {
		return res, nil
	}

	// Do not wrap application errors
	var app *temporal.ApplicationError
	if errors.As(err, &app) {
		return res, err
	}

	// Do not wrap timeout, canceled, and panic errors
	var (
		tm  *temporal.TimeoutError
		cn  *temporal.CanceledError
		pnc *temporal.PanicError
	)
	if errors.As(err, &tm) || errors.As(err, &cn) || errors.As(err, &pnc) {
		return res, err
	}

	tags := errors.Tags(err)
	if slices.Contains(tags, errTagNonRetriable) {
		return res, temporal.NewNonRetryableApplicationError(
			errors.Hint(err),
			err.Error(), // outer type
			err,         // cause
			ErrDetails{
				Stacktrace: errors.Trace(err),
				Meta:       errors.Meta(err),
			},
		)
	}

	// Wrap all other errors
	wrapped := temporal.NewApplicationErrorWithCause(
		errors.Hint(err),
		err.Error(), // outer type
		err,         // cause
		ErrDetails{
			Stacktrace: errors.Trace(err),
			Meta:       errors.Meta(err),
		},
	)

	return res, wrapped
}
