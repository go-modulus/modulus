package logger

import (
	"github.com/go-modulus/modulus/module"
)

type ModuleConfig struct {
	Level string `env:"LOGGER_LEVEL, default=debug"`
	Type  string `env:"LOGGER_TYPE, default=console"`
	App   string `env:"LOGGER_APP, default=modulus"`
}

func NewModule() *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/logger").
		AddProviders(
			NewLogger,
			NewSlog,
		).InitConfig(
		ModuleConfig{},
	)
}
