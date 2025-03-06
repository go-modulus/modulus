package http

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/errors/errwrap"
	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/go-modulus/modulus/module"
	"net/http"
)

var (
	ErrMethodNotAllowed = errwrap.Wrap(
		erruser.New("MethodNotAllowed", "Method not allowed"),
		errhttp.With(http.StatusMethodNotAllowed),
	)
	ErrNotFound = errwrap.Wrap(
		erruser.New("NotFound", "Not found"),
		errhttp.With(http.StatusNotFound),
	)
)

func NewRouter(errorPipeline *errhttp.ErrorPipeline, config ServeConfig) chi.Router {
	r := chi.NewRouter()
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
		).
		SetOverriddenProvider("http.ErrorPipeline", errhttp.NewDefaultErrorPipeline).
		SetOverriddenProvider(
			"http.MiddlewarePipeline", func(authMd *auth.Middleware) *Pipeline {
				return &Pipeline{
					Middlewares: []Middleware{},
				}
			},
		).
		InitConfig(ServeConfig{}).
		InitConfig(errhttp.ErrorLoggerConfig{})
}

func OverrideErrorPipeline(httpModule *module.Module, pipeline interface{}) *module.Module {
	return httpModule.SetOverriddenProvider("http.ErrorPipeline", pipeline)
}

func OverrideMiddlewarePipeline(httpModule *module.Module, pipeline interface{}) *module.Module {
	return httpModule.SetOverriddenProvider("http.MiddlewarePipeline", pipeline)
}
