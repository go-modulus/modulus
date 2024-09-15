package temporal

import "go.temporal.io/sdk/workflow"

func ExecuteActivity[O any](ctx workflow.Context, activity string, input any) Future[O] {
	return Future[O]{
		Future: workflow.ExecuteActivity(
			ctx,
			activity,
			input,
		),
	}
}

func WaitActivity[O any](ctx workflow.Context, activity string, input any) (O, error) {
	var output O
	return output, workflow.ExecuteActivity(
		ctx,
		activity,
		input,
	).Get(ctx, &output)
}
