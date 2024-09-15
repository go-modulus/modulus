package migrator

import (
	"braces.dev/errtrace"
	"context"
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	infraCli "github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/pgx"
	"github.com/sethvargo/go-envconfig"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
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

func NewModule(config ModuleConfig) fx.Option {
	return fx.Module(
		"migrator",
		fx.Provide(
			NewMigrate,
			NewAdd,
			NewRollback,
			NewRollbackAll,
			func() (*ModuleConfig, error) {
				return &config, envconfig.Process(context.Background(), &config)
			},
			infraCli.ProvideCommand(
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
			),
		),
	)
}
