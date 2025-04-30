package temporal

import (
	"context"
	"errors"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"time"
)

type WaitGroup struct {
	wg   workflow.WaitGroup
	errs []error
}

func NewWaitGroup(ctx workflow.Context) *WaitGroup {
	return &WaitGroup{
		wg: workflow.NewWaitGroup(ctx),
	}
}

func (w *WaitGroup) Go(ctx workflow.Context, f func(ctx workflow.Context) error) {
	w.wg.Add(1)
	workflow.Go(
		ctx,
		func(ctx workflow.Context) {
			defer w.wg.Done()
			err := f(ctx)
			if err != nil {
				w.errs = append(w.errs, err)
			}
		},
	)
}

func (w *WaitGroup) Wait(ctx workflow.Context) error {
	w.wg.Wait(ctx)
	return errors.Join(w.errs...)
}

type Channel[T any] struct {
	channel workflow.Channel
}

func NewChannel[T any](ctx workflow.Context) Channel[T] {
	return Channel[T]{channel: workflow.NewChannel(ctx)}
}

func (c Channel[T]) Receive(ctx workflow.Context) (t T, more bool) {
	return t, c.channel.Receive(ctx, &t)
}

func (c Channel[T]) ReceiveWithTimeout(ctx workflow.Context, timeout time.Duration) (t T, ok bool, more bool) {
	ok, more = c.channel.ReceiveWithTimeout(ctx, timeout, &t)
	return
}

func (c Channel[T]) Send(ctx workflow.Context, value T) {
	c.channel.Send(ctx, value)
}

func (c Channel[T]) SendAsync(value T) (ok bool) {
	return c.channel.SendAsync(value)
}

func (c Channel[T]) Close() {
	c.channel.Close()
}

func SideEffect[T any](ctx workflow.Context, f func(ctx workflow.Context) T) (T, error) {
	encodedT := workflow.SideEffect(
		ctx, func(ctx workflow.Context) interface{} {
			return f(ctx)
		},
	)
	var t T
	return t, encodedT.Get(&t)
}

func ExecuteWorkflow(
	ctx context.Context,
	starter Starter,
	options client.StartWorkflowOptions,
	workflow interface{},
	args ...interface{},
) (client.WorkflowRun, error) {
	name := getFunctionName(workflow)
	return starter.ExecuteWorkflow(
		ctx,
		options,
		name,
		args...,
	)
}

func WorkflowName(w interface{}) string {
	return getFunctionName(w)
}
