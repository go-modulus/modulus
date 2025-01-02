package migrator

import (
	"braces.dev/errtrace"
	"context"
	"fmt"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/urfave/cli/v2"
)

type Migrate struct {
	params CreateCommandParams
}

func NewMigrate(params CreateCommandParams) *Migrate {
	return &Migrate{params: params}
}

func (c *Migrate) Command() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Apply all migrations from the registered modules to the database",
		Action: func(ctx *cli.Context) error {
			return c.Invoke(ctx.Context)
		},
	}
}

func (c *Migrate) Invoke(ctx context.Context) error {
	db, err := newDBMate(c.params)
	if err != nil {
		return errtrace.Wrap(err)
	}

	fmt.Println("\nApplying...")
	return errtrace.Wrap(db.CreateAndMigrate())
}
