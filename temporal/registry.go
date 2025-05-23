package temporal

import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"reflect"
	"runtime"
	"strings"
)

func RegisterActivity(registry worker.Registry, a interface{}) {
	registry.RegisterActivityWithOptions(
		a, activity.RegisterOptions{
			Name: getFunctionName(a),
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
