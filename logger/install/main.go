package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
	"os"
)

func main() {
	install := module.InstallManifest{}
	install.AppendEnvVars(module.GetEnvVariablesFromConfig[logger.ModuleConfig](logger.ModuleConfig{})...)

	manifest, err := module.LoadLocalManifest("./")
	if err != nil {
		fmt.Println("Cannot load the manifest file modules.json:", color.RedString(err.Error()))
		os.Exit(1)
	}
	currentPackage := "github.com/go-modulus/modulus/logger"
	currentModule := module.ManifestModule{
		Name:        "Slog Logger with Zap Backend",
		Package:     currentPackage,
		Description: "Adds a slog logger with a zap backend to the Modulus framework.",
		Install:     install,
		Version:     "1.0.0",
	}

	manifest.UpdateModule(currentModule)

	err = manifest.SaveAsLocalManifest("./")
	if err != nil {
		fmt.Println("Cannot save the manifest file modules.json:", color.RedString(err.Error()))
		os.Exit(1)
	}
}
