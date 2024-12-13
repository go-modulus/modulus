package migrator

import (
	"context"
	"fmt"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/urfave/cli/v2"
)

type Add struct {
	params CreateCommandParams
}

func NewAdd(params CreateCommandParams) *Add {
	return &Add{params: params}
}

func (c *Add) Command() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a new migration to the module",
		Action: func(ctx *cli.Context) error {
			return c.Invoke(
				ctx.Context,
				ctx.String("module"),
				ctx.String("name"),
			)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "module",
				Usage:    "The module to add the migration to",
				Required: true,
				Aliases:  []string{"m"},
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "The name of migration to add",
				Required: true,
				Aliases:  []string{"n"},
			},
		},
	}
}

func (c *Add) Invoke(ctx context.Context, module string, name string) error {
	u, _ := url.Parse(c.params.Pgx.Dsn())
	db := dbmate.New(u)
	migrationDir := "internal/" + module + "/storage/migration"
	db.MigrationsDir = []string{migrationDir}

	fmt.Println("Add a migration to the dir:" + migrationDir)
	err := db.NewMigration(name)
	if err != nil {
		return err
	}

	fmt.Println("\nMigration is created.")

	return nil
}
