package debug

import "os"

// This file is used to set the TEMPORAL_DEBUG environment variable to 1
// for all tests in the package. This is useful for debugging purposes.
// The TEMPORAL_DEBUG environment variable is used by the Temporal SDK
// to avoid deadlock error if you are debugging the workflow using breakpoints.
// This is useful for debugging workflows and activities.
//
// Simply import this package in your main_test.go file before temporal importing, because
// Temporal works with envs only in its init function.
func init() {
	_ = os.Setenv("TEMPORAL_DEBUG", "1")
}
