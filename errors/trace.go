package errors

import (
	"strings"

	"braces.dev/errtrace"
)

// Trace returns the trace of the error.
// You must wrap the error with errtrace.Wrap before calling this function.
func Trace(err error) []string {
	message := errtrace.FormatString(err)
	msgParts := strings.Split(message, "\n")
	var trace []string
	for i, part := range msgParts {
		if i == 0 || part == "" {
			continue
		}
		trace = append(trace, part)
	}
	return trace
}
