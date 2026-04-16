package errlog_test

import (
	"errors"
	"testing"

	"github.com/go-modulus/modulus/errors/errlog"
	"github.com/stretchr/testify/assert"
)

func TestMeta(t *testing.T) {
	t.Parallel()
	err := errlog.Wrap(errors.New("test error"), "user", "aboba")
	assert.Equal(
		t,
		map[string]string{
			"user": "aboba",
		},
		errlog.Meta(err),
	)
}
