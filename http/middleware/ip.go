package middleware

import (
	"context"
	"github.com/go-modulus/modulus/logger"
	"net"
	"net/http"
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
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
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
