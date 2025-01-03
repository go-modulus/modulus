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
	Version     string
	Usage       string
	GlobalFlags []cli.Flag
}

type StartCliParams struct {
	fx.In

	Lc       fx.Lifecycle
	Commands []*cli.Command `group:"cli.commands"`
	Runner   *Runner
	Config   ModuleConfig
}

func NewApp(params StartCliParams) *cli.App {
	usage := "Run console commands"
	if params.Config.Usage != "" {
		usage = params.Config.Usage
	}
	commands := params.Commands
	addGlobalFlagsToAllSubcommands(commands, params.Config.GlobalFlags)
	app := &cli.App{
		Usage:                usage,
		Commands:             commands,
		EnableBashCompletion: true,
		Suggest:              true,
		Version:              params.Config.Version,
		Flags:                params.Config.GlobalFlags,
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

func addGlobalFlagsToAllSubcommands(
	commands []*cli.Command,
	flags []cli.Flag,
) {
	for _, command := range commands {
		command.Flags = append(command.Flags, flags...)
		if len(command.Subcommands) != 0 {
			addGlobalFlagsToAllSubcommands(command.Subcommands, flags)
		}
	}
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
	return module.NewModule("urfave cli").
		AddProviders(
			NewApp,
			NewRunner,
		).InitConfig(ModuleConfig{})
}
