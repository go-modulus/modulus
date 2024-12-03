package cli

import (
	"context"
	"github.com/go-modulus/modulus/module"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

type ModuleConfig struct {
	Version string
	Usage   string
}

type StartCliParams struct {
	fx.In

	Lc       fx.Lifecycle
	Commands []*cli.Command `group:"cli.commands"`
	Runner   *Runner
	Config   *ModuleConfig
}

func NewApp(params StartCliParams) *cli.App {
	usage := "Run console commands"
	if params.Config.Usage != "" {
		usage = params.Config.Usage
	}
	app := &cli.App{
		Usage:                usage,
		Commands:             params.Commands,
		EnableBashCompletion: true,
		Version:              params.Config.Version,
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

func NewModule(config ModuleConfig) *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/cli").
		AddProviders(
			NewApp,
			NewRunner,
			module.ConfigProvider[ModuleConfig](config),
		)
}
