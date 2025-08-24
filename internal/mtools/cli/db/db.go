package db

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"

	"braces.dev/errtrace"
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/config"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
	"github.com/laher/mergefs"
	"github.com/sethvargo/go-envconfig"
	"github.com/urfave/cli/v2"
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
	_ = os.Setenv("CONFIG_DIR", projPath)
	config.LoadDefaultEnv()

	cfg := pgx.ModuleConfig{}
	err := envconfig.Process(context.Background(), &cfg)
	if err != nil {
		return pgx.ModuleConfig{}, err
	}

	return cfg, nil
}

func commonMigrationFs(projPath string, manifestFile string) (fs.FS, error) {
	projFs := os.DirFS(projPath)
	if manifestFile == "" {
		manifestFile = "modules.json"
	}
	manifest, err := module.NewFromFs(projFs, manifestFile)
	if err != nil {
		fmt.Println(color.RedString("Cannot load the project manifest %s/modules.json: %s", projPath, err.Error()))
		return nil, errtrace.Wrap(err)
	}

	modulesFs := make([]fs.FS, 0)

	for _, md := range manifest.Modules {
		if md.LocalPath == "" {
			continue
		}

		storagePath := md.StoragePath(projPath)
		if _, err := os.Stat(storagePath); os.IsNotExist(err) {
			continue
		}
		modulesFs = append(modulesFs, os.DirFS(storagePath))
	}

	return mergefs.Merge(modulesFs...), nil
}

func NewDbCommand(
	updateSqlc *UpdateSQLCConfig,
	add *Add,
	migrate *Migrate,
	rollback *Rollback,
	generate *Generate,
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
			NewRollbackCommand(rollback),
			NewGenerateCommand(generate),
		},
	}
}
