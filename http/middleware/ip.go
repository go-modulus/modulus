package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/go-modulus/modulus/logger"
)

type ctxKeyIP string

const IPKey ctxKeyIP = "ip"

func GetIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if ip, ok := ctx.Value(IPKey).(string); ok {
		return ip
	}
	return ""
}

func IP(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			if ip == "" {
				ip = r.RemoteAddr
			}
			parsedIp := net.ParseIP(ip)
			ctx := context.WithValue(r.Context(), IPKey, parsedIp.String())
			ctx = logger.AddTags(ctx, "ip", parsedIp.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}

func realIP(r *http.Request) string {
	// Digital Ocean specific header
	if ip := r.Header.Get("do-connecting-ip"); ip != "" {
		return ip
	}
	// Get the real IP from the X-Forwarded-For header.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	// Fall back to the remote address.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
