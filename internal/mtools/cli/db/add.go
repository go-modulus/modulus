package db

import (
	"braces.dev/errtrace"
	"errors"
	"fmt"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/internal/mtools/action"
	"github.com/go-modulus/modulus/module"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"os"
)

type Add struct {
	action *action.UpdateSqlcConfig
}

func NewAdd(
	action *action.UpdateSqlcConfig,
) *Add {
	return &Add{
		action: action,
	}
}

func NewAddCommand(updateSqlc *Add) *cli.Command {
	return &cli.Command{
		Name: "add",
		Usage: `Adds a migration to the storage/migration folder of the selected module.
Example: mtools db add
Example: mtools db add --proj-path=/path/to/project/root --module=example --name=create_table
`,
		Action: updateSqlc.Invoke,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "module",
				Usage: "A module name to add a migration to",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "A name of the migration",
			},
		},
	}
}

func (c *Add) Invoke(ctx *cli.Context) error {
	projPath := ctx.String("proj-path")
	manifest, err := module.LoadLocalManifest(projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot load the project manifest %s/modules.json: %s", projPath, err.Error()))
		return err
	}
	moduleName := ctx.String("module")
	if moduleName == "" {
		moduleName = c.askModuleName(manifest.Modules)
		if moduleName == "" {
			return errors.New("module name is empty")
		}
	}

	migrationName := ctx.String("name")
	if migrationName == "" {
		migrationName = c.askMigrationName()
		if migrationName == "" {
			return errors.New("migration name is empty")
		}
	}

	found := false
	for _, md := range manifest.Modules {
		if !md.IsLocalModule {
			continue
		}
		if md.Name != moduleName {
			continue
		}
		found = true
		storagePath := md.StoragePath(projPath)
		migrationFs := os.DirFS(projPath)

		config, err := newPgxConfig(projPath)
		if err != nil {
			fmt.Println(color.RedString("Cannot load the project config: %s", err.Error()))
			return errtrace.Wrap(err)
		}
		dbMate := newDBMate(config, migrationFs, []string{storagePath + "/migration"})
		err = dbMate.NewMigration(migrationName)
		if err != nil {
			return errtrace.Wrap(err)
		}

		fmt.Println(
			color.GreenString(
				"Migration is created. Fill it with SQL queries.",
			),
		)
	}
	if !found {
		fmt.Println(color.RedString("Module %s not found in the project", moduleName))
	}
	return nil
}

func (c *Add) askModuleName(modules []module.ManifestItem) string {
	items := make([]string, 0)
	for _, md := range modules {
		if md.IsLocalModule {
			items = append(items, md.Name)
		}
	}
	sel := promptui.Select{
		Label: "Select a module to add migration to",
		Items: items,
	}

	_, val, err := sel.Run()
	if err != nil {
		fmt.Println(color.RedString("Cannot ask module name: %s", err.Error()))
		return ""
	}
	return val
}

func (c *Add) askMigrationName() string {
	for {
		prompt := promptui.Prompt{
			Label: "Enter a migration name",
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
