package otel

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

func NewMiddleware(operation string, opts ...otelhttp.Option) func(http.Handler) http.Handler {
	baseMiddleware := otelhttp.NewMiddleware(operation, opts...)
	return func(next http.Handler) http.Handler {
		return baseMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// явно проставляем маршрут для otelhttp
			labeler, ok := otelhttp.LabelerFromContext(r.Context())
			if ok {
				labeler.Add(attribute.String("http.route", r.URL.Path))
			}
			next.ServeHTTP(w, r)
		}))
	}
}
