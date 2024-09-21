package main

import (
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/config"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	loggerOption := fx.WithLogger(
		func(logger *zap.Logger) fxevent.Logger {
			logger = logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))

			return &fxevent.ZapLogger{Logger: logger}
		},
	)
	// Add your project modules here
	// for example:
	// cli.NewModule().BuildFx(),
	projectModulesOptions := []fx.Option{
		loggerOption,
	}

	// DO NOT EDIT. It will be replaced by the add-module CLI command.
	importedModulesOptions := []fx.Option{
		cli.NewModule().BuildFx(),
	}

	invokes := []fx.Option{
		fx.Invoke(cli.Start),
	}

	app := fx.New(
		append(
			append(
				projectModulesOptions,
				importedModulesOptions...,
			), invokes...,
		)...,
	)

	app.Run()
}

func init() {
	config.LoadDefaultEnv()
}
