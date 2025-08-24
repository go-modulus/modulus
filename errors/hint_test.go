package errors_test

import (
	syserrors "errors"
	"testing"

	"github.com/go-modulus/modulus/errors"
	"github.com/stretchr/testify/assert"
)

func TestHint(t *testing.T) {
	t.Run(
		"empty hint for nil error", func(t *testing.T) {
			err := errors.WithHint(nil, "")
			assert.Equal(t, "", errors.Hint(err))
		},
	)

	t.Run(
		"empty hint for golang native error", func(t *testing.T) {
			err := syserrors.New("code")
			assert.Equal(t, "", errors.Hint(err))
		},
	)

	t.Run(
		"hint equals code for default modulus error", func(t *testing.T) {
			err := errors.New("code")
			assert.Equal(t, "code", errors.Hint(err))
		},
	)

	t.Run(
		"hint successfully overridden", func(t *testing.T) {
			err := errors.WithHint(errors.New("code"), "hint")
			assert.Equal(t, "hint", errors.Hint(err))
		},
	)

	t.Run(
		"hint successfully overridden twice", func(t *testing.T) {
			err := errors.WithHint(errors.New("code"), "hint")
			err = errors.WithHint(err, "hint2")
			assert.Equal(t, "hint2", errors.Hint(err))
		},
	)

	t.Run(
		"create new mError on hinting the system error", func(t *testing.T) {
			err := syserrors.New("code")
			hintedErr := errors.WithHint(err, "hint")

			assert.Equal(t, "hint", errors.Hint(hintedErr))
			assert.Equal(t, "code", errors.Cause(hintedErr).Error())
			assert.Equal(t, errors.InternalErrorCode, hintedErr.Error())
			assert.Len(t, errors.Trace(hintedErr), 1)
		},
	)
}
