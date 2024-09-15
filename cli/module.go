package cli

import (
	"context"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

func ProvideCommand(command interface{}) interface{} {
	return fx.Annotate(command, fx.ResultTags(`group:"cli.commands"`))
}

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
	return runner.start(func() error {
		return app.Run(os.Args)
	})
}

func NewModule() fx.Option {
	return fx.Module("cli", fx.Provide(NewApp, NewRunner))
}
