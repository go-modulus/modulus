package graphql

import (
	"embed"
	"github.com/go-modulus/modulus/module"
	"go.uber.org/fx"
	"io/fs"
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
