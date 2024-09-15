package main

import (
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/internal/mtools"
	"github.com/go-modulus/modulus/logger"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		append(
			[]fx.Option{
				cli.NewModule().BuildFx(),
				logger.NewModule(logger.ModuleConfig{App: "modulus cli"}).BuildFx(),
				mtools.NewModule().BuildFx(),
			},
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
