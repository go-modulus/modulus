package user

import (
	"blog/internal/user/action"
	"blog/internal/user/graphql"
	"blog/internal/user/storage"
	"embed"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"io/fs"
)

//go:embed storage/migration/*.sql
var migrationFS embed.FS

type ModuleConfig struct {
	// Add your module configuration here
	// e.g. Var1 string `env:"MYMODULE_VAR1, default=test"`
}

func NewModule() *module.Module {
	return module.NewModule("user").
		// Add all dependencies of a module here
		AddDependencies(
			pgx.NewModule(),
		).
		// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
		AddProviders(
			action.NewRegisterUser,
			action.NewLoginUser,
			graphql.NewResolver,
			func(db *pgxpool.Pool) storage.DBTX {
				return db
			},
			func(db storage.DBTX) *storage.Queries {
				return storage.New(db)
			},
			fx.Annotate(func() fs.FS { return migrationFS }, fx.ResultTags(`group:"migrator.migration-fs"`)),
		).
		// Add all your CLI commands here
		AddCliCommands().
		// Add all your configs here
		InitConfig(ModuleConfig{})
}
