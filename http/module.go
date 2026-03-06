package http

import (
	"net/http"

	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/errors/errwrap"
	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/go-modulus/modulus/http/middleware"
	"github.com/go-modulus/modulus/module"
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

func NewModule(options ...module.Option) *module.Module {
	module := module.NewModule("http").
		AddCliCommands(
			NewServeCommand,
		).
		AddProviders(
			NewServe,
		).
		SetOverriddenProvider("http.Router", NewDefaultRouter).
		SetOverriddenProvider("http.ErrorPipeline", errhttp.NewDefaultErrorPipeline).
		SetOverriddenProvider(
			"http.MiddlewarePipeline", NewDefaultPipeline(),
		).
		InitConfig(ServeConfig{}).
		InitConfig(middleware.CorsConfig{}).
		InitConfig(errhttp.ErrorLoggerConfig{}).
		WithOptions(options...)

	return module
}

func OverrideRouter[T Router](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider("http.Router", func(impl T) Router { return impl })
}

func OverrideErrorPipeline[T errhttp.ErrorPipelineFactory](httpModule *module.Module) *module.Module {
	return httpModule.SetOverriddenProvider(
		"http.ErrorPipeline",
		func(impl T) *errhttp.ErrorPipeline { return impl.New() },
	)
}

func OverrideMiddlewarePipeline[T PipelineFactory](httpModule *module.Module) *module.Module {
	return httpModule.SetOverriddenProvider("http.MiddlewarePipeline", func(impl T) *Pipeline { return impl.New() })
}
