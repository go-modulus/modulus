package blog

import (
	"blog/internal/blog/api"
	"blog/internal/blog/graphql"
	"blog/internal/blog/storage"
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
	return module.NewModule("blog").
		// Add all dependencies of a module here
		AddDependencies(
			pgx.NewModule(),
		).
		// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
		AddProviders(
			graphql.NewResolver,
			func(db *pgxpool.Pool) storage.DBTX {
				return db
			},
			func(db storage.DBTX) *storage.Queries {
				return storage.New(db)
			},
			fx.Annotate(func() fs.FS { return migrationFS }, fx.ResultTags(`group:"migrator.migration-fs"`)),
			api.NewMain,
			api.NewMainRoute,
		).
		// Add all your CLI commands here
		AddCliCommands().
		// Add all your configs here
		InitConfig(ModuleConfig{})
}
