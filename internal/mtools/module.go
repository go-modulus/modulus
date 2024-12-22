package mtools

import (
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/internal/mtools/action"
	cmdRoot "github.com/go-modulus/modulus/internal/mtools/cli"
	cmdDb "github.com/go-modulus/modulus/internal/mtools/cli/db"
	cmdMmodule "github.com/go-modulus/modulus/internal/mtools/cli/module"
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/mtools").
		AddCliCommands(
			cmdDb.NewDbCommand,
			cmdRoot.NewInitProjectCommand,
			cmdMmodule.NewModuleCommand,
		).
		AddProviders(
			cmdRoot.NewInitProject,
			cmdMmodule.NewInstall,
			cmdMmodule.NewCreate,
			action.NewInstallStorage,
			action.NewUpdateSqlcConfig,
			cmdDb.NewUpdateSQLCConfig,
		).
		AddDependencies(
			*logger.NewModule(logger.ModuleConfig{}),
			*cli.NewModule(cli.ModuleConfig{}),
		)
}
