package temporal

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

func NewFuture[T any](ctx workflow.Context) (Future[T], Settable[T]) {
	future, settable := workflow.NewFuture(ctx)
	return Future[T]{Future: future}, Settable[T]{Settable: settable}
}

type Future[T any] struct {
	Future workflow.Future
}

func (f Future[T]) Get(ctx workflow.Context) (T, error) {
	var t T
	return t, f.Future.Get(ctx, &t)
}

func (f Future[T]) GetOrDefault(ctx workflow.Context) T {
	var t T
	if f.IsReady() {
		_ = f.Future.Get(ctx, &t)
	}
	return t
}

func (f Future[T]) GetWithTimeout(ctx workflow.Context, timeout time.Duration) (T, error) {
	var t T
	if !f.IsReady() {
		_, err := workflow.AwaitWithTimeout(
			ctx,
			timeout,
			func() bool { return f.IsReady() },
		)
		if err != nil {
			return t, err
		}
	}

	return t, f.Future.Get(ctx, &t)
}

func (f Future[T]) IsReady() bool {
	return f.Future.IsReady()
}

type Settable[T any] struct {
	Settable workflow.Settable
}

func (f Settable[T]) Set(value T, err error) {
	f.Settable.Set(value, err)
}

func (f Settable[T]) SetValue(value T) {
	f.Settable.SetValue(value)
}

func (f Settable[T]) SetError(err error) {
	f.Settable.SetError(err)
}
