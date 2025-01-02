package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/db/migrator"
	"github.com/go-modulus/modulus/module"
	"os"
)

func main() {
	install := module.InstallManifest{}
	install.AppendEnvVars(module.GetEnvVariablesFromConfig(migrator.ModuleConfig{})...)

	manifest, err := module.LoadLocalManifest("./")
	if err != nil {
		fmt.Println("Cannot load the manifest file modules.json:", color.RedString(err.Error()))
		os.Exit(1)
	}
	currentPackage := "github.com/go-modulus/modulus/db/migrator"
	currentModule := module.ManifestModule{
		Name:        "db migrator",
		Package:     currentPackage,
		Description: "Several CLI commands to use DBMate (https://github.com/amacneil/dbmate) migration tool inside your application.",
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