package otel

import (
	"log/slog"
	"net/http"

	"github.com/go-modulus/modulus/logger"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/log"
)

func NewMiddleware(operation string, opts ...otelhttp.Option) func(http.Handler) http.Handler {
	baseMiddleware := otelhttp.NewMiddleware(operation, opts...)
	return func(next http.Handler) http.Handler {
		return baseMiddleware(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					labeler, ok := otelhttp.LabelerFromContext(r.Context())
					if ok {
						labeler.Add(attribute.String("http.route", r.URL.Path))
					}
					next.ServeHTTP(w, r)
				},
			),
		)
	}
}

type LogMiddlewareFactory struct {
	config   Config
	provider *log.LoggerProvider
}

func NewLogMiddlewareFactory(config Config, provider *log.LoggerProvider) *LogMiddlewareFactory {
	return &LogMiddlewareFactory{
		config:   config,
		provider: provider,
	}
}

func (f *LogMiddlewareFactory) SlogMiddleware() logger.Middleware {
	if !f.config.LogViaOtel {
		return func(next slog.Handler) slog.Handler {
			return next
		}
	}
	return NewLogMiddleware(f.config.ServiceName, f.provider)
}

// NewLogMiddleware builds a logger.Middleware that fans every log record out
// to an otelslog handler backed by the given LoggerProvider, in addition to
// whatever handler is next in the modulus logger pipeline.
func NewLogMiddleware(name string, provider *log.LoggerProvider, opts ...otelslog.Option) logger.Middleware {
	otelHandler := otelslog.NewHandler(
		name,
		append([]otelslog.Option{otelslog.WithLoggerProvider(provider)}, opts...)...,
	)

	return func(next slog.Handler) slog.Handler {
		return slogmulti.Fanout(next, otelHandler)
	}
}
