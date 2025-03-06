package errors_test

import (
	syserrors "errors"
	"github.com/go-modulus/modulus/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run(
		"New modulus error", func(t *testing.T) {
			err := errors.New("code")

			assert.True(t, errors.IsSystemError(err))
			assert.Equal(t, "code", err.Error())
			assert.Equal(t, "code", errors.Hint(err))
		},
	)

	t.Run(
		"New system error", func(t *testing.T) {
			err := syserrors.New("code")

			assert.True(t, errors.IsSystemError(err))
			assert.Equal(t, "code", err.Error())
			assert.Equal(t, "", errors.Hint(err))
		},
	)
}

func TestIs(t *testing.T) {
	t.Run(
		"Is modulus errors are equal", func(t *testing.T) {
			err := errors.New("code")
			target := err

			assert.True(t, errors.Is(err, target))
		},
	)

	t.Run(
		"Is with meta errors are equal", func(t *testing.T) {
			err := errors.WithAddedMeta(errors.New("code"), "key", "value")
			target := err

			assert.True(t, errors.Is(err, target))
		},
	)
}
