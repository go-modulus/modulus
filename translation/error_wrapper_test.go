package translation

import (
	"errors"
	"fmt"
	errors2 "github.com/go-modulus/modulus/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHintArguments(t *testing.T) {
	err := errors.New("some error")
	err = WithDomainHint(err, "arg1", "string: %s, int: %d, bool: %t", "test", 123, true)

	args := HintArguments(err)
	hint := errors2.Hint(err)

	formatedString := fmt.Sprintf(hint, args...)
	expectedString := fmt.Sprintf("string: %s, int: %d, bool: %t", "test", 123, true)
	require.Len(t, args, 3)
	require.Equal(t, expectedString, formatedString)
	require.Equal(t, "test", args[0])
	require.Equal(t, int64(123), args[1])
	require.Equal(t, true, args[2])
}
