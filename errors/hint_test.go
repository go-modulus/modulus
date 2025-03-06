package errors

import (
	syserrors "errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHint(t *testing.T) {
	t.Run(
		"empty hint for nil error", func(t *testing.T) {
			err := WithHint(nil, "")
			assert.Equal(t, "", Hint(err))
		},
	)

	t.Run(
		"empty hint for golang native error", func(t *testing.T) {
			err := syserrors.New("code")
			assert.Equal(t, "", Hint(err))
		},
	)

	t.Run(
		"hint equals code for default modulus error", func(t *testing.T) {
			err := New("code")
			assert.Equal(t, "code", Hint(err))
		},
	)

	t.Run(
		"hint successfully overridden", func(t *testing.T) {
			err := WithHint(New("code"), "hint")
			assert.Equal(t, "hint", Hint(err))
		},
	)

	t.Run(
		"hint successfully overridden twice", func(t *testing.T) {
			err := WithHint(New("code"), "hint")
			err = WithHint(err, "hint2")
			assert.Equal(t, "hint2", Hint(err))
		},
	)
}
