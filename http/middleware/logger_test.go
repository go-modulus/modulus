package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-modulus/modulus/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newJSONLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run(
		"logs method and path", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			handler := middleware.NewLogger(newJSONLogger(&buf))(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					},
				),
			)

			req := httptest.NewRequest(http.MethodGet, "/hello", nil)
			handler.ServeHTTP(httptest.NewRecorder(), req)

			var entry map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
			assert.Equal(t, "GET", entry["method"])
			assert.Equal(t, "/hello", entry["path"])
		},
	)

	t.Run(
		"logs response status", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			handler := middleware.NewLogger(newJSONLogger(&buf))(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusCreated)
					},
				),
			)

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))

			var entry map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
			assert.Equal(t, float64(http.StatusCreated), entry["status"])
		},
	)

	t.Run(
		"logs bytes written", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			handler := middleware.NewLogger(newJSONLogger(&buf))(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("hello"))
					},
				),
			)

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			var entry map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
			assert.Equal(t, float64(5), entry["bytes"])
		},
	)

	t.Run(
		"logs duration", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			handler := middleware.NewLogger(newJSONLogger(&buf))(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					},
				),
			)

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			var entry map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
			assert.NotEmpty(t, entry["duration"])
		},
	)

	t.Run(
		"log message is 'handled request'", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			handler := middleware.NewLogger(newJSONLogger(&buf))(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					},
				),
			)

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			var entry map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
			assert.Equal(t, "handled request", entry["msg"])
		},
	)

	t.Run(
		"calls next handler", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			nextCalled := false
			handler := middleware.NewLogger(newJSONLogger(&buf))(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						nextCalled = true
						w.WriteHeader(http.StatusOK)
					},
				),
			)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

			require.True(t, nextCalled)
			assert.Equal(t, http.StatusOK, rr.Code)
		},
	)
}
