package main

import (
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/config"
	"github.com/go-modulus/modulus/module"

	"go.uber.org/fx"
)

func main() {
	// DO NOT Remove. It will be edited by the `mtools module create` CLI command.
	modules := []*module.Module{
		cli.NewModule().InitConfig(
			cli.ModuleConfig{
				Version: "0.1.0",
				Usage:   "Run project commands",
			},
		),
	}

	invokes := []fx.Option{
		fx.Invoke(cli.Start),
	}

	app := fx.New(
		module.BuildFx(modules...),
		fx.Module("invokes", invokes...),
	)

	app.Run()
}

func init() {
	config.LoadDefaultEnv()
}
