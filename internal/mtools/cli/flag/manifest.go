package flag

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus"
	"github.com/go-modulus/modulus/module"
	"github.com/urfave/cli/v2"
	"io/fs"
	"os"
	"path/filepath"
)

func NewManifest(usage string) cli.Flag {
	return &cli.StringFlag{
		Name:    "manifest",
		Usage:   usage,
		Aliases: []string{"mf"},
	}
}

func ManifestValue(ctx *cli.Context) (*module.Manifest, error) {
	manifestPath := ctx.String("manifest")
	var manifestFs fs.FS
	var err error
	var manifestFile string
	if manifestPath != "" {
		manifestDir := filepath.Dir(manifestPath)
		manifestFile = filepath.Base(manifestPath)
		manifestFs = os.DirFS(manifestDir)
	} else {
		manifestFs = modulus.ManifestFs
		manifestFile = "modules.json"
	}
	availableModulesManifest, err := module.NewFromFs(manifestFs, manifestFile)
	if err != nil {
		fmt.Println(color.RedString("Cannot read from the manifest file: %s", err.Error()))
		return nil, err
	}
	return availableModulesManifest, nil
}
