package errsys_test

import (
	"braces.dev/errtrace"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errsys"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run(
		"new system error", func(t *testing.T) {
			err := errsys.New("code", "hint")
			target := err

			assert.True(t, errors.IsSystemError(err))
			assert.Equal(t, "code", err.Error())
			assert.Equal(t, "hint", errors.Hint(err))
			assert.True(t, errors.Is(err, target))
		},
	)

	t.Run(
		"new system error with meta", func(t *testing.T) {
			errInit := errsys.New("code", "hint")

			err := errors.WithAddedMeta(errInit, "key", "value")
			err = errors.WithAddedMeta(err, "key2", "value2")
			target := errtrace.Wrap(err)

			assert.True(t, errors.IsSystemError(err))
			assert.Equal(t, "code", err.Error())
			assert.Equal(t, "hint", errors.Hint(err))
			assert.Equal(t, "value", errors.Meta(err)["key"])
			assert.Equal(t, "value2", errors.Meta(err)["key2"])
			assert.True(t, errors.Is(errInit, target))
		},
	)
}

func TestNewWithCause(t *testing.T) {
	t.Run(
		"new system error with cause", func(t *testing.T) {
			errCause := errsys.New("cause", "hint")
			err := errsys.NewWithCause("code", "hint", errCause)

			assert.True(t, errors.IsSystemError(err))
			assert.Equal(t, "code", err.Error())
			assert.Equal(t, "hint", errors.Hint(err))
			assert.Equal(t, errCause, errors.Cause(err))
		},
	)

	t.Run(
		"new system error with cause and meta", func(t *testing.T) {
			errCause := errsys.New("cause", "hint")
			errInit := errsys.NewWithCause("code", "hint", errCause)

			err := errors.WithAddedMeta(errInit, "key", "value")

			assert.True(t, errors.IsSystemError(err))
			assert.Equal(t, "code", err.Error())
			assert.Equal(t, "hint", errors.Hint(err))
			assert.Equal(t, errCause, errors.Cause(err))
			assert.Equal(t, "value", errors.Meta(err)["key"])
		},
	)
}

func TestWithCause(t *testing.T) {
	t.Run(
		"new system error with cause", func(t *testing.T) {
			errCause := errsys.New("cause", "hint")
			errNew := errsys.New("code", "hint2")
			err := errsys.WithCause(errNew, errCause)

			assert.True(t, errors.IsSystemError(err))
			assert.Equal(t, "code", err.Error())
			assert.Equal(t, "hint2", errors.Hint(err))
			assert.Equal(t, errCause, errors.Cause(err))
		},
	)
}
