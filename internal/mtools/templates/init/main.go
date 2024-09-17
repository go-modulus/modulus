package main

import (
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/config"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	// Add your project modules here
	var projectModules []fx.Option
	app := fx.New(
		append(
			projectModules,
			cli.NewModule().BuildFx(),
			fx.WithLogger(
				func(logger *zap.Logger) fxevent.Logger {
					logger = logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))

					return &fxevent.ZapLogger{Logger: logger}
				},
			),
			fx.Invoke(cli.Start),
		)...,
	)

	app.Run()
}

func init() {
	config.LoadDefaultEnv()
}
