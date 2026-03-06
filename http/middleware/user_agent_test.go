package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-modulus/modulus/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserAgent(t *testing.T) {
	t.Parallel()

	t.Run(
		"stores User-Agent header value in context", func(t *testing.T) {
			t.Parallel()

			var capturedUA string
			next := http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					capturedUA = middleware.GetUserAgent(r.Context())
					w.WriteHeader(http.StatusOK)
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("User-Agent", "TestBrowser/1.0")

			middleware.UserAgent(next).ServeHTTP(httptest.NewRecorder(), req)

			assert.Equal(t, "TestBrowser/1.0", capturedUA)
		},
	)

	t.Run(
		"stores empty string when User-Agent header is absent", func(t *testing.T) {
			t.Parallel()

			var capturedUA string
			next := http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					capturedUA = middleware.GetUserAgent(r.Context())
					w.WriteHeader(http.StatusOK)
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			// no User-Agent header set

			middleware.UserAgent(next).ServeHTTP(httptest.NewRecorder(), req)

			assert.Equal(t, "", capturedUA)
		},
	)

	t.Run(
		"calls next handler", func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			next := http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					nextCalled = true
					w.WriteHeader(http.StatusOK)
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			middleware.UserAgent(next).ServeHTTP(rr, req)

			require.True(t, nextCalled)
			assert.Equal(t, http.StatusOK, rr.Code)
		},
	)
}

func TestGetUserAgent(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns empty string for nil context", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "", middleware.GetUserAgent(context.TODO()))
		},
	)

	t.Run(
		"returns empty string when key not set", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "", middleware.GetUserAgent(context.Background()))
		},
	)

	t.Run(
		"returns value when key is set", func(t *testing.T) {
			t.Parallel()
			ctx := context.WithValue(context.Background(), middleware.UserAgentKey, "MyAgent/2.0")
			assert.Equal(t, "MyAgent/2.0", middleware.GetUserAgent(ctx))
		},
	)

	t.Run(
		"returns empty string when value is not a string", func(t *testing.T) {
			t.Parallel()
			ctx := context.WithValue(context.Background(), middleware.UserAgentKey, 42)
			assert.Equal(t, "", middleware.GetUserAgent(ctx))
		},
	)
}
