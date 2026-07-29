package errpipe

import (
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
)

func NewModule(options ...module.Option) *module.Module {
	httpModule := module.NewModule("errors/errpipe").
		AddDependencies(
			logger.NewModule(),
		).
		AddProviders().
		SetOverriddenProvider("errpipe.ErrorPipeline", NewDefaultErrorPipeline).
		InitConfig(ErrorLoggerConfig{}).
		WithOptions(options...)

	return httpModule
}

// OverrideErrorPipeline - overrides a flow or error transformers to make errors more secure and user friendly
func OverrideErrorPipeline[T ErrorPipelineFactory](httpModule *module.Module) *module.Module {
	return httpModule.SetOverriddenProvider(
		"http.ErrorPipeline",
		func(impl T) *ErrorPipeline { return impl.New() },
	)
}

func NewManifesto() module.Manifesto {
	httpModule := module.NewManifesto(
		NewModule(),
		"github.com/go-modulus/modulus/errors/errpipe",
		"Error pipeline module to transform errors to the user-friendly representation.",
		"1.0.0",
	)
	httpModule.Install.EnvVars = []module.ConfigEnvVariable{
		{
			Key:     "USER_ERROR_LOG_LEVEL",
			Value:   "dont_log",
			Comment: "Log user errors or not.: dont_log, error, warn, info, debug",
		},
		{
			Key:     "SYSTEM_ERROR_LOG",
			Value:   "error",
			Comment: "Log system errors or not.: dont_log, error, warn, info, debug",
		},
	}

	return httpModule
}
