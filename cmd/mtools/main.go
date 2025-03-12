package main

import (
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/internal/mtools"
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
	cli2 "github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"os"
)

func main() {
	// current path
	path, _ := os.Getwd()
	cliModule := cli.NewModule().InitConfig(
		cli.ModuleConfig{
			Version: "0.3.0",
			Usage:   "This is a CLI tool for the Modulus framework. It helps you to create a new project, add modules, and run the project.",
			GlobalFlags: []cli2.Flag{
				&cli2.StringFlag{
					Name:    "proj-path",
					Usage:   "Set the path to the project if you want to run the command from another directory",
					Value:   path,
					Aliases: []string{"p"},
					EnvVars: []string{"PROJECT_PATH"},
				},
			},
		},
	)

	loggerModule := logger.NewModule().InitConfig(
		&logger.ModuleConfig{
			Type: "console",
			App:  "modulus cli",
		},
	)

	modules := []*module.Module{
		cliModule,
		loggerModule,
		mtools.NewModule(),
	}

	app := fx.New(
		module.BuildFx(modules...),
		fx.WithLogger(
			func(logger *zap.Logger) fxevent.Logger {
				logger = logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))

				return &fxevent.ZapLogger{Logger: logger}
			},
		),
		fx.Invoke(cli.Start),
	)

	app.Run()
}
