package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
	"os"
)

func main() {
	install := module.InstallManifest{}
	install.AppendEnvVars(module.GetEnvVariablesFromConfig(pgx.ModuleConfig{})...)

	manifest, err := module.LoadLocalManifest("./")
	if err != nil {
		fmt.Println("Cannot load the manifest file modules.json:", color.RedString(err.Error()))
		os.Exit(1)
	}
	currentPackage := "github.com/go-modulus/modulus/db/pgx"
	currentModule := module.ManifestModule{
		Name:        "pgx",
		Package:     currentPackage,
		Description: "A wrapper for the pgx package to integrate it into the Modulus framework.",
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
