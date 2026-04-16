package errlog_test

import (
	"errors"
	"testing"

	errlog2 "github.com/go-modulus/modulus/errors/errlog"
	"github.com/stretchr/testify/assert"
)

func TestLoggable(t *testing.T) {
	t.Parallel()
	err := errlog2.Wrap(errors.New("test error"), "user", "aboba")
	err = errlog2.WrapLoggable(err, true)
	assert.Equal(
		t,
		map[string]string{
			"user": "aboba",
		},
		errlog2.Meta(err),
	)
	assert.Equal(
		t,
		true,
		errlog2.IsLoggable(err),
	)
}
