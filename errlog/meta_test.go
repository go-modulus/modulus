package errlog_test

import (
	"errors"
	"github.com/go-modulus/modulus/errlog"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMeta(t *testing.T) {
	err := errlog.Wrap(errors.New("test error"), "user", "aboba")
	assert.Equal(
		t,
		map[string]string{
			"user": "aboba",
		},
		errlog.Meta(err),
	)
}
