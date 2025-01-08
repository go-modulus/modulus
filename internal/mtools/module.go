package mtools

import (
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/internal/mtools/action"
	cmdRoot "github.com/go-modulus/modulus/internal/mtools/cli"
	cmdDb "github.com/go-modulus/modulus/internal/mtools/cli/db"
	cmdModule "github.com/go-modulus/modulus/internal/mtools/cli/module"
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/mtools").
		AddCliCommands(
			cmdDb.NewDbCommand,
			cmdRoot.NewInitProjectCommand,
			cmdModule.NewModuleCommand,
		).
		AddProviders(
			cmdRoot.NewInitProject,
			cmdModule.NewInstall,
			cmdModule.NewCreate,
			cmdModule.NewAddCli,
			action.NewInstallStorage,
			action.NewUpdateSqlcConfig,
			cmdDb.NewUpdateSQLCConfig,
			cmdDb.NewAdd,
			cmdDb.NewMigrate,
			cmdDb.NewRollback,
			cmdDb.NewGenerate,
		).
		AddDependencies(
			logger.NewModule(),
			cli.NewModule(),
			pgx.NewModule(),
		)
}
