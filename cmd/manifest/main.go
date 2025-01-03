package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/db/migrator"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
	"os"
)

func main() {
	modules := []module.ManifestModule{
		module.NewManifestModule(
			cli.NewModule(),
			"github.com/go-modulus/modulus/cli",
			"Adds ability to create cli applications in the Modulus framework.",
			"1.0.0",
		),
		module.NewManifestModule(
			pgx.NewModule(),
			"github.com/go-modulus/modulus/db/pgx",
			"A wrapper for the pgx package to integrate it into the Modulus framework.",
			"1.0.0",
		),
		module.NewManifestModule(
			logger.NewModule(),
			"github.com/go-modulus/modulus/logger",
			"Adds a slog logger with a zap backend to the Modulus framework.",
			"1.0.0",
		),
		module.NewManifestModule(
			migrator.NewModule(),
			"github.com/go-modulus/modulus/db/migrator",
			"Several CLI commands to use DBMate (https://github.com/amacneil/dbmate) migration tool inside your application.",
			"1.0.0",
		),
	}

	manifest, err := module.LoadLocalManifest("./")
	if err != nil {
		fmt.Println("Cannot load the manifest file modules.json:", color.RedString(err.Error()))
		os.Exit(1)
	}

	for _, currentModule := range modules {
		fmt.Println("Updating module", color.BlueString(currentModule.Name))
		manifest.UpdateModule(currentModule)
	}

	err = manifest.SaveAsLocalManifest("./")
	if err != nil {
		fmt.Println("Cannot save the manifest file modules.json:", color.RedString(err.Error()))
		os.Exit(1)
	}
}
