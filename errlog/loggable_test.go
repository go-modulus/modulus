package errlog_test

import (
	"errors"
	"github.com/go-modulus/modulus/errlog"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoggable(t *testing.T) {
	err := errlog.Wrap(errors.New("test error"), "user", "aboba")
	err = errlog.WrapLoggable(err, true)
	assert.Equal(
		t,
		map[string]string{
			"user": "aboba",
		},
		errlog.Meta(err),
	)
	assert.Equal(
		t,
		true,
		errlog.IsLoggable(err),
	)
}
