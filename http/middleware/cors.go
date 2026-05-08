package middleware

import (
	"net/http"
	"regexp"

	"github.com/rs/cors"
)

type CorsConfig struct {
	Host           string   `env:"CORS_HOST, default=^https?://(localhost|127.0.0.1)(:[0-9]+)?$"`
	AllowedMethods []string `env:"CORS_ALLOWED_METHODS"`
	AllowedHeaders []string `env:"CORS_ALLOWED_HEADERS"`
}

func NewCors(config CorsConfig) *cors.Cors {
	host := config.Host
	if host == "*" {
		host = ".+"
	}
	corsRegexp := regexp.MustCompile(host)

	allowedMethods := config.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodOptions,
			http.MethodDelete,
		}
	}

	allowedHeaders := config.AllowedHeaders
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{
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
		}
	}

	return cors.New(
		cors.Options{
			AllowOriginFunc:    func(origin string) bool { return corsRegexp.Match([]byte(origin)) },
			AllowedMethods:     allowedMethods,
			AllowedHeaders:     allowedHeaders,
			ExposedHeaders:     nil,
			MaxAge:             3600,
			AllowCredentials:   true,
			OptionsPassthrough: false,
			Debug:              false,
		},
	)
}
