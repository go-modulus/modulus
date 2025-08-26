package erruser_test

import (
	syserrors "errors"
	"testing"

	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/stretchr/testify/assert"
)

func TestNewValidationErrorName(t *testing.T) {
	t.Run(
		"new validation error from several system errors", func(t *testing.T) {
			err1 := syserrors.New("err1")
			err2 := syserrors.New("err2")

			err := erruser.NewValidationError(err1, err2)
			meta := errors.Meta(err)

			assert.Equal(t, "invalid input", err.Error())
			assert.True(t, errors.IsUserError(err))
			assert.Equal(t, "", meta["err1"])
			assert.Equal(t, "", meta["err2"])
			assert.True(t, errors.Is(err, err1))
			assert.True(t, errors.Is(err, err2))
		},
	)

	t.Run(
		"new validation error from several modulus errors", func(t *testing.T) {
			err1 := erruser.New("err1", "Error 1 Hint")
			err2 := erruser.New("err2", "Error 2 Hint")

			err := erruser.NewValidationError(err1, err2)
			meta := errors.Meta(err)

			assert.Equal(t, "invalid input", err.Error())
			assert.True(t, errors.IsUserError(err))
			assert.Equal(t, "Error 1 Hint", meta["err1"])
			assert.Equal(t, "Error 2 Hint", meta["err2"])
			assert.True(t, errors.Is(err, err1))
			assert.True(t, errors.Is(err, err2))
		},
	)
}
