package middleware

import (
	"net/http"
	"regexp"

	"github.com/rs/cors"
)

type CorsConfig struct {
	Host string `env:"CORS_HOST, default=^https?://(localhost|127.0.0.1)(:[0-9]+)?$"`
}

func NewCors(config CorsConfig) *cors.Cors {
	host := config.Host
	if host == "*" {
		host = ".+"
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
			AllowedHeaders: []string{
				"accept",
				"Accept-Encoding",
				"Accept-Language",
				"Authorization",
				"authorization",
				"Content-Type",
				"Content-Length",
				"Cache-Control",
				"Connection",
				"Pragma",
				"Cookie",
				"Access-Control-Allow-Origin",
				"User-Agent",
			},
			ExposedHeaders:     nil,
			MaxAge:             3600,
			AllowCredentials:   true,
			OptionsPassthrough: false,
			Debug:              false,
		},
	)
}
