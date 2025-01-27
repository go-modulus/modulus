package http

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	errhttp2 "github.com/go-modulus/modulus/errors/errhttp"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/errors/errwrap"
	"github.com/go-modulus/modulus/module"
	"log/slog"
	"net/http"
)

var (
	ErrMethodNotAllowed = errwrap.Wrap(
		erruser.New("MethodNotAllowed", "Method not allowed"),
		errhttp2.With(http.StatusMethodNotAllowed),
	)
	ErrNotFound = errwrap.Wrap(
		erruser.New("NotFound", "Not found"),
		errhttp2.With(http.StatusNotFound),
	)
)

func NewRouter(logger *slog.Logger, config ServeConfig) chi.Router {
	r := chi.NewRouter()
	r.MethodNotAllowed(
		errhttp2.WrapHandler(
			logger,
			func(w http.ResponseWriter, req *http.Request) error {
				return ErrMethodNotAllowed
			},
		),
	)
	r.NotFound(
		errhttp2.WrapHandler(
			logger,
			func(w http.ResponseWriter, req *http.Request) error {
				return ErrNotFound
			},
		),
	)
	if config.TTL > 0 {
		r.Use(chiMiddleware.Timeout(config.TTL))
	}
	if config.RequestSizeLimit > 0 {
		r.Use(chiMiddleware.RequestSize(int64(config.RequestSizeLimit.Bytes())))
	}
	return r
}

func NewModule() *module.Module {
	return module.NewModule("chi http").
		AddCliCommands(
			NewServeCommand,
		).
		AddProviders(
			NewRouter,
			NewServe,
		).InitConfig(ServeConfig{})
}
