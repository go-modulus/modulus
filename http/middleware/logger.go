package middleware

import (
	"bufio"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the
// status code and bytes written. It also implements http.Flusher and http.Hijacker
// if the underlying ResponseWriter supports them.
type responseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(status int) {
	if rw.wroteHeader {
		return
	}
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
	rw.wroteHeader = true
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// Flush implements the http.Flusher interface.
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements the http.Hijacker interface.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func NewLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			}
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			start := time.Now()
			defer func() {
				attrs = append(
					attrs,
					slog.Int("status", ww.status),
					slog.Int("bytes", ww.bytes),
					slog.String("duration", time.Since(start).String()),
				)

				logger.LogAttrs(
					r.Context(),
					slog.LevelInfo,
					"handled request",
					attrs...,
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
