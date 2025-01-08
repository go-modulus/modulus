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
	"github.com/go-modulus/modulus/module"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	// DO NOT Remove. It will be edited by the add-module CLI command.
	modules := []*module.Module{
		cli.NewModule(
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
		append(
			module.BuildFx(modules...),
			invokes...,
		)...,
	)

	app.Run()
}

func init() {
	config.LoadDefaultEnv()
}

`

const moduleContent = `package example

import (
	"go.uber.org/fx"
)

type ModuleConfig struct {
}

func NewModule() *module.Module {
	return module.NewModule("example").
		// Add all dependencies of a module here
		AddDependencies(
			pgx.NewModule(),
		).
		// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
		AddProviders(
			func(db *pgxpool.Pool) storage.DBTX {
				return db
			},
			func(db storage.DBTX) *storage.Queries {
				return storage.New(db)
			},
			fx.Annotate(func() fs.FS { return migrationFS }, fx.ResultTags("group:\"migrator.migration-fs\"")),
		).
		// Add all your CLI commands here
		AddCliCommands().
		// Add all your configs here
		InitConfig(ModuleConfig{})
}

`

const moduleContentEmptyProvider = `package example

import (
	"go.uber.org/fx"
)

func NewModule() *module.Module {
	return module.NewModule("example").
		// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
		AddProviders()
}

`

const moduleContentEmptyProviderNotTheLast = `package example

import (
	"go.uber.org/fx"
)

func NewModule() *module.Module {
	return module.NewModule("example").
		// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
		AddProviders().
		// Add all your CLI commands here
		AddCliCommands()
}

`

const moduleContentModuleInVar = `package example

import (
	"go.uber.org/fx"
)

func NewModule() *module.Module {
	m := module.NewModule("example").
		// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
		AddProviders().
		// Add all your CLI commands here
		AddCliCommands()
	return m
}

`

const moduleContentImportIsAlreadyAdded = `package example

import (
	"go.uber.org/fx"
	tes "github.com/stretchr/testify"
)

func NewModule() *module.Module {
	m := module.NewModule("example").
		// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
		AddProviders(
			tes.NewTest,
		).
		// Add all your CLI commands here
		AddCliCommands()
	return m
}

`

// use http://goast.yuroyoro.net/ to see the AST of the code
// https://astexplorer.net/ to see the AST of the code
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
			assert.Contains(t, string(fc), "testify.NewModule(),")
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
			assert.Contains(t, string(fc), "cfg2.NewModule(),")
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

			// one cli.NewModule() is present in comment and one is present in the array of imported modules
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

func TestAddConstructorToProvider(t *testing.T) {
	t.Run(
		"add provider to the empty AddProviders() function", func(t *testing.T) {
			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(moduleContentEmptyProvider), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}
			err = files.AddConstructorToProvider(
				"github.com/stretchr/testify",
				"NewTestProvider",
				fn,
			)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given a module constructor with the empty AddProviders() function")
			t.Log("When new provider is added to the module")
			t.Log("	The new provider should be added to the AddProviders() function")
			assert.Contains(t, string(fc), "testify.NewTestProvider")
			t.Log("	The new import should be added to the go file")
			assert.Contains(t, string(fc), "\"github.com/stretchr/testify\"")
		},
	)

	t.Run(
		"add provider to the empty AddProviders() if the function is not the last", func(t *testing.T) {
			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(moduleContentEmptyProviderNotTheLast), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}
			err = files.AddConstructorToProvider(
				"github.com/stretchr/testify",
				"NewTestProvider",
				fn,
			)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given a module constructor with the empty AddProviders() function that has a call after it")
			t.Log("When new provider is added to the module")
			t.Log("	The new provider should be added to the AddProviders() function")
			assert.Contains(t, string(fc), "testify.NewTestProvider")
			t.Log("	The new import should be added to the go file")
			assert.Contains(t, string(fc), "\"github.com/stretchr/testify\"")
		},
	)

	t.Run(
		"add provider if the module is created to variable", func(t *testing.T) {
			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(moduleContentModuleInVar), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}
			err = files.AddConstructorToProvider(
				"github.com/stretchr/testify",
				"NewTestProvider",
				fn,
			)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given a module constructor assigned to a variable")
			t.Log("When new provider is added to the module")
			t.Log("	The new provider should be added to the AddProviders() function")
			assert.Contains(t, string(fc), "testify.NewTestProvider")
			t.Log("	The new import should be added to the go file")
			assert.Contains(t, string(fc), "\"github.com/stretchr/testify\"")
		},
	)

	t.Run(
		"add provider if AddProviders() has params", func(t *testing.T) {
			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(moduleContent), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}
			err = files.AddConstructorToProvider(
				"github.com/stretchr/testify",
				"NewTestProvider",
				fn,
			)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given AddProviders() function with params")
			t.Log("When new provider is added to the module")
			t.Log("	The new provider should be added to the AddProviders() function")
			assert.Contains(t, string(fc), "testify.NewTestProvider")
			t.Log("	The new import should be added to the go file")
			assert.Contains(t, string(fc), "\"github.com/stretchr/testify\"")
		},
	)

	t.Run(
		"use already created import", func(t *testing.T) {
			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(moduleContentImportIsAlreadyAdded), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}
			err = files.AddConstructorToProvider(
				"github.com/stretchr/testify",
				"NewTestProvider",
				fn,
			)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given added import to the go file")
			t.Log("When new provider is added to the module")
			t.Log("	The new provider should be added to the AddProviders() function with the existing import alias")
			assert.Contains(t, string(fc), "tes.NewTestProvider")
		},
	)
}
