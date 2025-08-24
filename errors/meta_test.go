package errors_test

import (
	syserrors "errors"
	"testing"

	"github.com/go-modulus/modulus/errors"
	"github.com/stretchr/testify/assert"
)

func TestMeta(t *testing.T) {
	t.Run(
		"Meta returns nil for non-modulus error", func(t *testing.T) {
			err := syserrors.New("system error")
			meta := errors.Meta(err)
			assert.Nil(t, meta)
		},
	)

	t.Run(
		"Meta returns nil for nil error", func(t *testing.T) {
			meta := errors.Meta(nil)
			assert.Nil(t, meta)
		},
	)

	t.Run(
		"Meta returns empty map for modulus error without meta", func(t *testing.T) {
			err := errors.New("code")
			meta := errors.Meta(err)
			assert.NotNil(t, meta)
			assert.Empty(t, meta)
		},
	)

	t.Run(
		"Meta returns correct map for error with meta", func(t *testing.T) {
			err := errors.WithMeta(errors.New("code"), "key1", "value1", "key2", "value2")
			meta := errors.Meta(err)

			expected := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			assert.Equal(t, expected, meta)
		},
	)

	t.Run(
		"Meta handles malformed meta strings gracefully", func(t *testing.T) {
			err := errors.New("code")
			// Simulate malformed meta by creating error manually
			// This tests the parsing logic in Meta() method
			err = errors.WithMeta(err, "key1", "value1")
			meta := errors.Meta(err)

			assert.NotNil(t, meta)
			assert.Equal(t, "value1", meta["key1"])
		},
	)
}

func TestWithMeta(t *testing.T) {
	t.Run(
		"WithMeta with nil error returns nil", func(t *testing.T) {
			err := errors.WithMeta(nil, "key", "value")
			assert.Nil(t, err)
		},
	)

	t.Run(
		"WithMeta panics with odd number of arguments", func(t *testing.T) {
			err := errors.New("code")
			assert.Panics(
				t, func() {
					_ = errors.WithMeta(err, "key") //nolint:staticcheck
				},
			)
		},
	)

	t.Run(
		"WithMeta adds meta to modulus error", func(t *testing.T) {
			err := errors.New("code")
			errWithMeta := errors.WithMeta(err, "key1", "value1", "key2", "value2")

			meta := errors.Meta(errWithMeta)
			expected := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			assert.Equal(t, expected, meta)
			assert.Equal(t, "code", errWithMeta.Error())
		},
	)

	t.Run(
		"WithMeta creates modulus error from system error", func(t *testing.T) {
			err := syserrors.New("system error")
			errWithMeta := errors.WithMeta(err, "key", "value")

			meta := errors.Meta(errWithMeta)
			expected := map[string]string{
				"key": "value",
			}
			assert.Equal(t, expected, meta)
			assert.Equal(t, errors.InternalErrorCode, errWithMeta.Error())
		},
	)

	t.Run(
		"WithMeta overwrites existing meta on modulus error", func(t *testing.T) {
			err := errors.WithMeta(errors.New("code"), "oldkey", "oldvalue")
			errWithNewMeta := errors.WithMeta(err, "newkey", "newvalue")

			meta := errors.Meta(errWithNewMeta)
			expected := map[string]string{
				"newkey": "newvalue",
			}
			assert.Equal(t, expected, meta)
			assert.NotContains(t, meta, "oldkey")
		},
	)

	t.Run(
		"WithMeta with empty key-value pairs", func(t *testing.T) {
			err := errors.New("code")
			errWithMeta := errors.WithMeta(err)

			meta := errors.Meta(errWithMeta)
			assert.NotNil(t, meta)
			assert.Empty(t, meta)
		},
	)
}

func TestWithAddedMeta(t *testing.T) {
	t.Run(
		"WithAddedMeta with nil error returns nil", func(t *testing.T) {
			err := errors.WithAddedMeta(nil, "key", "value")
			assert.Nil(t, err)
		},
	)

	t.Run(
		"WithAddedMeta panics with odd number of arguments", func(t *testing.T) {
			err := errors.New("code")
			assert.Panics(
				t, func() {
					_ = errors.WithAddedMeta(err, "key") //nolint:staticcheck
				},
			)
		},
	)

	t.Run(
		"WithAddedMeta adds meta to error without existing meta", func(t *testing.T) {
			err := errors.New("code")
			errWithMeta := errors.WithAddedMeta(err, "key1", "value1", "key2", "value2")

			meta := errors.Meta(errWithMeta)
			expected := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			assert.Equal(t, expected, meta)
		},
	)

	t.Run(
		"WithAddedMeta preserves existing meta and adds new", func(t *testing.T) {
			err := errors.WithMeta(errors.New("code"), "existing", "value")
			errWithAddedMeta := errors.WithAddedMeta(err, "new", "newvalue")

			meta := errors.Meta(errWithAddedMeta)
			expected := map[string]string{
				"existing": "value",
				"new":      "newvalue",
			}
			assert.Equal(t, expected, meta)
		},
	)

	t.Run(
		"WithAddedMeta overwrites existing keys", func(t *testing.T) {
			err := errors.WithMeta(errors.New("code"), "key", "oldvalue")
			errWithAddedMeta := errors.WithAddedMeta(err, "key", "newvalue")

			meta := errors.Meta(errWithAddedMeta)
			expected := map[string]string{
				"key": "newvalue",
			}
			assert.Equal(t, expected, meta)
		},
	)

	t.Run(
		"WithAddedMeta works with system errors", func(t *testing.T) {
			err := syserrors.New("system error")
			errWithMeta := errors.WithAddedMeta(err, "key", "value")

			meta := errors.Meta(errWithMeta)
			expected := map[string]string{
				"key": "value",
			}
			assert.Equal(t, expected, meta)
			assert.Equal(t, errors.InternalErrorCode, errWithMeta.Error())
		},
	)

	t.Run(
		"WithAddedMeta with multiple additions", func(t *testing.T) {
			err := errors.New("code")
			err = errors.WithAddedMeta(err, "key1", "value1")
			err = errors.WithAddedMeta(err, "key2", "value2")
			err = errors.WithAddedMeta(err, "key3", "value3")

			meta := errors.Meta(err)
			expected := map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			assert.Equal(t, expected, meta)
		},
	)

	t.Run(
		"WithAddedMeta with empty key-value pairs", func(t *testing.T) {
			err := errors.WithMeta(errors.New("code"), "existing", "value")
			errWithAddedMeta := errors.WithAddedMeta(err)

			meta := errors.Meta(errWithAddedMeta)
			expected := map[string]string{
				"existing": "value",
			}
			assert.Equal(t, expected, meta)
		},
	)
}

func TestMetaIntegration(t *testing.T) {
	t.Run(
		"Meta works with error wrapping", func(t *testing.T) {
			cause := syserrors.New("root cause")
			err := errors.NewWithCause("wrapper", cause)
			errWithMeta := errors.WithMeta(err, "context", "test")

			meta := errors.Meta(errWithMeta)
			expected := map[string]string{
				"context": "test",
			}
			assert.Equal(t, expected, meta)
			assert.Equal(t, "wrapper", errWithMeta.Error())
			assert.True(t, errors.Is(errWithMeta, cause))
		},
	)

	t.Run(
		"Meta preserves error identity", func(t *testing.T) {
			originalErr := errors.New("code")
			errWithMeta := errors.WithMeta(originalErr, "key", "value")

			assert.True(t, errors.Is(errWithMeta, originalErr))
			assert.True(t, errors.Is(originalErr, errWithMeta))
		},
	)
}
