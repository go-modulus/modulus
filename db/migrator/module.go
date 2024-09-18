package migrator

import (
	"braces.dev/errtrace"
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	infraCli "github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
	"github.com/urfave/cli/v2"
	"io/fs"
	"net/url"
)

type ModuleConfig struct {
	Pgx *pgx.ModuleConfig
	FS  fs.FS
}

func newDBMate(cfg *ModuleConfig) (*dbmate.DB, error) {
	u, _ := url.Parse(cfg.Pgx.Dsn())
	db := dbmate.New(u)
	db.FS = cfg.FS
	db.AutoDumpSchema = false

	migrationsDir, err := fs.Glob(cfg.FS, "./*/storage/migration")
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	db.MigrationsDir = migrationsDir

	return db, nil
}

func NewModule(config ModuleConfig) *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/db/migrator").
		AddDependencies(
			*pgx.NewModule(pgx.ModuleConfig{}),
			*infraCli.NewModule(),
		).
		AddProviders(
			module.ConfigProvider[ModuleConfig](config),
			NewMigrate,
			NewAdd,
			NewRollback,
			NewRollbackAll,
		).AddCliCommands(
		func(
			migrate *Migrate,
			add *Add,
			rollback *Rollback,
			rollbackAll *RollbackAll,
		) *cli.Command {
			return &cli.Command{
				Name:  "migrator",
				Usage: "Migrate your database",
				Subcommands: []*cli.Command{
					migrate.Command(),
					add.Command(),
					rollback.Command(),
					rollbackAll.Command(),
				},
			}
		},
	)
}
