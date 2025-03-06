package middleware

import (
	"braces.dev/errtrace"
	"bytes"
	"github.com/go-modulus/modulus/http/errhttp"
	"io"
	"log/slog"
	"net/http"
)

type RequestBody struct {
	*bytes.Reader
}

func (RequestBody) Close() error { return nil }

// BodySeeker is a middleware that reads the request body and replaces it with a new RequestBody
// that implements the io.Seeker interface to read body in handlers multiple times.
func NewBodySeeker(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				data, err := io.ReadAll(r.Body)
				if err != nil {
					errhttp.SendError(logger, w, r, errtrace.Wrap(err))
				}
				r.Body = RequestBody{bytes.NewReader(data)}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
