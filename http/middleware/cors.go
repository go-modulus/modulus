package middleware

import (
	"net/http"
	"regexp"

	"github.com/rs/cors"
)

type CorsConfig struct {
	Host                     string   `env:"CORS_HOST, default=^https?://(localhost|127.0.0.1)(:[0-9]+)?$" comment:"Regular expression pattern for allowed origins or * for all origins"`
	AdditionalAllowedHeaders []string `env:"CORS_ADDITIONAL_ALLOWED_HEADERS" comment:"Comma-separated list of additional allowed headers for CORS requests"`
	MaxAge                   int      `env:"CORS_MAX_AGE, default=3600"`
}

func NewCors(config CorsConfig) *cors.Cors {
	host := config.Host
	if host == "*" || host == "" {
		return cors.AllowAll()
	}
	corsRegexp := regexp.MustCompile(host)

	return cors.New(
		cors.Options{
			AllowOriginFunc: func(origin string) bool {
				return corsRegexp.Match([]byte(origin))
			},
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodOptions,
				http.MethodDelete,
			},
			AllowedHeaders: append(
				[]string{
					"Accept",
					"Accept-Encoding",
					"Accept-Language",
					"Authorization",
					"Content-Type",
					"Content-Length",
					"Cache-Control",
					"Connection",
					"Pragma",
					"Cookie",
					"Access-Control-Allow-Origin",
					"User-Agent",
					"Last-Event-Id",
				}, config.AdditionalAllowedHeaders...,
			),
			ExposedHeaders:     nil,
			MaxAge:             config.MaxAge,
			AllowCredentials:   true,
			OptionsPassthrough: false,
			Debug:              false,
		},
	)
}
