package files_test

import (
	"fmt"
	"github.com/go-modulus/modulus/internal/mtools/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
	"os"
	"strings"
	"testing"
)

var fileContent = `//go:build tools
// +build tools

package tools

import _ "github.com/vektra/mockery/v2"
import _ "github.com/rakyll/gotest"

`

var entrypointContent = `package main

import (
	"github.com/go-modulus/modulus/cli"
	cfg "github.com/go-modulus/modulus/config"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	loggerOption := fx.WithLogger(
		func(logger *zap.Logger) fxevent.Logger {
			logger = logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))

			return &fxevent.ZapLogger{Logger: logger}
		},
	)
	// Add your project modules here
	// for example:
	// cli.NewModule().BuildFx(),
	projectModulesOptions := []fx.Option{
		loggerOption,
	}

	// DO NOT Remove. It will be edited by the add-module CLI command.
	importedModulesOptions := []fx.Option{
		cli.NewModule(
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
			append(
				projectModulesOptions,
				importedModulesOptions...,
			), invokes...,
		)...,
	)

	app.Run()
}

func init() {
	config.LoadDefaultEnv()
}

`

func TestAddPackageToGoFile(t *testing.T) {
	t.Run(
		"Add a new package to the go file if import is not exist", func(t *testing.T) {

			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(fileContent), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}

			_, err = files.AddImportToGoFile("github.com/stretchr/testify", "_", fn)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given a go file")
			t.Log("When add a new package to the go file")
			t.Log("	The new package should be added to the go file")
			require.Contains(t, string(fc), "import _ \"github.com/stretchr/testify\"")
		},
	)

	t.Run(
		"Do nothing if package is exist", func(t *testing.T) {

			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(fileContent), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}

			_, err = files.AddImportToGoFile("github.com/rakyll/gotest", "a", fn)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given a go file")
			t.Log("When add the existent package to the go file")
			t.Log("	The old package should not be changed")
			require.Contains(t, string(fc), "import _ \"github.com/rakyll/gotest\"")
			t.Log("	The new package should not be added")
			require.NotContains(t, string(fc), "import a \"github.com/rakyll/gotest\"")
		},
	)
}

func TestAddImportToTools(t *testing.T) {
	t.Run(
		"Create tools.go if not exists", func(t *testing.T) {
			dir := fmt.Sprintf("/tmp/%s", randstr.String(10))
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err = os.Mkdir(dir, 0755)
				if err != nil {
					t.Fatal("Cannot create /tmp/testproj dir", err)
				}
			}
			defer os.Remove("tools.go")
			defer os.Remove(dir)

			err := os.Chdir(dir)
			require.NoError(t, err)

			err = files.AddImportToTools("github.com/stretchr/testify")
			require.NoError(t, err)

			fc, err := os.ReadFile("tools.go")
			require.NoError(t, err)
			t.Log("Given a go file")
			t.Log("When add a new package to the go file")
			t.Log("	The new package should be added to the tools.go file")
			require.Contains(t, string(fc), "import _ \"github.com/stretchr/testify\"")
			t.Log("The tools.go file should be created with package tools")
			require.Contains(t, string(fc), "package tools")
		},
	)
}

func TestAddModuleToEntrypoint(t *testing.T) {
	t.Run(
		"Add a module to the CLI entrypoint without package alias", func(t *testing.T) {
			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(entrypointContent), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}
			err = files.AddModuleToEntrypoint(
				"github.com/stretchr/testify",
				fn,
			)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given a list of packages in the go file that does not contain the new package alias")
			t.Log("When add a new module to the go file")
			t.Log("	The new module should be added to the array of imported modules")
			assert.Contains(t, string(fc), "testify.NewModule().BuildFx(),")
			t.Log("	The new import should be added to the go file")
			assert.Contains(t, string(fc), "\"github.com/stretchr/testify\"")

		},
	)

	t.Run(
		"Add a module to the CLI entrypoint with package alias", func(t *testing.T) {
			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(entrypointContent), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}
			err = files.AddModuleToEntrypoint(
				"github.com/stretchr/cfg",
				fn,
			)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given an imported package with alias in the go file")
			t.Log("When add a new module to the go file with different package but the same alias")
			t.Log("	The new module should be added to the array of imported modules with the new alias")
			assert.Contains(t, string(fc), "cfg2.NewModule().BuildFx(),")
			t.Log("	The new import should be added to the go file with the new alias")
			assert.Contains(t, string(fc), "cfg2 \"github.com/stretchr/cfg\"")

		},
	)

	t.Run(
		"Skip adding a module if it is already added", func(t *testing.T) {
			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(entrypointContent), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}
			err = files.AddModuleToEntrypoint(
				"github.com/go-modulus/modulus/cli",
				fn,
			)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			// one cli.NewModule().BuildFx() is present in comment and one is present in the array of imported modules
			modulesInitsCount := strings.Count(string(fc), "cli.NewModule(")
			importsCount := strings.Count(string(fc), "\"github.com/go-modulus/modulus/cli\"")

			t.Log("Given an already added module to the go file")
			t.Log("When add the same module to the go file")
			t.Log("	The new module should NOT be added to the array of imported modules")
			assert.Equal(t, 2, modulesInitsCount)
			t.Log("	The new import should be added to the go file")
			assert.Equal(t, 1, importsCount)

		},
	)
}
