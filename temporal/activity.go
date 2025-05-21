package temporal

import "go.temporal.io/sdk/workflow"

func ExecuteActivity[O any](ctx workflow.Context, activity interface{}, args ...any) Future[O] {
	name := getFunctionName(activity)
	return Future[O]{
		Future: workflow.ExecuteActivity(
			ctx,
			name,
			args...,
		),
	}
}

func WaitActivity[O any](ctx workflow.Context, activity interface{}, input ...any) (O, error) {
	return ExecuteActivity[O](
		ctx,
		activity,
		input...,
	).Get(ctx)
}

func WaitActivityWithoutResult(ctx workflow.Context, activity interface{}, input ...any) error {
	name := getFunctionName(activity)
	return workflow.ExecuteActivity(
		ctx,
		name,
		input...,
	).Get(ctx, nil)
}
