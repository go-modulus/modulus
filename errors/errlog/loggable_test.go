package errlog_test

import (
	"errors"
	errlog2 "github.com/go-modulus/modulus/errors/errlog"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoggable(t *testing.T) {
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
