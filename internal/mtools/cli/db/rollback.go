package db

import (
	"braces.dev/errtrace"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/internal/mtools/action"
	"github.com/urfave/cli/v2"
)

type Rollback struct {
	action *action.UpdateSqlcConfig
}

func NewRollback(
	action *action.UpdateSqlcConfig,
) *Rollback {
	return &Rollback{
		action: action,
	}
}

func NewRollbackCommand(updateSqlc *Rollback) *cli.Command {
	return &cli.Command{
		Name: "rollback",
		Usage: `Rollbacks the last applied migration.
Example: mtools db rollback
Example: mtools db rollback --proj-path=/path/to/project/root
`,
		Action: updateSqlc.Invoke,
	}
}

func (c *Rollback) Invoke(ctx *cli.Context) error {
	projPath := ctx.String("proj-path")
	config, err := newPgxConfig(projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot load the project config: %s", err.Error()))
		return errtrace.Wrap(err)
	}

	projFs, err := commonMigrationFs(projPath)
	if err != nil {
		return errtrace.Wrap(err)
	}

	dbMate := newDBMate(config, projFs, []string{"migration"})
	err = dbMate.Rollback()
	if err != nil {
		return errtrace.Wrap(err)
	}

	fmt.Println(
		color.GreenString(
			"The last migration is rolled back.",
		),
	)

	return nil
}
