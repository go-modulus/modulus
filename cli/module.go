package cli

import (
	"context"
	"github.com/go-modulus/modulus/module"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

type StartCliParams struct {
	fx.In

	Lc       fx.Lifecycle
	Commands []*cli.Command `group:"cli.commands"`
	Runner   *Runner
}

func NewApp(params StartCliParams) *cli.App {
	app := &cli.App{
		Usage:                "Run console commands",
		Commands:             params.Commands,
		EnableBashCompletion: true,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	params.Lc.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				return params.Runner.stop()
			},
		},
	)

	return app
}

func Start(
	runner *Runner,
	app *cli.App,
) error {
	return runner.start(
		func() error {
			return app.Run(os.Args)
		},
	)
}

func NewModule() *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/cli").
		AddConstructors(NewApp, NewRunner)
}
