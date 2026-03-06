package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-modulus/modulus/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func corsRequest(t *testing.T, corsMiddleware http.Handler, method, origin string, extraHeaders map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/", nil)
	req.Header.Set("Origin", origin)
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	rr := httptest.NewRecorder()
	corsMiddleware.ServeHTTP(rr, req)
	return rr
}

func TestNewCors(t *testing.T) {
	t.Parallel()

	t.Run(
		"wildcard host allows any origin", func(t *testing.T) {
			t.Parallel()

			c := middleware.NewCors(middleware.CorsConfig{Host: "*"})
			handler := c.Handler(okHandler)

			rr := corsRequest(t, handler, http.MethodGet, "https://any-origin.example.com", nil)

			require.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
		},
	)

	t.Run(
		"empty host allows any origin", func(t *testing.T) {
			t.Parallel()

			c := middleware.NewCors(middleware.CorsConfig{Host: ""})
			handler := c.Handler(okHandler)

			rr := corsRequest(t, handler, http.MethodGet, "https://any-origin.example.com", nil)

			require.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
		},
	)

	t.Run(
		"default pattern allows localhost origins", func(t *testing.T) {
			t.Parallel()

			defaultPattern := `^https?://(localhost|127\.0\.0\.1)(:[0-9]+)?$`
			c := middleware.NewCors(middleware.CorsConfig{Host: defaultPattern})
			handler := c.Handler(okHandler)

			allowedOrigins := []string{
				"http://localhost",
				"https://localhost",
				"http://localhost:3000",
				"https://localhost:8080",
				"http://127.0.0.1",
				"https://127.0.0.1:443",
			}
			for _, origin := range allowedOrigins {
				rr := corsRequest(t, handler, http.MethodGet, origin, nil)
				assert.Equal(t, origin, rr.Header().Get("Access-Control-Allow-Origin"),
					"expected origin %q to be allowed", origin)
			}
		},
	)

	t.Run(
		"default pattern blocks non-localhost origins", func(t *testing.T) {
			t.Parallel()

			defaultPattern := `^https?://(localhost|127\.0\.0\.1)(:[0-9]+)?$`
			c := middleware.NewCors(middleware.CorsConfig{Host: defaultPattern})
			handler := c.Handler(okHandler)

			blockedOrigins := []string{
				"https://evil.com",
				"https://notlocalhost.com",
				"http://192.168.1.1",
			}
			for _, origin := range blockedOrigins {
				rr := corsRequest(t, handler, http.MethodGet, origin, nil)
				assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"),
					"expected origin %q to be blocked", origin)
			}
		},
	)

	t.Run(
		"preflight accepts every configured method", func(t *testing.T) {
			t.Parallel()

			c := middleware.NewCors(middleware.CorsConfig{Host: "https://example.com"})
			handler := c.Handler(okHandler)

			// rs/cors echoes only the requested method; check each method is individually accepted.
			for _, method := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"} {
				rr := corsRequest(t, handler, http.MethodOptions, "https://example.com", map[string]string{
					"Access-Control-Request-Method": method,
				})
				assert.Equal(t, method, rr.Header().Get("Access-Control-Allow-Methods"),
					"method %s should be allowed", method)
			}
		},
	)

	t.Run(
		"preflight sets max age from config", func(t *testing.T) {
			t.Parallel()

			c := middleware.NewCors(middleware.CorsConfig{Host: "https://example.com", MaxAge: 7200})
			handler := c.Handler(okHandler)

			rr := corsRequest(t, handler, http.MethodOptions, "https://example.com", map[string]string{
				"Access-Control-Request-Method": http.MethodGet,
			})

			assert.Equal(t, "7200", rr.Header().Get("Access-Control-Max-Age"))
		},
	)

	t.Run(
		"preflight allows credentials", func(t *testing.T) {
			t.Parallel()

			c := middleware.NewCors(middleware.CorsConfig{Host: "https://example.com"})
			handler := c.Handler(okHandler)

			rr := corsRequest(t, handler, http.MethodOptions, "https://example.com", map[string]string{
				"Access-Control-Request-Method": http.MethodGet,
			})

			assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
		},
	)

	t.Run(
		"preflight accepts additional allowed headers", func(t *testing.T) {
			t.Parallel()

			c := middleware.NewCors(middleware.CorsConfig{
				Host:                     "https://example.com",
				AdditionalAllowedHeaders: []string{"X-Custom-Header", "X-Api-Key"},
			})
			handler := c.Handler(okHandler)

			// rs/cors v1.11 stores allowed headers lowercased and the Accepts check is
			// case-sensitive, so the request must send header names in lowercase (as
			// browsers do per the CORS spec).
			rr := corsRequest(t, handler, http.MethodOptions, "https://example.com", map[string]string{
				"Access-Control-Request-Method":  http.MethodGet,
				"Access-Control-Request-Headers": "x-custom-header",
			})

			assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Origin"),
				"origin should be allowed when requested header is in the allowed list")
			assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "x-custom-header")
		},
	)

	t.Run(
		"preflight accepts standard allowed headers", func(t *testing.T) {
			t.Parallel()

			c := middleware.NewCors(middleware.CorsConfig{Host: "https://example.com"})
			handler := c.Handler(okHandler)

			// Header name must be lowercase — rs/cors normalises allowed headers to
			// lowercase and performs a case-sensitive lookup against the request value.
			rr := corsRequest(t, handler, http.MethodOptions, "https://example.com", map[string]string{
				"Access-Control-Request-Method":  http.MethodPost,
				"Access-Control-Request-Headers": "authorization",
			})

			assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Origin"),
				"origin should be allowed when requested header is in the allowed list")
			assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "authorization")
		},
	)
}
