package http

import (
	"net/http"

	httpinIntegration "github.com/ggicci/httpin/integration"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/go-modulus/modulus/http/middleware"
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
)

var (
	ErrMethodNotAllowed = errhttp.ErrWithHttpCode(
		erruser.New("method not allowed", "HTTP method is not allowed"),
		http.StatusMethodNotAllowed,
	)
	ErrNotFound = errhttp.ErrWithHttpCode(
		erruser.New("not found", "Not found"),
		http.StatusNotFound,
	)
)

func NewModule(options ...module.Option) *module.Module {
	httpModule := module.NewModule("http").
		AddDependencies(
			logger.NewModule(),
		).
		AddCliCommands(
			NewServeCommand,
		).
		AddProviders(
			NewServe,
		).
		SetOverriddenProvider("http.Router", NewDefaultRouter).
		SetOverriddenProvider("http.ErrorPipeline", errhttp.NewDefaultErrorPipeline).
		SetOverriddenProvider(
			"http.MiddlewarePipeline", NewDefaultPipeline,
		).
		InitConfig(ServeConfig{}).
		InitConfig(middleware.CorsConfig{}).
		InitConfig(errhttp.ErrorLoggerConfig{}).
		WithOptions(options...)

	return httpModule
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

func NewManifesto() module.Manifesto {
	httpModule := module.NewManifesto(
		NewModule(),
		"github.com/go-modulus/modulus/http",
		"Base package for http server. It is based on the net/http server that is able to work without any dependencies, but the main purpose of this package is working together with another router like Chi provided in the separate module.",
		"1.0.0",
	)

	return httpModule
}

func init() {
	httpinIntegration.UseHttpPathVariable("path")
}
