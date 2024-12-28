package db

import (
	"braces.dev/errtrace"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/internal/mtools/action"
	"github.com/go-modulus/modulus/module"
	"github.com/laher/mergefs"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"io/fs"
	"os"
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
	}
}

func (c *Migrate) Invoke(ctx *cli.Context) error {
	projPath := ctx.String("proj-path")
	config, err := newPgxConfig(projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot load the project config: %s", err.Error()))
		return errtrace.Wrap(err)
	}
	manifest, err := module.LoadLocalManifest(projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot load the project manifest %s/modules.json: %s", projPath, err.Error()))
		return errtrace.Wrap(err)
	}

	modulesFs := make([]fs.FS, 0)

	for _, md := range manifest.Modules {
		if !md.IsLocalModule {
			continue
		}

		storagePath := md.StoragePath(projPath)
		modulesFs = append(modulesFs, os.DirFS(storagePath))
	}

	projFs := mergefs.Merge(modulesFs...)

	dbMate := newDBMate(config, projFs, []string{"migration"})
	err = dbMate.Migrate()
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

func (c *Migrate) askModuleName() string {
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

func (c *Migrate) askMigrationName() string {
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
