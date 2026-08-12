package http

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-modulus/modulus/http/errhttp"
)

type Router interface {
	http.Handler

	// Use appends one or more middlewares onto the Router stack.
	Use(middlewares ...func(http.Handler) http.Handler)

	// Method adds routes for `pattern` that matches
	// the `method` HTTP method.
	Method(method, pattern string, h http.Handler)
}

type DefaultRouter struct {
	mux              *http.ServeMux
	middlewares      []func(http.Handler) http.Handler
	notFoundHandler  http.Handler
	methodNotAllowed http.Handler
}

func (r *DefaultRouter) Use(middlewares ...func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *DefaultRouter) Method(method, pattern string, h http.Handler) {
	r.mux.Handle(method+" "+pattern, h)
}

func (r *DefaultRouter) NotFound(h http.Handler) {
	r.notFoundHandler = h
}

func (r *DefaultRouter) MethodNotAllowed(h http.Handler) {
	r.methodNotAllowed = h
}

func (r *DefaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var handler http.Handler = http.HandlerFunc(r.route)
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}
	handler.ServeHTTP(w, req)
}

func (r *DefaultRouter) route(w http.ResponseWriter, req *http.Request) {
	buf := &responseBuffer{headers: make(http.Header), code: http.StatusOK, underlying: w}
	r.mux.ServeHTTP(buf, req)

	if buf.hijacked {
		// The connection was taken over (e.g. a WebSocket upgrade) and must not
		// be touched through w again.
		return
	}

	switch buf.code {
	case http.StatusNotFound:
		if r.notFoundHandler != nil {
			r.notFoundHandler.ServeHTTP(w, req)
			return
		}
	case http.StatusMethodNotAllowed:
		if r.methodNotAllowed != nil {
			if allow := buf.headers.Get("Allow"); allow != "" {
				w.Header().Set("Allow", allow)
			}
			r.methodNotAllowed.ServeHTTP(w, req)
			return
		}
	}
	buf.flush(w)
}

// responseBuffer captures the mux response so custom not-found and
// method-not-allowed handlers can be invoked before writing to the real writer.
type responseBuffer struct {
	headers    http.Header
	code       int
	body       []byte
	underlying http.ResponseWriter
	hijacked   bool
}

func (rb *responseBuffer) Header() http.Header  { return rb.headers }
func (rb *responseBuffer) WriteHeader(code int) { rb.code = code }
func (rb *responseBuffer) Write(b []byte) (int, error) {
	rb.body = append(rb.body, b...)
	return len(b), nil
}

// Hijack lets protocol upgrades (e.g. WebSocket) take over the underlying
// connection. Without this, handlers see a buffered http.ResponseWriter that
// isn't an http.Hijacker and the upgrade fails. Any status/headers already
// recorded on the buffer (e.g. the 101 Switching Protocols response written
// by the upgrader right before hijacking) are flushed to the real writer
// first, so they actually reach the connection before it's taken over.
func (rb *responseBuffer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := hijacker(rb.underlying)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	rb.flush(rb.underlying)
	rb.hijacked = true
	return hj.Hijack()
}

// responseWriterUnwrapper is the de facto standard convention (also used by
// net/http.ResponseController and coder/websocket) for a http.ResponseWriter
// wrapper to expose the writer it wraps.
type responseWriterUnwrapper interface {
	Unwrap() http.ResponseWriter
}

// hijacker walks a chain of wrapped http.ResponseWriters (e.g. middlewares
// that embed http.ResponseWriter and implement Unwrap) to find one that
// actually supports hijacking. A wrapper only needs to implement Unwrap for
// this to see through it; it doesn't need to implement http.Hijacker itself.
func hijacker(w http.ResponseWriter) (http.Hijacker, bool) {
	for {
		switch t := w.(type) {
		case http.Hijacker:
			return t, true
		case responseWriterUnwrapper:
			w = t.Unwrap()
		default:
			return nil, false
		}
	}
}

func (rb *responseBuffer) flush(w http.ResponseWriter) {
	for k, v := range rb.headers {
		w.Header()[k] = v
	}
	if rb.code != http.StatusOK {
		w.WriteHeader(rb.code)
	}
	if len(rb.body) > 0 {
		_, _ = w.Write(rb.body)
	}
}

func NewDefaultRouter(errorPipeline *errhttp.ErrorPipeline, config ServeConfig) Router {
	r := &DefaultRouter{
		mux: http.NewServeMux(),
	}
	r.MethodNotAllowed(
		errhttp.WrapHandler(
			errorPipeline,
			func(w http.ResponseWriter, req *http.Request) error {
				return ErrMethodNotAllowed
			},
		),
	)
	r.NotFound(
		errhttp.WrapHandler(
			errorPipeline,
			func(w http.ResponseWriter, req *http.Request) error {
				return ErrNotFound
			},
		),
	)
	if config.TTL > 0 || config.StreamTTL > 0 {
		r.Use(timeout(config.TTL, config.StreamTTL))
	}
	if config.RequestSizeLimit > 0 {
		r.Use(requestSize(int64(config.RequestSizeLimit.Bytes())))
	}
	return r
}

// timeout bounds the request context lifetime. Regular requests get ttl.
// Long-lived requests (WebSocket upgrades and SSE subscriptions) get
// streamTTL instead, since ttl is normally far too short to hold a
// subscription open. Either duration being <= 0 means "no deadline" for
// that class of request.
func timeout(ttl time.Duration, streamTTL time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			d := ttl
			if isStreamingRequest(r) {
				d = streamTTL
			}
			if d <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					w.WriteHeader(http.StatusGatewayTimeout)
				}
			}()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// isStreamingRequest reports whether the request is a WebSocket upgrade or an
// SSE subscription, mirroring how gqlgen's own transports (transport.Websocket
// and transport.SSE) detect them.
func isStreamingRequest(r *http.Request) bool {
	if r.Header.Get("Upgrade") != "" {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "text/event-stream")
}

func requestSize(bytes int64) func(http.Handler) http.Handler {
	f := func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, bytes)
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	return f
}
