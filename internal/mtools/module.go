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
			cli2.NewCreateModuleCommand,
		).
		AddProviders(
			cli2.NewInitProject,
			cli2.NewAddModule,
			cli2.NewCreateModule,
		).
		AddDependencies(
			*logger.NewModule(logger.ModuleConfig{}),
			*cli.NewModule(cli.ModuleConfig{}),
		)
}
