package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpContext "github.com/go-modulus/modulus/http/context"
	"github.com/go-modulus/modulus/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	t.Run(
		"sets X-Request-Id response header", func(t *testing.T) {
			t.Parallel()

			handler := middleware.RequestID(okHandler)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

			require.Equal(t, http.StatusOK, rr.Code)
			assert.NotEmpty(t, rr.Header().Get(middleware.RequestIDHeader))
		},
	)

	t.Run(
		"stores request ID in context", func(t *testing.T) {
			t.Parallel()

			var capturedID string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedID = httpContext.GetRequestID(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.RequestID(next)
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			assert.NotEmpty(t, capturedID)
		},
	)

	t.Run(
		"context ID matches response header", func(t *testing.T) {
			t.Parallel()

			var capturedID string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedID = httpContext.GetRequestID(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.RequestID(next)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

			assert.Equal(t, capturedID, rr.Header().Get(middleware.RequestIDHeader))
		},
	)

	t.Run(
		"generates unique ID per request", func(t *testing.T) {
			t.Parallel()

			handler := middleware.RequestID(okHandler)

			rr1 := httptest.NewRecorder()
			handler.ServeHTTP(rr1, httptest.NewRequest(http.MethodGet, "/", nil))

			rr2 := httptest.NewRecorder()
			handler.ServeHTTP(rr2, httptest.NewRequest(http.MethodGet, "/", nil))

			id1 := rr1.Header().Get(middleware.RequestIDHeader)
			id2 := rr2.Header().Get(middleware.RequestIDHeader)
			assert.NotEmpty(t, id1)
			assert.NotEmpty(t, id2)
			assert.NotEqual(t, id1, id2)
		},
	)
}
