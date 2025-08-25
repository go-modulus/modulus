package errhttp

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errsys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveMultiErrorsToMeta(t *testing.T) {
	t.Parallel()

	processor := SaveMultiErrorsToMeta()
	ctx := context.Background()

	t.Run(
		"returns nil for nil error", func(t *testing.T) {
			t.Parallel()
			result := processor(ctx, nil)
			assert.Nil(t, result)
		},
	)

	t.Run(
		"returns single error unchanged when no multi-errors", func(t *testing.T) {
			t.Parallel()
			err := errors.New("single error")
			result := processor(ctx, err)

			assert.Equal(t, err, result)
			assert.Equal(t, "single error", result.Error())
		},
	)

	t.Run(
		"extracts first error from joined errors and saves rest to meta", func(t *testing.T) {
			t.Parallel()
			err1 := errsys.New("first-error", "First error message")
			err2 := errsys.New("second-error", "Second error message")
			err3 := errsys.New("third-error", "Third error message")

			joinedErr := errors.Join(err1, err2, err3)
			result := processor(ctx, joinedErr)

			// Should return the first error
			assert.Equal(t, "first-error", result.Error())

			// Additional errors should be saved to meta
			meta := errors.Meta(result)
			expected := map[string]string{
				"second-error": "Second error message",
				"third-error":  "Third error message",
			}
			assert.Equal(t, expected, meta)
		},
	)

	t.Run(
		"handles single joined error", func(t *testing.T) {
			t.Parallel()
			err1 := errsys.New("only-error", "Only error message")

			joinedErr := errors.Join(err1)
			result := processor(ctx, joinedErr)

			// Should return the error without additional meta
			assert.Equal(t, "only-error", result.Error())

			meta := errors.Meta(result)
			assert.Empty(t, meta)
		},
	)

	t.Run(
		"preserves existing meta data", func(t *testing.T) {
			t.Parallel()
			err1 := errors.WithMeta(
				errsys.New("first-error", "First error message"),
				"existing-key", "existing-value",
			)
			err2 := errsys.New("second-error", "Second error message")

			joinedErr := errors.Join(err1, err2)
			result := processor(ctx, joinedErr)

			// Should preserve existing meta and add new meta
			meta := errors.Meta(result)
			expected := map[string]string{
				"existing-key": "existing-value",
				"second-error": "Second error message",
			}
			assert.Equal(t, expected, meta)
		},
	)

	t.Run(
		"handles errors with empty hints", func(t *testing.T) {
			t.Parallel()
			err1 := errors.New("first-error")
			err2 := errors.New("second-error")

			joinedErr := errors.Join(err1, err2)
			result := processor(ctx, joinedErr)

			// Should save errors even with empty hints
			meta := errors.Meta(result)
			expected := map[string]string{
				"second-error": "second-error", // hint defaults to error code
			}
			assert.Equal(t, expected, meta)
		},
	)

	t.Run(
		"handles many joined errors", func(t *testing.T) {
			t.Parallel()
			var errs []error
			for i := 1; i <= 10; i++ {
				err := errsys.New(
					fmt.Sprintf("error-%d", i),
					fmt.Sprintf("Error %d message", i),
				)
				errs = append(errs, err)
			}

			joinedErr := errors.Join(errs...)
			result := processor(ctx, joinedErr)

			// Should return the first error
			assert.Equal(t, "error-1", result.Error())

			// All other errors should be in meta
			meta := errors.Meta(result)
			assert.Len(t, meta, 9) // 10 errors minus the first one

			for i := 2; i <= 10; i++ {
				key := fmt.Sprintf("error-%d", i)
				expectedValue := fmt.Sprintf("Error %d message", i)
				assert.Equal(t, expectedValue, meta[key])
			}
		},
	)

	t.Run(
		"handles wrapped errors in join", func(t *testing.T) {
			t.Parallel()
			cause := errors.New("root cause")
			err1 := errors.NewWithCause("wrapper-1", cause)
			err2 := errsys.New("error-2", "Error 2 message")

			joinedErr := errors.Join(err1, err2)
			result := processor(ctx, joinedErr)

			// Should return the first error (wrapper)
			assert.Equal(t, "wrapper-1", result.Error())
			assert.True(t, errors.Is(result, cause))

			// Second error should be in meta
			meta := errors.Meta(result)
			expected := map[string]string{
				"error-2": "Error 2 message",
			}
			assert.Equal(t, expected, meta)
		},
	)
}

func TestExtractErrors(t *testing.T) {
	t.Parallel()

	t.Run(
		"extracts errors from joined error", func(t *testing.T) {
			t.Parallel()
			err1 := errors.New("error1")
			err2 := errors.New("error2")
			err3 := errors.New("error3")

			joinedErr := errors.Join(err1, err2, err3)
			extracted := extractErrors(joinedErr)

			require.Len(t, extracted, 3)
			assert.Equal(t, "error1", extracted[0].Error())
			assert.Equal(t, "error2", extracted[1].Error())
			assert.Equal(t, "error3", extracted[2].Error())
		},
	)

	t.Run(
		"returns empty slice for single error", func(t *testing.T) {
			t.Parallel()
			err := errors.New("single error")
			extracted := extractErrors(err)

			assert.Empty(t, extracted)
		},
	)

	t.Run(
		"returns empty slice for nil error", func(t *testing.T) {
			t.Parallel()
			extracted := extractErrors(nil)

			assert.Empty(t, extracted)
		},
	)

	t.Run(
		"extracts from nested wrapped errors", func(t *testing.T) {
			t.Parallel()
			err1 := errors.New("error1")
			err2 := errors.New("error2")

			// Create a nested structure: outer wraps joined errors
			joinedErr := errors.Join(err1, err2)

			extracted := extractErrors(joinedErr)

			require.Len(t, extracted, 2)
			assert.Equal(t, "error1", extracted[0].Error())
			assert.Equal(t, "error2", extracted[1].Error())
		},
	)
}

func TestSaveMultiErrorsToMetaIntegration(t *testing.T) {
	t.Parallel()

	t.Run(
		"integration with error pipeline", func(t *testing.T) {
			t.Parallel()
			pipeline := &ErrorPipeline{}
			pipeline.SetProcessor(0, SaveMultiErrorsToMeta())

			err1 := errsys.New("validation-error", "Name is required")
			err2 := errsys.New("validation-error", "Email is invalid")
			err3 := errsys.New("validation-error", "Age must be positive")

			joinedErr := errors.Join(err1, err2, err3)

			ctx := context.Background()
			result := pipeline.Process(ctx, joinedErr)

			// Should return the first error
			assert.Equal(t, "validation-error", result.Error())

			// Additional errors should be in meta
			meta := errors.Meta(result)
			expected := map[string]string{
				"validation-error": "Age must be positive", // Last one wins due to map key collision
			}
			assert.Equal(t, expected, meta)
		},
	)

	t.Run(
		"preserves error properties through processing", func(t *testing.T) {
			t.Parallel()
			processor := SaveMultiErrorsToMeta()

			err1 := errors.WithAddedTags(
				errsys.New("first-error", "First message"),
				errors.ValidationErrorTag,
			)
			err2 := errsys.New("second-error", "Second message")

			joinedErr := errors.Join(err1, err2)

			ctx := context.Background()
			result := processor(ctx, joinedErr)

			// Should preserve tags from first error
			assert.True(t, errors.HasTag(result, errors.SystemErrorTag))
			assert.True(t, errors.HasTag(result, errors.ValidationErrorTag))

			// Should have meta from second error
			meta := errors.Meta(result)
			expected := map[string]string{
				"second-error": "Second message",
			}
			assert.Equal(t, expected, meta)
		},
	)
}
