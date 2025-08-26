package errbuilder_test

import (
	syserrors "errors"
	"testing"

	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errbuilder"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Build(t *testing.T) {
	t.Run(
		"simple error from string", func(t *testing.T) {

			b := errbuilder.New("test error")
			err := b.Build()

			t.Log("When create a new error from a string")
			t.Log("	Should return a new error with default code")
			assert.Equal(t, "test error", err.Error())
			t.Log("	Should have system error tag")
			assert.Equal(t, []string{errors.SystemErrorTag}, errors.Tags(err))
			t.Log("	Should return the input as message if there is no translation")
			assert.Equal(t, "test error", errors.Hint(err))
		},
	)

	t.Run(
		"checking error for is clause", func(t *testing.T) {
			b := errbuilder.New("test error").
				WithHint("hint").
				WithMeta("key", "value").
				WithCause(syserrors.New("cause")).
				WithTags("tag1", "tag2")

			err := b.Build()

			errFunc := func() error {
				return err
			}

			err2 := errFunc()

			t.Log("When return a named error")
			t.Log("	Is should return true when comparing the error with the named error")
			assert.ErrorIs(t, err, err2)
		},
	)

	t.Run(
		"simple error from system error", func(t *testing.T) {
			cause := syserrors.New("test error")
			b := errbuilder.NewE(cause)
			err := b.Build()

			t.Log("When create a new error from a string")
			t.Log("	Should return a new error with the error as input")
			assert.Equal(t, "test error", err.Error())
			t.Log("	Should not have any additional fields")
			assert.Nil(t, errors.Tags(err))
			t.Log("	Should return the empty message for the system error")
			assert.Equal(t, "", errors.Hint(err))
			assert.True(t, errors.Is(err, cause))
		},
	)

	t.Run(
		"error with tags", func(t *testing.T) {
			b := errbuilder.New("test error").WithTags("tag1", "tag2")
			err := b.Build()

			err = errors.WithAddedTags(err, "tag3")

			t.Log("When create a new error with tags")
			t.Log("  and add more tags")
			t.Log("	Should return both initial and added tags")
			assert.Len(t, errors.Tags(err), 4)
			assert.Contains(t, errors.Tags(err), "tag1")
			assert.Contains(t, errors.Tags(err), "tag2")
			assert.Contains(t, errors.Tags(err), "tag3")

			assert.Equal(t, "test error", err.Error())
		},
	)

	t.Run(
		"system error with added tags", func(t *testing.T) {
			err := syserrors.New("test error")

			err = errors.WithAddedTags(err, "tag3")

			t.Log("When add tags to a system error")
			t.Log("	Should return added tags")
			assert.Len(t, errors.Tags(err), 1)
			assert.Contains(t, errors.Tags(err), "tag3")

			assert.Equal(t, errors.InternalErrorCode, err.Error())
		},
	)

	t.Run(
		"wrap cause", func(t *testing.T) {
			cause := syserrors.New("test error")

			err := errors.New("custom error")

			err = errors.WithCause(err, cause)

			t.Log("When add a cause to a custom error")
			t.Log("	Should have custom error as the main error")
			assert.Equal(t, "custom error", err.Error())
			t.Log("	Should have the cause as the cause")
			assert.Equal(t, "test error", errors.Cause(err).Error())
			assert.True(t, errors.Is(err, cause))
		},
	)

	t.Run(
		"wrap cause on error with tags", func(t *testing.T) {
			cause := syserrors.New("test error")
			cause = errors.WithAddedTags(cause, "tag3")

			err := errors.New("custom error")
			err = errors.WithAddedTags(err, "tag1", "tag2")

			err = errors.WithCause(err, cause)

			t.Log("When add a cause to an error with tags")
			t.Log("	Should have custom error as the main error")
			assert.Equal(t, "custom error", err.Error())
			t.Log("	Should have the cause as the cause")
			assert.Equal(t, errors.InternalErrorCode, errors.Cause(err).Error())
			assert.True(t, errors.Is(err, cause))
			assert.Len(t, errors.Tags(errors.Cause(err)), 1)
			assert.Contains(t, errors.Tags(errors.Cause(err)), "tag3")

			t.Log("	Should have the tags from the original error")
			assert.Contains(t, errors.Tags(err), "tag1")
			assert.Contains(t, errors.Tags(err), "tag2")
		},
	)
}
