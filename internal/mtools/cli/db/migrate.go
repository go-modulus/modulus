package db

import (
	"braces.dev/errtrace"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/internal/mtools/action"
	"github.com/urfave/cli/v2"
)

type Migrate struct {
	action *action.UpdateSqlcConfig
}

func NewMigrate(
	action *action.UpdateSqlcConfig,
) *Migrate {
	return &Migrate{
		action: action,
	}
}

func NewMigrateCommand(updateSqlc *Migrate) *cli.Command {
	return &cli.Command{
		Name: "migrate",
		Usage: `Migrates all migrations in all modules.
Example: mtools db migrate
Example: mtools db migrate --proj-path=/path/to/project/root
`,
		Action: updateSqlc.Invoke,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "local-manifest",
				Usage:   "Local manifest file related to the project root. Default is modules.json",
				Aliases: []string{"lmf"},
			},
		},
	}
}

func (c *Migrate) Invoke(ctx *cli.Context) error {
	projPath := ctx.String("proj-path")
	config, err := newPgxConfig(projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot load the project config: %s", err.Error()))
		return errtrace.Wrap(err)
	}

	manifest := ctx.String("local-manifest")

	projFs, err := commonMigrationFs(projPath, manifest)
	if err != nil {
		return errtrace.Wrap(err)
	}

	dbMate := newDBMate(config, projFs, []string{"migration"})
	err = dbMate.CreateAndMigrate()
	if err != nil {
		return errtrace.Wrap(err)
	}

	fmt.Println(
		color.GreenString(
			"All migrations are processed.",
		),
	)

	return nil
}
