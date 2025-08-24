package translation

import (
	"context"
	"errors"
	"testing"

	errors2 "github.com/go-modulus/modulus/errors"
	"github.com/stretchr/testify/require"
)

func TestLocalizeErrorHint(t *testing.T) {
	err := errors.New("some error")
	err = WithDomainHint(err, "arg1", "string: %s, int: %d, bool: %t", "test", 123, true)
	err = LocalizeErrorHint()(context.Background(), err)

	require.NotNil(t, err)
	hint := errors2.Hint(err)
	expectedHint := "string: test, int: 123, bool: true"
	require.Equal(t, expectedHint, hint)
}
