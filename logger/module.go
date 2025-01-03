package logger

import (
	"github.com/go-modulus/modulus/module"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

type ModuleConfig struct {
	Level        string `env:"LOGGER_LEVEL, default=debug" comment:"Use one of \"debug\", \"info\", \"warn\", \"error\". It sets the maximum level of the log messages that should be logged"`
	Type         string `env:"LOGGER_TYPE, default=console" comment:"Use either \"console\" or \"json\" value"`
	App          string `env:"LOGGER_APP, default=modulus"`
	FxEventLevel string `env:"LOGGER_FX_EVENT_LEVEL, default=info" comment:"Use one of \"debug\", \"info\", \"warn\", \"error\". It sets the maximum level of the fx events that should be logged"`
}

func NewModule() *module.Module {
	return module.NewModule("slog logger").
		AddProviders(
			NewLogger,
			NewSlog,
		).InitConfig(
		ModuleConfig{},
	)
}

func WithLoggerOption() fx.Option {
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
