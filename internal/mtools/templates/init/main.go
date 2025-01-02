package main

import (
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/config"

	"go.uber.org/fx"
)

func main() {
	// DO NOT Remove. It will be edited by the `mtools module create` CLI command.
	importedModulesOptions := []fx.Option{
		cli.NewModule().InitConfig(
			cli.ModuleConfig{
				Version: "0.1.0",
				Usage:   "Run project commands",
			},
		).BuildFx(),
	}

	invokes := []fx.Option{
		fx.Invoke(cli.Start),
	}

	app := fx.New(
		append(
			importedModulesOptions,
			invokes...,
		)...,
	)

	app.Run()
}

func init() {
	config.LoadDefaultEnv()
}
