package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/go-modulus/modulus/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEmptyPipeline() *errhttp.ErrorPipeline {
	return &errhttp.ErrorPipeline{}
}

func TestNewBodySeeker(t *testing.T) {
	t.Parallel()

	t.Run(
		"passes through nil body unchanged", func(t *testing.T) {
			t.Parallel()

			var receivedBody io.ReadCloser
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedBody = r.Body
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Body = nil // httptest.NewRequest sets http.NoBody; force truly nil
			rr := httptest.NewRecorder()

			middleware.NewBodySeeker(newEmptyPipeline())(next).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			assert.Nil(t, receivedBody)
		},
	)

	t.Run(
		"replaces body with seekable RequestBody", func(t *testing.T) {
			t.Parallel()

			var receivedBody io.ReadCloser
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedBody = r.Body
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
			rr := httptest.NewRecorder()

			middleware.NewBodySeeker(newEmptyPipeline())(next).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			_, ok := receivedBody.(io.ReadSeeker)
			assert.True(t, ok, "body should implement io.ReadSeeker")
		},
	)

	t.Run(
		"body can be read multiple times after seeking", func(t *testing.T) {
			t.Parallel()

			const payload = "reusable body content"
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seeker, ok := r.Body.(io.ReadSeeker)
				require.True(t, ok)

				first, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, payload, string(first))

				_, err = seeker.Seek(0, io.SeekStart)
				require.NoError(t, err)

				second, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, payload, string(second))

				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
			rr := httptest.NewRecorder()

			middleware.NewBodySeeker(newEmptyPipeline())(next).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
		},
	)

	t.Run(
		"close on RequestBody is a no-op", func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := r.Body.Close()
				require.NoError(t, err)

				// body is still readable after close
				data, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Equal(t, "data", string(data))

				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("data"))
			rr := httptest.NewRecorder()

			middleware.NewBodySeeker(newEmptyPipeline())(next).ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
		},
	)

	t.Run(
		"calls next handler after buffering body", func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("body"))
			rr := httptest.NewRecorder()

			middleware.NewBodySeeker(newEmptyPipeline())(next).ServeHTTP(rr, req)

			assert.True(t, nextCalled)
			assert.Equal(t, http.StatusNoContent, rr.Code)
		},
	)

	t.Run(
		"error reading body sends error response and does not call next", func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/", &errorReader{})
			rr := httptest.NewRecorder()

			middleware.NewBodySeeker(newEmptyPipeline())(next).ServeHTTP(rr, req)

			assert.False(t, nextCalled)
			assert.Equal(t, http.StatusInternalServerError, rr.Code)
		},
	)
}

// errorReader always returns an error on Read to simulate a broken request body.
type errorReader struct{}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
