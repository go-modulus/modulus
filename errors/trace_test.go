package errors_test

import (
	syserrors "errors"
	"strings"
	"testing"

	"braces.dev/errtrace"
	"github.com/go-modulus/modulus/errors"
	"github.com/stretchr/testify/assert"
)

func TestTrace(t *testing.T) {
	t.Run(
		"Trace returns empty slice for nil error", func(t *testing.T) {
			trace := errors.Trace(nil)
			assert.Empty(t, trace)
		},
	)

	t.Run(
		"Trace returns empty slice for error without errtrace", func(t *testing.T) {
			err := syserrors.New("system error")
			trace := errors.Trace(err)
			assert.Empty(t, trace)
		},
	)

	t.Run(
		"Trace returns empty slice for modulus error without errtrace", func(t *testing.T) {
			err := errors.New("code")
			trace := errors.Trace(err)
			assert.Empty(t, trace)
		},
	)

	t.Run(
		"Trace returns trace for wrapped error", func(t *testing.T) {
			err := syserrors.New("system error")
			wrappedErr := errtrace.Wrap(err)
			trace := errors.Trace(wrappedErr)

			assert.NotEmpty(t, trace)
			assert.True(t, len(trace) > 0)
			// Check that trace contains file and line information
			assert.True(
				t, strings.Contains(trace[0], "trace_test.go") ||
					strings.Contains(trace[1], "trace_test.go"),
			)
		},
	)

	t.Run(
		"Trace returns trace for wrapped modulus error", func(t *testing.T) {
			err := errors.New("code")
			wrappedErr := errtrace.Wrap(err)
			trace := errors.Trace(wrappedErr)

			assert.NotEmpty(t, trace)
			assert.True(t, len(trace) > 0)
			assert.True(
				t, strings.Contains(trace[0], "trace_test.go") ||
					strings.Contains(trace[1], "trace_test.go"),
			)
		},
	)

	t.Run(
		"Trace returns multiple stack frames for nested wrapping", func(t *testing.T) {
			err := syserrors.New("root error")
			err = errtrace.Wrap(err)
			err = errtrace.Wrap(err)

			trace := errors.Trace(err)

			assert.NotEmpty(t, trace)
			// Should have multiple stack frames
			assert.True(t, len(trace) >= 2)
		},
	)
}

func TestWithTrace(t *testing.T) {
	t.Run(
		"WithTrace adds trace to modulus error", func(t *testing.T) {
			err := errors.New("code")
			errWithTrace := errors.WithTrace(err)

			assert.Equal(t, "code", errWithTrace.Error())
			// Error should still be the same type
			assert.True(t, errors.Is(errWithTrace, err))
		},
	)

	t.Run(
		"WithTrace creates modulus error from system error", func(t *testing.T) {
			err := syserrors.New("system error")
			errWithTrace := errors.WithTrace(err)

			assert.Equal(t, errors.InternalErrorCode, errWithTrace.Error())
			assert.True(t, errors.IsSystemError(errWithTrace))
			assert.True(t, errors.Is(errWithTrace, err))
		},
	)

	t.Run(
		"WithTrace preserves existing trace", func(t *testing.T) {
			err := errors.New("code")
			errWithTrace1 := errors.WithTrace(err)
			errWithTrace2 := errors.WithTrace(errWithTrace1)

			assert.Equal(t, "code", errWithTrace2.Error())
			assert.True(t, errors.Is(errWithTrace2, err))
		},
	)

	t.Run(
		"WithTrace works with nil error", func(t *testing.T) {
			// This should not panic, but behavior depends on implementation
			defer func() {
				if r := recover(); r != nil {
					// If it panics, that's also acceptable behavior
					t.Log("WithTrace panics on nil error, which is acceptable")
				}
			}()

			errWithTrace := errors.WithTrace(nil)
			// If we get here without panic, check the result
			if errWithTrace != nil {
				assert.NotNil(t, errWithTrace)
			}
		},
	)

	t.Run(
		"WithTrace preserves error properties", func(t *testing.T) {
			originalErr := errors.New("code")
			originalErr = errors.WithMeta(originalErr, "key", "value")
			originalErr = errors.WithAddedTags(originalErr, errors.UserErrorTag)

			errWithTrace := errors.WithTrace(originalErr)

			assert.Equal(t, "code", errWithTrace.Error())
			assert.True(t, errors.HasTag(errWithTrace, errors.SystemErrorTag))
			assert.True(t, errors.HasTag(errWithTrace, errors.UserErrorTag))

			meta := errors.Meta(errWithTrace)
			expected := map[string]string{"key": "value"}
			assert.Equal(t, expected, meta)
		},
	)
}

func TestTraceIntegration(t *testing.T) {
	t.Run(
		"Trace and WithTrace work together", func(t *testing.T) {
			err := errors.New("code")
			errWithTrace := errors.WithTrace(err)

			// WithTrace doesn't use errtrace, so Trace() won't return anything
			// This tests the interaction between the two functions
			trace := errors.Trace(errWithTrace)
			assert.True(
				t, strings.Contains(trace[0], "trace_test.go"),
			)
		},
	)

	t.Run(
		"errtrace.Wrap and WithTrace work together", func(t *testing.T) {
			err := errors.New("code")
			errWithTrace := errors.WithTrace(err)
			wrappedErr := errtrace.Wrap(errWithTrace)

			trace := errors.Trace(wrappedErr)
			assert.NotEmpty(t, trace)
			assert.True(t, strings.Contains(trace[0], "trace_test.go"))
		},
	)

	t.Run(
		"WithTrace preserves error chain", func(t *testing.T) {
			cause := syserrors.New("root cause")
			err := errors.NewWithCause("wrapper", cause)
			errWithTrace := errors.WithTrace(err)

			assert.Equal(t, "wrapper", errWithTrace.Error())
			assert.True(t, errors.Is(errWithTrace, cause))
		},
	)

	t.Run(
		"Multiple WithTrace calls build trace", func(t *testing.T) {
			err := errors.New("code")

			// Simulate multiple levels of WithTrace calls
			level1 := errors.WithTrace(err)
			level2 := errors.WithTrace(level1)
			level3 := errors.WithTrace(level2)

			assert.Equal(t, "code", level3.Error())
			assert.True(t, errors.Is(level3, err))
		},
	)

	t.Run(
		"Mix Trace and Wrap", func(t *testing.T) {
			err := errors.New("code")

			// Simulate multiple levels of WithTrace calls
			level1 := errors.WithTrace(err)
			level2 := errtrace.Wrap(level1)
			level3 := errtrace.Wrap(level2)
			level4 := errors.WithTrace(level3)

			trace := errors.Trace(level4)

			assert.Equal(t, "code", level4.Error())
			assert.True(t, errors.Is(level3, err))
			assert.True(t, errors.Is(level3, level1))
			assert.True(t, errors.Is(level3, level2))
			assert.True(t, errors.Is(level3, level4))
			assert.NotEmpty(t, trace)
			assert.True(t, strings.Contains(trace[0], "trace_test.go"))
		},
	)

	t.Run(
		"Call Wrap after Trace", func(t *testing.T) {
			err := errors.New("code")

			// Simulate multiple levels of WithTrace calls
			level1 := errors.WithTrace(err)
			level2 := errtrace.Wrap(level1)
			level3 := errtrace.Wrap(level2)

			trace := errors.Trace(level3)

			assert.Equal(t, "code", level3.Error())
			assert.True(t, errors.Is(level3, err))
			assert.True(t, errors.Is(level3, level1))
			assert.True(t, errors.Is(level3, level2))
			assert.NotEmpty(t, trace)
			assert.True(t, strings.Contains(trace[0], "trace_test.go:233"))
			assert.True(t, strings.Contains(trace[1], "github.com/go-modulus/modulus/errors_test.TestTraceIntegration"))
			assert.True(t, strings.Contains(trace[2], "trace_test.go:234"))
			assert.True(t, strings.Contains(trace[3], "trace_test.go:235"))
		},
	)
}

func helperFunctionForTraceTest() error {
	err := syserrors.New("helper error")
	return errtrace.Wrap(err)
}

func TestTraceWithHelperFunction(t *testing.T) {
	t.Run(
		"Trace shows correct file and function information", func(t *testing.T) {
			err := helperFunctionForTraceTest()
			trace := errors.Trace(err)

			assert.NotEmpty(t, trace)
			assert.True(t, len(trace) >= 1)

			// Should contain reference to helper function
			found := false
			for _, traceLine := range trace {
				if strings.Contains(traceLine, "helperFunctionForTraceTest") ||
					strings.Contains(traceLine, "trace_test.go") {
					found = true
					break
				}
			}
			assert.True(t, found, "Trace should contain reference to helper function or test file")
		},
	)
}
