package flag

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/module"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

func NewModule(usage string) cli.Flag {
	return &cli.StringFlag{
		Name:    "module",
		Usage:   usage,
		Aliases: []string{"m"},
	}
}

func ModuleValue(ctx *cli.Context) (module.ManifestModule, error) {
	isSilent := ctx.Bool("silent")
	moduleName := ctx.String("module")
	projPath := ctx.String("proj-path")
	manifest, err := module.LoadLocalManifest(projPath)
	if err != nil {
		fmt.Printf(
			"Cannot load the project manifest %s/modules.json: %s\n",
			color.BlueString(projPath),
			color.RedString(err.Error()),
		)
		return module.ManifestModule{}, err
	}

	if moduleName == "" {
		if isSilent {
			fmt.Println(color.RedString("The module name is required. Use the --module flag"))
			return module.ManifestModule{}, errors.New("module name is required")
		} else {
			moduleName, err = askModuleName(manifest)
			if err != nil {
				return module.ManifestModule{}, err
			}
		}
	}

	mod, found := manifest.FindLocalModule(moduleName)
	if !found {
		fmt.Println(
			color.RedString("Module with name"),
			color.BlueString(moduleName),
			color.RedString("is not found in the local manifest file"),
			color.BlueString("%s/modules.json", projPath),
		)
		fmt.Printf("Add one of the following values to the %s flag:\n", color.BlueString("--module"))
		for _, manifestModule := range manifest.LocalModules() {
			fmt.Println(color.BlueString(manifestModule.Name))
		}
		fmt.Println("")
		return module.ManifestModule{}, errors.New("module not found")
	}

	return mod, nil
}

func askModuleName(
	manifest module.Manifest,
) (string, error) {
	items := make([]string, 0)
	for _, md := range manifest.LocalModules() {
		items = append(items, md.Name)
	}
	sel := promptui.Select{
		Label: "Select a module",
		Items: items,
	}

	_, val, err := sel.Run()
	if err != nil {
		fmt.Println(color.RedString("Cannot ask module name: %s", err.Error()))
		return "", err
	}

	return val, nil
}
