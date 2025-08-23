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

	// make a trace saved by errtrace.Wrap
	message := errtrace.FormatString(err)
	msgParts := strings.Split(message, "\n")
	var trace []string
	for i, part := range msgParts {
		if i == 0 || part == "" {
			continue
		}
		trace = append(trace, part)
	}

	// append trace saved by WithTrace
	trace = append(trace, getTrace(err)...)

	return trace
}

func getTrace(err error) []string {
	result := make([]string, 0)
	if err == nil {
		return result
	}
	var e mError
	if errors.As(err, &e) {
		if e.trace != "" {
			result = strings.Split(e.trace, "\n")
		}
		result = append(result, getTrace(e.cause)...)
		return result
	}
	return result
}

func WithTrace(err error) error {
	traceItem := ""
	_, file, line, ok := runtime.Caller(1)
	if ok {
		traceItem = fmt.Sprintf("%s:%d", file, line)
	}

	var e mError
	if errors.As(err, &e) {
		if e.trace != "" {
			traceItem = e.trace + "\n" + traceItem
		}
		e.trace = traceItem
		return e
	}
	e = new(err.Error())
	e.trace = traceItem
	return e
}
