package middleware

import (
	"context"
	"github.com/go-modulus/modulus/logger"
	"net/http"
)

type ctxKeyUserAgent string

const UserAgentKey ctxKeyUserAgent = "userAgent"

func GetUserAgent(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userAgent, ok := ctx.Value(UserAgentKey).(string); ok {
		return userAgent
	}
	return ""
}

func UserAgent(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.UserAgent()
		ctx := context.WithValue(r.Context(), UserAgentKey, userAgent)
		ctx = logger.AddTags(ctx, "userAgent", userAgent)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
