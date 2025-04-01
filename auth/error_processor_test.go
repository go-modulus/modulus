package auth_test

import (
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertAuthTagsToHttpCode(t *testing.T) {
	t.Parallel()
	t.Run(
		"add unauthenticated code to error", func(t *testing.T) {
			t.Parallel()
			err := auth.AddHttpCode()(context.Background(), auth.ErrUnauthenticated)
			t.Log("When try to add unauthenticated code to error")
			t.Log("	Then the error gets the meta with httpCode 401")
			require.Error(t, err)
			require.Equal(t, 401, errhttp.HttpCode(err))
		},
	)

	t.Run(
		"add unauthorized code to error", func(t *testing.T) {
			t.Parallel()
			err := auth.AddHttpCode()(context.Background(), auth.ErrUnauthorized)
			t.Log("When try to add unauthorized code to error")
			t.Log("	Then the error gets the meta with httpCode 403")
			require.Error(t, err)
			require.Equal(t, 403, errhttp.HttpCode(err))
		},
	)
}
