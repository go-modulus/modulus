package temporal

import (
	"reflect"
	"runtime"
	"strings"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func RegisterActivity(registry worker.Registry, a interface{}) {
	name := getFunctionName(a)

	registry.RegisterActivityWithOptions(
		a, activity.RegisterOptions{
			Name:                          name,
			DisableAlreadyRegisteredCheck: true,
		},
	)
}
func RegisterWorkflow(registry worker.Registry, w interface{}) {
	registry.RegisterWorkflowWithOptions(
		w, workflow.RegisterOptions{
			Name: getFunctionName(w),
		},
	)
}

func getFunctionName(i interface{}) string {
	if fullName, ok := i.(string); ok {
		return fullName
	}
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	return strings.TrimSuffix(fullName, "-fm")
}
