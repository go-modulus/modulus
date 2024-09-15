package middleware

import (
	"github.com/go-modulus/modulus/http/context"
	"github.com/go-modulus/modulus/logger"
	"github.com/rs/xid"
	"net/http"
)

const (
	RequestIDHeader = "X-Request-Id"
)

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		requestID := xid.New().String()
		ctx := context.WithRequestID(r.Context(), requestID)
		ctx = logger.AddTags(ctx, "requestId", requestID)
		w.Header().Add(RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
