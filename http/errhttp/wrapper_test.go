package errhttp

import (
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errsys"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendError(t *testing.T) {
	t.Parallel()
	t.Run(
		"send error", func(t *testing.T) {
			t.Parallel()
			err := errors.WithMeta(
				errsys.New("test error", "test error message"),
				"test-key",
				"test-value",
			)

			rr := httptest.NewRecorder()
			SendError(rr, err)

			content := rr.Body.String()

			t.Log("When try to send error")
			t.Log("	Then the error is sent")
			require.Equal(t, http.StatusInternalServerError, rr.Code)
			require.Contains(t, content, `"message":"test error message"`)
			require.Contains(t, content, `"test-key":"test-value"`)
			require.Contains(t, content, `"code":"test error"`)
		},
	)
}
