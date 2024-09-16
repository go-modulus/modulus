package logger

import (
	"github.com/go-modulus/modulus/module"
)

type ModuleConfig struct {
	Level string `env:"LOGGER_LEVEL, default=debug"`
	Type  string `env:"LOGGER_TYPE, default=json"`
	App   string `env:"LOGGER_APP, default=modulus"`
}

func NewModule(config ModuleConfig) *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/logger").
		AddProviders(
			NewLogger,
			NewSlog,
			module.ConfigProvider[ModuleConfig](config),
		)
}
