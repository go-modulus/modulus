package migrator

import (
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	infraCli "github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
	"github.com/laher/mergefs"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"io/fs"
	"net/url"
)

type ModuleConfig struct {
}

type CreateCommandParams struct {
	fx.In

	Fs  []fs.FS `group:"migrator.migration-fs"`
	Pgx pgx.ModuleConfig
}

func newDBMate(params CreateCommandParams) (*dbmate.DB, error) {
	u, _ := url.Parse(params.Pgx.Dsn())
	db := dbmate.New(u)
	db.FS = mergefs.Merge(params.Fs...)
	db.AutoDumpSchema = false
	db.MigrationsDir = []string{"./storage/migration"}

	return db, nil
}

func NewModule() *module.Module {
	return module.NewModule("dbmate migrator").
		AddDependencies(
			pgx.NewModule(),
			infraCli.NewModule(),
		).
		InitConfig(ModuleConfig{}).
		AddProviders(
			NewMigrate,
			NewRollback,
			NewRollbackAll,
		).AddCliCommands(
		func(
			migrate *Migrate,
			rollback *Rollback,
			rollbackAll *RollbackAll,
		) *cli.Command {
			return &cli.Command{
				Name:  "migrator",
				Usage: "Migrate your database",
				Subcommands: []*cli.Command{
					migrate.Command(),
					rollback.Command(),
					rollbackAll.Command(),
				},
			}
		},
	)
}
