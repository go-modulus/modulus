package logger

import (
	"github.com/go-modulus/modulus/module"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

type ModuleConfig struct {
	Level        string `env:"LOGGER_LEVEL, default=debug" comment:"Use one of \"debug\", \"info\", \"warn\", \"error\". It sets the minimum level of the log messages that should be logged"`
	Type         string `env:"LOGGER_TYPE, default=console" comment:"Use either \"console\" or \"json\" value"`
	App          string `env:"LOGGER_APP, default=modulus"`
	FxEventLevel string `env:"LOGGER_FX_EVENT_LEVEL, default=info" comment:"Use one of \"debug\", \"info\", \"warn\", \"error\". It sets the minimum level of the fx events that should be logged"`
}

func NewModule(options ...module.Option) *module.Module {
	return module.NewModule("logger").
		AddProviders(
			NewLogger,
			NewSlog,
		).
		SetOverriddenProvider("logger.MiddlewarePipeline", NewDefaultPipeline).
		InitConfig(ModuleConfig{}).
		WithOptions(options...)
}

func SetConfig(config ModuleConfig) module.Option {
	return func(m *module.Module) *module.Module {
		return m.InitConfig(config)
	}
}

// OverrideMiddlewarePipeline lets a module replace the default, ranked set of
// logger middlewares with a custom one.
func OverrideMiddlewarePipeline[T PipelineFactory](loggerModule *module.Module) *module.Module {
	return loggerModule.SetOverriddenProvider("logger.MiddlewarePipeline", func(impl T) *Pipeline { return impl.New() })
}

// AddMiddlewareToPipeline registers an additional logger middleware at the
// given rank. Middlewares with a lower rank run first, closer to the log
// call site.
func AddMiddlewareToPipeline(rank int, addedMiddleware Middleware) module.Option {
	return func(loggerModule *module.Module) *module.Module {
		return loggerModule.AddInvokes(
			func(pipeline *Pipeline) *Pipeline {
				pipeline.SetMiddleware(rank, addedMiddleware)
				return pipeline
			},
		)
	}
}

// AddMiddlewareFactoryToPipeline registers an additional logger middleware,
// built by the given MiddlewareFactory, at the given rank. Middlewares with
// a lower rank run first, closer to the log call site.
func AddMiddlewareFactoryToPipeline[T MiddlewareFactory](rank int) module.Option {
	return func(loggerModule *module.Module) *module.Module {
		return loggerModule.AddInvokes(
			func(pipeline *Pipeline, factory T) *Pipeline {
				pipeline.SetMiddleware(rank, factory.SlogMiddleware())
				return pipeline
			},
		)
	}
}

func NewManifesto() module.Manifesto {
	return module.NewManifesto(
		NewModule(),
		"github.com/go-modulus/modulus/logger",
		"Slog logger with a zap backend for the Modulus framework.",
		"1.0.0",
	)
}

func FxLoggerOption() fx.Option {
	loggerOption := fx.WithLogger(
		func(logger *zap.Logger, config ModuleConfig) fxevent.Logger {
			level, err := zap.ParseAtomicLevel(config.FxEventLevel)
			if err != nil {
				level = zap.NewAtomicLevel()
				level.SetLevel(zap.InfoLevel)
			}
			logger = logger.WithOptions(zap.IncreaseLevel(level))

			return &fxevent.ZapLogger{Logger: logger}
		},
	)

	return loggerOption
}
