package db

import (
	"context"
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	"github.com/go-modulus/modulus/config"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/sethvargo/go-envconfig"
	"github.com/urfave/cli/v2"
	"io/fs"
	"net/url"
	"os"
)

func newDBMate(
	config pgx.ModuleConfig,
	projRootFs fs.FS,
	migrationsDir []string,
) *dbmate.DB {
	u, _ := url.Parse(config.Dsn())
	db := dbmate.New(u)
	db.FS = projRootFs
	db.AutoDumpSchema = false

	db.MigrationsDir = migrationsDir

	return db
}

func newPgxConfig(projPath string) (pgx.ModuleConfig, error) {
	config.LoadEnv(projPath, "", false)
	config.LoadEnv(projPath, os.Getenv("APP_ENV"), true)

	cfg := pgx.ModuleConfig{}
	err := envconfig.Process(context.Background(), &cfg)
	if err != nil {
		return pgx.ModuleConfig{}, err
	}

	return cfg, nil
}

func NewDbCommand(
	updateSqlc *UpdateSQLCConfig,
	add *Add,
	migrate *Migrate,
) *cli.Command {
	return &cli.Command{
		Name: "db",
		Usage: `A set of commands for working with PostgreSQL database in modules.
Example: mtools db
`,
		Subcommands: []*cli.Command{
			NewUpdateSQLCConfigCommand(updateSqlc),
			NewAddCommand(add),
			NewMigrateCommand(migrate),
		},
	}
}
