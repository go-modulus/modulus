package middleware

import (
	"github.com/go-modulus/modulus/http/context"
	"net/http"
)

func Request(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithRequest(r.Context(), r)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
