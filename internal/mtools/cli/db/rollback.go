package db

import (
	"braces.dev/errtrace"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/internal/mtools/action"
	"github.com/manifoldco/promptui"
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

func (c *Rollback) askModuleName() string {
	for {
		prompt := promptui.Prompt{
			Label: "What is the name of module to add migration to?: ",
		}

		moduleName, err := prompt.Run()
		if err != nil {
			fmt.Println(color.RedString("Cannot ask module name: %s", err.Error()))
			return ""
		}
		if moduleName == "" {
			fmt.Println(color.RedString("The module name cannot be empty"))
			continue
		}
		return moduleName
	}
}

func (c *Rollback) askMigrationName() string {
	for {
		prompt := promptui.Prompt{
			Label: "Enter a migration name : ",
		}

		migrationName, err := prompt.Run()
		if err != nil {
			fmt.Println(color.RedString("Cannot ask migration name: %s", err.Error()))
			return ""
		}
		if migrationName == "" {
			fmt.Println(color.RedString("The migration name cannot be empty"))
			continue
		}
		return migrationName
	}
}
