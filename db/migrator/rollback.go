package migrator

import (
	"braces.dev/errtrace"
	"context"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/urfave/cli/v2"
)

type Rollback struct {
	params CreateCommandParams
}

func NewRollback(params CreateCommandParams) *Rollback {
	return &Rollback{params: params}
}

func (c *Rollback) Command() *cli.Command {
	return &cli.Command{
		Name:  "rollback",
		Usage: "Rollback the last migration",
		Action: func(ctx *cli.Context) error {
			return c.Invoke(ctx.Context)
		},
	}
}

func (c *Rollback) Invoke(ctx context.Context) error {
	db, err := newDBMate(c.params)
	if err != nil {
		return errtrace.Wrap(err)
	}

	return errtrace.Wrap(db.Rollback())
}
