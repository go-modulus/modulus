package mtools

import (
	"github.com/go-modulus/modulus/cli"
	cli2 "github.com/go-modulus/modulus/internal/mtools/cli"
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/mtools").
		AddCliCommands(
			cli2.NewInitProjectCommand,
			cli2.NewAddModuleCommand,
		).
		AddProviders(
			cli2.NewInitProject,
			cli2.NewAddModule,
		).
		AddDependencies(
			*logger.NewModule(logger.ModuleConfig{}),
			*cli.NewModule(),
		)
}
