package db

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/internal/mtools/action"
	"github.com/go-modulus/modulus/module"
	"github.com/urfave/cli/v2"
)

type UpdateSQLCConfig struct {
	action *action.UpdateSqlcConfig
}

func NewUpdateSQLCConfig(action *action.UpdateSqlcConfig) *UpdateSQLCConfig {
	return &UpdateSQLCConfig{
		action: action,
	}
}

func NewUpdateSQLCConfigCommand(updateSqlc *UpdateSQLCConfig) *cli.Command {
	return &cli.Command{
		Name: "update-sqlc-config",
		Usage: `Updates the sqlc config file in all modules of the project.
Example: mtools db update-sqlc-config
`,
		Action: updateSqlc.Invoke,
	}
}

func (c *UpdateSQLCConfig) Invoke(ctx *cli.Context) error {
	projPath := ctx.String("proj-path")
	manifest, err := module.LoadLocalManifest(projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot load the project manifest %s/modules.json: %s", projPath, err.Error()))
		return err
	}
	for _, md := range manifest.Modules {
		if !md.IsLocalModule {
			continue
		}
		storagePath := md.StoragePath(projPath)
		err := c.action.Update(ctx.Context, storagePath, projPath)
		if err != nil {
			if errors.Is(err, action.ErrNoSqlcTmpl) {
				fmt.Println(
					color.YellowString(
						"No %s/storage/sqlc.tmpl.yaml template file found in the module. Skipping...",
						md.LocalPath,
					),
				)
				continue
			}
			fmt.Println(
				color.RedString(
					"Cannot update %s/storage/sqlc.yaml file for the module %s: %s",
					md.LocalPath,
					md.Name,
					err.Error(),
				),
			)
			continue
		}
		fmt.Println(color.GreenString("%s/storage/sqlc.yaml file updated", md.LocalPath))
	}
	return nil
}
