package graphql

import (
	"blog/internal/graphql/generated"
	"blog/internal/graphql/resolver"

	"github.com/99designs/gqlgen/graphql"
)

//go:embed storage/migration/*.sql
var migrationFS embed.FS

func NewModule() *module.Module {
	return module.NewModule(
		"auth",
	).AddProviders(
		fx.Annotate(func() fs.FS { return migrationFS }, fx.ResultTags(`group:"migrator.migration-fs"`)),
	)
}
