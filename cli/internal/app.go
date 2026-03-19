package internal

import (
	"context"
	"os"
	"sort"

	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
)

type ModuleConfig struct {
	Version        string
	Usage          string
	DefaultCommand string
	GlobalFlags    []cli.Flag
}

type StartCliParams struct {
	fx.In

	Lc       fx.Lifecycle
	Commands []*cli.Command `group:"cli.commands"`
	Runner   *Runner
	Config   ModuleConfig
}

type App interface {
	Run(ctx context.Context, osArgs []string) (deferErr error)
}

func NewApp(params StartCliParams) App {
	usage := "Run console commands"
	if params.Config.Usage != "" {
		usage = params.Config.Usage
	}
	commands := params.Commands
	addGlobalFlagsToAllSubcommands(commands, params.Config.GlobalFlags)
	app := &cli.Command{
		Usage:                 usage,
		Version:               params.Config.Version,
		DefaultCommand:        params.Config.DefaultCommand,
		Commands:              commands,
		Flags:                 params.Config.GlobalFlags,
		EnableShellCompletion: true,
		Suggest:               true,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Slice(
		app.Commands, func(i, j int) bool {
			return app.Commands[i].Name < app.Commands[j].Name
		},
	)

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
		if len(command.Commands) != 0 {
			addGlobalFlagsToAllSubcommands(command.Commands, flags)
		}
	}
}

func Start(
	runner *Runner,
	app App,
) error {
	ctx := context.Background()
	return runner.start(
		func() error {
			return app.Run(ctx, os.Args)
		},
	)
}
