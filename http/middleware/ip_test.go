package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-modulus/modulus/http/middleware"
	"github.com/stretchr/testify/require"
)

func TestIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expectedIP string
	}{
		{
			name:       "uses do-connecting-ip header when present",
			remoteAddr: "192.168.1.1:8080",
			headers: map[string]string{
				"do-connecting-ip": "203.0.113.1",
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:       "uses X-Forwarded-For header when do-connecting-ip not present",
			remoteAddr: "192.168.1.1:8080",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.2",
			},
			expectedIP: "203.0.113.2",
		},
		{
			name:       "uses first IP from X-Forwarded-For when multiple IPs",
			remoteAddr: "192.168.1.1:8080",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.3, 192.168.1.2, 10.0.0.1",
			},
			expectedIP: "203.0.113.3",
		},
		{
			name:       "falls back to remote address when no headers",
			remoteAddr: "203.0.113.4:8080",
			headers:    map[string]string{},
			expectedIP: "203.0.113.4",
		},
		{
			name:       "uses remote address directly when no port",
			remoteAddr: "203.0.113.5",
			headers:    map[string]string{},
			expectedIP: "203.0.113.5",
		},
		{
			name:       "prefers do-connecting-ip over X-Forwarded-For",
			remoteAddr: "192.168.1.1:8080",
			headers: map[string]string{
				"do-connecting-ip": "203.0.113.6",
				"X-Forwarded-For":  "203.0.113.7",
			},
			expectedIP: "203.0.113.6",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				var capturedIP string
				nextHandler := http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						capturedIP = middleware.GetIP(r.Context())
						w.WriteHeader(http.StatusOK)
					},
				)

				middleware := middleware.IP(nextHandler)

				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = tt.remoteAddr
				for key, value := range tt.headers {
					req.Header.Set(key, value)
				}

				rr := httptest.NewRecorder()
				middleware.ServeHTTP(rr, req)

				require.Equal(t, tt.expectedIP, capturedIP)
				require.Equal(t, http.StatusOK, rr.Code)
			},
		)
	}
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "returns empty string for nil context",
			ctx:      nil,
			expected: "",
		},
		{
			name:     "returns empty string when IP not set in context",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "returns IP when set in context",
			ctx:      context.WithValue(context.Background(), middleware.IPKey, "192.168.1.1"),
			expected: "192.168.1.1",
		},
		{
			name:     "returns empty string when value is not a string",
			ctx:      context.WithValue(context.Background(), middleware.IPKey, 123),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := middleware.GetIP(tt.ctx)
				require.Equal(t, tt.expected, result)
			},
		)
	}
}
