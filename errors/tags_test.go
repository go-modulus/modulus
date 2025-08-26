package errors_test

import (
	syserrors "errors"
	"testing"

	"braces.dev/errtrace"
	"github.com/go-modulus/modulus/errors"
	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
	t.Run(
		"Tags returns nil for nil error", func(t *testing.T) {
			tags := errors.Tags(nil)
			assert.Nil(t, tags)
		},
	)

	t.Run(
		"Tags returns nil for non-modulus error", func(t *testing.T) {
			err := syserrors.New("system error")
			tags := errors.Tags(err)
			assert.Nil(t, tags)
		},
	)

	t.Run(
		"Tags returns system error tag for new modulus error", func(t *testing.T) {
			err := errors.New("code")
			tags := errors.Tags(err)
			expected := []string{errors.SystemErrorTag}
			assert.Equal(t, expected, tags)
		},
	)

	t.Run(
		"Tags returns multiple tags", func(t *testing.T) {
			err := errors.WithAddedTags(errors.New("code"), errors.UserErrorTag, errors.ValidationErrorTag)
			tags := errors.Tags(err)
			expected := []string{errors.SystemErrorTag, errors.UserErrorTag, errors.ValidationErrorTag}
			assert.Equal(t, expected, tags)
		},
	)

	t.Run(
		"Tags handles empty tags string", func(t *testing.T) {
			err := errors.New("code")
			err = errors.WithAddedTags(err, "")
			tags := errors.Tags(err)
			expected := []string{errors.SystemErrorTag, ""}
			assert.Equal(t, expected, tags)
		},
	)

	t.Run(
		"Tags returns multiple tags when error is wrapped", func(t *testing.T) {
			baseErr := errors.New("code")
			err := errors.WithAddedTags(baseErr, errors.UserErrorTag, errors.ValidationErrorTag)
			err = errtrace.Wrap(err)
			err = errors.WithAddedTags(err, "test")
			tags := errors.Tags(err)
			trace := errors.Trace(err)
			expected := []string{errors.SystemErrorTag, errors.UserErrorTag, errors.ValidationErrorTag, "test"}
			assert.Equal(t, expected, tags)
			assert.True(t, errors.Is(err, baseErr))
			assert.Equal(t, 2, len(trace))
		},
	)
}

func TestHasTag(t *testing.T) {
	t.Run(
		"HasTag returns false for nil error", func(t *testing.T) {
			hasTag := errors.HasTag(nil, errors.SystemErrorTag)
			assert.False(t, hasTag)
		},
	)

	t.Run(
		"HasTag returns false for non-modulus error", func(t *testing.T) {
			err := syserrors.New("system error")
			hasTag := errors.HasTag(err, errors.SystemErrorTag)
			assert.False(t, hasTag)
		},
	)

	t.Run(
		"HasTag returns true for system error tag on new modulus error", func(t *testing.T) {
			err := errors.New("code")
			hasTag := errors.HasTag(err, errors.SystemErrorTag)
			assert.True(t, hasTag)
		},
	)

	t.Run(
		"HasTag returns false for tag not present", func(t *testing.T) {
			err := errors.New("code")
			hasTag := errors.HasTag(err, errors.UserErrorTag)
			assert.False(t, hasTag)
		},
	)

	t.Run(
		"HasTag returns true for added tag", func(t *testing.T) {
			err := errors.WithAddedTags(errors.New("code"), errors.UserErrorTag)
			hasTag := errors.HasTag(err, errors.UserErrorTag)
			assert.True(t, hasTag)
		},
	)

	t.Run(
		"HasTag returns true for multiple tags", func(t *testing.T) {
			err := errors.WithAddedTags(errors.New("code"), errors.UserErrorTag, errors.ValidationErrorTag)
			assert.True(t, errors.HasTag(err, errors.SystemErrorTag))
			assert.True(t, errors.HasTag(err, errors.UserErrorTag))
			assert.True(t, errors.HasTag(err, errors.ValidationErrorTag))
		},
	)

	t.Run(
		"HasTag returns false for partial match", func(t *testing.T) {
			err := errors.WithAddedTags(errors.New("code"), "custom-tag")
			hasTag := errors.HasTag(err, "custom")
			assert.False(t, hasTag)
		},
	)
}

func TestWithAddedTags(t *testing.T) {
	t.Run(
		"WithAddedTags with nil error returns nil", func(t *testing.T) {
			err := errors.WithAddedTags(nil, errors.UserErrorTag)
			assert.Nil(t, err)
		},
	)

	t.Run(
		"WithAddedTags adds single tag to modulus error", func(t *testing.T) {
			err := errors.New("code")
			errWithTag := errors.WithAddedTags(err, errors.UserErrorTag)

			tags := errors.Tags(errWithTag)
			expected := []string{errors.SystemErrorTag, errors.UserErrorTag}
			assert.Equal(t, expected, tags)
			assert.Equal(t, "code", errWithTag.Error())
		},
	)

	t.Run(
		"WithAddedTags adds multiple tags to modulus error", func(t *testing.T) {
			err := errors.New("code")
			errWithTags := errors.WithAddedTags(err, errors.UserErrorTag, errors.ValidationErrorTag)

			tags := errors.Tags(errWithTags)
			expected := []string{errors.SystemErrorTag, errors.UserErrorTag, errors.ValidationErrorTag}
			assert.Equal(t, expected, tags)
		},
	)

	t.Run(
		"WithAddedTags creates modulus error from system error", func(t *testing.T) {
			err := syserrors.New("system error")
			errWithTag := errors.WithAddedTags(err, errors.UserErrorTag)

			tags := errors.Tags(errWithTag)
			expected := []string{errors.UserErrorTag}
			assert.Equal(t, expected, tags)
			assert.Equal(t, errors.InternalErrorCode, errWithTag.Error())
		},
	)

	t.Run(
		"WithAddedTags preserves existing tags", func(t *testing.T) {
			err := errors.WithAddedTags(errors.New("code"), errors.UserErrorTag)
			errWithMoreTags := errors.WithAddedTags(err, errors.ValidationErrorTag)

			tags := errors.Tags(errWithMoreTags)
			expected := []string{errors.SystemErrorTag, errors.UserErrorTag, errors.ValidationErrorTag}
			assert.Equal(t, expected, tags)
		},
	)

	t.Run(
		"WithAddedTags shows only unique tags", func(t *testing.T) {
			err := errors.New("code")
			errWithDuplicates := errors.WithAddedTags(err, errors.SystemErrorTag, errors.UserErrorTag)

			tags := errors.Tags(errWithDuplicates)
			expected := []string{errors.SystemErrorTag, errors.UserErrorTag}
			assert.Equal(t, expected, tags)
		},
	)

	t.Run(
		"WithAddedTags with no tags", func(t *testing.T) {
			err := errors.New("code")
			errWithNoNewTags := errors.WithAddedTags(err)

			tags := errors.Tags(errWithNoNewTags)
			expected := []string{errors.SystemErrorTag}
			assert.Equal(t, expected, tags)
		},
	)

	t.Run(
		"WithAddedTags with custom tags", func(t *testing.T) {
			err := errors.New("code")
			errWithCustomTags := errors.WithAddedTags(err, "custom-tag", "another-tag")

			tags := errors.Tags(errWithCustomTags)
			expected := []string{errors.SystemErrorTag, "custom-tag", "another-tag"}
			assert.Equal(t, expected, tags)

			assert.True(t, errors.HasTag(errWithCustomTags, "custom-tag"))
			assert.True(t, errors.HasTag(errWithCustomTags, "another-tag"))
		},
	)
}

func TestTagsIntegration(t *testing.T) {
	t.Run(
		"Tags work with error wrapping", func(t *testing.T) {
			cause := syserrors.New("root cause")
			err := errors.NewWithCause("wrapper", cause)
			errWithTags := errors.WithAddedTags(err, errors.UserErrorTag)

			tags := errors.Tags(errWithTags)
			expected := []string{errors.SystemErrorTag, errors.UserErrorTag}
			assert.Equal(t, expected, tags)
			assert.Equal(t, "wrapper", errWithTags.Error())
			assert.True(t, errors.Is(errWithTags, cause))
		},
	)

	t.Run(
		"Tags preserve error identity", func(t *testing.T) {
			originalErr := errors.New("code")
			errWithTags := errors.WithAddedTags(originalErr, errors.UserErrorTag)

			assert.True(t, errors.Is(errWithTags, originalErr))
			assert.True(t, errors.Is(originalErr, errWithTags))
		},
	)

	t.Run(
		"Tags work with meta data", func(t *testing.T) {
			err := errors.New("code")
			err = errors.WithMeta(err, "key", "value")
			err = errors.WithAddedTags(err, errors.UserErrorTag)

			tags := errors.Tags(err)
			meta := errors.Meta(err)

			expected := []string{errors.SystemErrorTag, errors.UserErrorTag}
			assert.Equal(t, expected, tags)
			assert.Equal(t, map[string]string{"key": "value"}, meta)
		},
	)

	t.Run(
		"Multiple tag additions", func(t *testing.T) {
			err := errors.New("code")
			err = errors.WithAddedTags(err, "tag1")
			err = errors.WithAddedTags(err, "tag2")
			err = errors.WithAddedTags(err, "tag3")

			tags := errors.Tags(err)
			expected := []string{errors.SystemErrorTag, "tag1", "tag2", "tag3"}
			assert.Equal(t, expected, tags)

			assert.True(t, errors.HasTag(err, "tag1"))
			assert.True(t, errors.HasTag(err, "tag2"))
			assert.True(t, errors.HasTag(err, "tag3"))
			assert.False(t, errors.HasTag(err, "tag4"))
		},
	)
}

func TestTagConstants(t *testing.T) {
	t.Run(
		"Tag constants are defined correctly", func(t *testing.T) {
			assert.Equal(t, "system-error", errors.SystemErrorTag)
			assert.Equal(t, "user-error", errors.UserErrorTag)
			assert.Equal(t, "validation-error", errors.ValidationErrorTag)
		},
	)
}
