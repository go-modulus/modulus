package migrator

import (
	"braces.dev/errtrace"
	"context"
	"errors"
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type RollbackAll struct {
	cfg    *ModuleConfig
	logger *zap.Logger
}

func NewRollbackAll(cfg *ModuleConfig, logger *zap.Logger) *RollbackAll {
	return &RollbackAll{cfg: cfg, logger: logger}
}

func (c *RollbackAll) Command() *cli.Command {
	return &cli.Command{
		Name:  "rollback-all",
		Usage: "Rollback all migrations",
		Action: func(ctx *cli.Context) error {
			return c.Invoke(ctx.Context)
		},
	}
}

func (c *RollbackAll) Invoke(ctx context.Context) error {
	db, err := newDBMate(c.cfg)
	if err != nil {
		return errtrace.Wrap(err)
	}

	for {
		err := db.Rollback()
		if err != nil {
			if errors.Is(err, dbmate.ErrNoRollback) {
				break
			}
			return errtrace.Wrap(err)
		}
	}

	return nil
}
