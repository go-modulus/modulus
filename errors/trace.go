package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"braces.dev/errtrace"
)

// Trace returns the trace of the error.
// You must wrap the error with errtrace.Wrap before calling this function.
func Trace(err error) []string {
	if err == nil {
		return nil
	}
	allCauses := make([]error, 0)
	allCauses = append(allCauses, err)
	for cause := errors.Unwrap(err); cause != nil; cause = errors.Unwrap(cause) {
		allCauses = append(allCauses, cause)
	}
	trace := make([]string, 0)
	for i := len(allCauses) - 1; i >= 0; i-- {
		cause := allCauses[i]
		items := getErrTraceTrace(cause)

		items = append(items, getMErrorTrace(cause)...)

		if len(items) == 0 {
			continue
		}
		trace = append(trace, items...)
	}

	// remove duplicates
	seen := make(map[string]bool)
	unique := make([]string, 0)
	for _, item := range trace {
		if !seen[item] {
			seen[item] = true
			unique = append(unique, item)
		}
	}

	return unique
}

func getErrTraceTrace(err error) []string {
	message := errtrace.FormatString(err)
	msgParts := strings.Split(message, "\n")
	var trace []string
	for i, part := range msgParts {
		// remove
		if i == 0 || part == "" {
			continue
		}
		trace = append(trace, part)
	}
	return trace
}

func getMErrorTrace(err error) []string {
	result := make([]string, 0)
	if err == nil {
		return result
	}
	var e mError
	if errors.As(err, &e) {
		if e.trace != "" {
			result = strings.Split(e.trace, "\n")
		}
		return result
	}
	return result
}
func WithTrace(err error) error {
	return WithTraceSkip(err, 2)
}
func WithTraceSkip(err error, skipFrames int) error {
	traceItem := ""
	_, file, line, ok := runtime.Caller(skipFrames)
	if ok {
		traceItem = fmt.Sprintf("%s:%d", file, line)
	}

	e := copyErr(err)

	if e.trace != "" {
		traceItem = e.trace + "\n" + traceItem
	}
	e.trace = traceItem
	return e
}
