package http

import (
	"context"
	"net/http"
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
	buf := &responseBuffer{headers: make(http.Header), code: http.StatusOK}
	r.mux.ServeHTTP(buf, req)

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
	headers http.Header
	code    int
	body    []byte
}

func (rb *responseBuffer) Header() http.Header  { return rb.headers }
func (rb *responseBuffer) WriteHeader(code int) { rb.code = code }
func (rb *responseBuffer) Write(b []byte) (int, error) {
	rb.body = append(rb.body, b...)
	return len(b), nil
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
	if config.TTL > 0 {
		r.Use(timeout(config.TTL))
	}
	if config.RequestSizeLimit > 0 {
		r.Use(requestSize(int64(config.RequestSizeLimit.Bytes())))
	}
	return r
}

func timeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
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
