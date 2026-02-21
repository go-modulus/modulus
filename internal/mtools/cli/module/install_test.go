package module_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

const localModulesJson = `{
  "name": "Modulus framework modules manifest",
  "version": "1.0.0",
  "description": "List of installed modules for the Modulus framework",
  "modules": [
    {
      "name": "urfave cli",
      "package": "github.com/go-modulus/modulus/cli",
      "description": "Adds ability to create cli applications in the Modulus framework.",
      "install": null,
      "version": "1.0.0"
    }
  ]
}`

const availableModulesJson = ` {
  "name": "Modulus framework modules manifest",
  "description": "List of modules available for the Modulus framework",
  "version": "1.0.0",
  "modules": [
    {
      "name": "urfave cli",
      "package": "github.com/go-modulus/modulus/cli",
      "description": "Adds ability to create cli applications in the Modulus framework.",
      "install": {
      },
      "version": "1.0.0"
    },
    {
      "name": "pgx",
      "package": "github.com/go-modulus/modulus/db/pgx",
      "description": "A wrapper for the pgx package to integrate it into the Modulus framework.",
      "install": {
        "envVars": [
          {
            "key": "DB_NAME",
            "value": "test",
            "comment": "Test comment"
          }
        ],
        "dependencies": [
          "slog logger"
        ]
      },
      "version": "1.0.0"
    },
    {
      "name": "slog logger",
      "package": "github.com/go-modulus/modulus/logger",
      "description": "Adds a slog logger with a zap backend to the Modulus framework.",
      "install": {
        "envVars": [
          {
            "key": "LOGGER_APP",
            "value": "modulus",
            "comment": ""
          }
        ]
      },
      "version": "1.0.0"
    },
    {
      "name": "dbmate migrator",
      "package": "github.com/go-modulus/modulus/db/migrator",
      "description": "Several CLI commands to use DBMate (https://github.com/amacneil/dbmate) migration tool inside your application.",
      "install": {
        "dependencies": [
          "pgx",
          "urfave cli"
        ]
      },
      "version": "1.0.0"
    },
	{
      "name": "gqlgen",
      "package": "github.com/go-modulus/modulus/graphql",
      "description": "Graphql server and generator. It is based on the gqlgen library. It also provides a playground for the graphql server.",
      "install": {
        "envVars": [
          {
            "key": "GQL_API_URL",
            "value": "/graphql",
            "comment": ""
          }
        ],
		"files": [
          {
            "sourceUrl": "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/module.go.tmpl",
            "destFile": "internal/graphql/module.go"
          }
		]
      },
      "version": "1.0.0",
      "localPath": "internal/graphql"
    }
  ]
}`

const localToolsGo = `//go:build tools
// +build tools

package tools

import _ "github.com/go-modulus/modulus/cli"
`

const consoleEntrypoint = `
package main

import (
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/config"

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
			module.BuildFx(modules),
			invokes...,
		)...,
	)

	app.Run()
}

func init() {
	config.LoadDefaultEnv()
}
`

const envFile = `# Environment variables for the project
APP_ENV=local
PG_HOST=myhost
`

const goModFile = `module testproj

go 1.23.1

require (
	github.com/go-modulus/modulus v0.0.4
)
`

const goModFullPathFile = `module github.com/test/testproj

go 1.23.1

require (
	github.com/go-modulus/modulus v0.0.4
)
`

func createFile(t *testing.T, projDir, filename, content string) {
	fn := fmt.Sprintf("%s/%s", projDir, filename)
	err := os.WriteFile(fn, []byte(content), 0644)
	if err != nil {
		t.Fatal("Cannot create "+fn+" file", err)
	}
}

func initProject(t *testing.T, projDir string, goModFile string) func() {
	if _, err := os.Stat(projDir); os.IsNotExist(err) {
		err = os.Mkdir(projDir, 0755)
		if err != nil {
			t.Fatal("Cannot create "+projDir+" dir", err)
		}
		manifestDir := projDir + "/manifest"
		err = os.Mkdir(manifestDir, 0755)
		if err != nil {
			t.Fatal("Cannot create "+manifestDir+" dir", err)
		}
		createFile(t, projDir, "tools.go", localToolsGo)
		createFile(t, projDir, "modules.json", localModulesJson)
		createFile(t, manifestDir, "modules.json", availableModulesJson)
		createFile(t, projDir, ".env", envFile)
		createFile(t, projDir, "go.mod", goModFile)

		err = os.Mkdir(fmt.Sprintf("%s/cmd", projDir), 0755)
		if err != nil {
			t.Fatal("Cannot create "+projDir+"/cmd dir", err)
		}
		err = os.Mkdir(fmt.Sprintf("%s/cmd/console", projDir), 0755)
		if err != nil {
			t.Fatal("Cannot create "+projDir+"/cmd/console dir", err)
		}
		createFile(t, projDir, "cmd/console/main.go", consoleEntrypoint)
	}

	return func() {
		_ = os.RemoveAll(projDir)
	}
}

func TestInstall_Invoke(t *testing.T) {
	t.Run(
		"install module without dependencies", func(t *testing.T) {
			projDir := "/tmp/testproj"
			rb := initProject(t, projDir, goModFile)
			defer rb()

			err := os.Chdir(projDir)
			require.NoError(t, err)
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			set.Var(cli.NewStringSlice("pgx"), "modules", "doc")
			_ = os.Chdir(projDir)
			set.String("manifest", "manifest/modules.json", "doc")
			ctx := cli.NewContext(app, set, nil)
			err = installModule.Invoke(ctx)

			toolsFileContent, errCont := os.ReadFile(fmt.Sprintf("%s/tools.go", projDir))
			entrypointFileContent, errCont2 := os.ReadFile(fmt.Sprintf("%s/cmd/console/main.go", projDir))
			envContent, errCont3 := os.ReadFile(fmt.Sprintf("%s/.env", projDir))
			modulesContent, errCont4 := os.ReadFile(fmt.Sprintf("%s/modules.json", projDir))

			t.Log("Given the tools.go file in the root of the project")
			t.Log("When install a new module to a project")
			t.Log("	The error should be nil")
			require.NoError(t, err)
			t.Log("	The new package should be added to the tools.go file")
			require.NoError(t, errCont)
			require.Contains(t, string(toolsFileContent), "github.com/go-modulus/modulus/db/pgx")
			t.Log("	The entrypoint file should be updated with the new module")
			require.NoError(t, errCont2)
			require.Contains(t, string(entrypointFileContent), "github.com/go-modulus/modulus/db/pgx")
			require.Contains(t, string(entrypointFileContent), "pgx.NewModule()")
			t.Log("	The .env file should be changed with new env variables")
			require.NoError(t, errCont3)
			require.Contains(t, string(envContent), "DB_NAME=test")
			t.Log("	The old env variables should not be overwritten")
			require.Contains(t, string(envContent), "APP_ENV=local")
			require.Contains(t, string(envContent), "PG_HOST=myhost")
			t.Log("	The comment should be added to the new env variable")
			require.Contains(t, string(envContent), "# Test comment")
			t.Log("	The modules.json file should be updated with the new module")
			require.NoError(t, errCont4)
			require.Contains(t, string(modulesContent), "github.com/go-modulus/modulus/db/pgx")
		},
	)

	t.Run(
		"install module with dependencies", func(t *testing.T) {
			projDir := "/tmp/testproj"
			rb := initProject(t, projDir, goModFullPathFile)
			defer rb()

			err := os.Chdir(projDir)
			require.NoError(t, err)
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			set.Var(cli.NewStringSlice("dbmate migrator"), "modules", "doc")
			set.String("manifest", projDir+"/manifest/modules.json", "doc")
			ctx := cli.NewContext(app, set, nil)
			err = installModule.Invoke(ctx)

			toolsFileContent, errCont := os.ReadFile(fmt.Sprintf("%s/tools.go", projDir))
			entrypointFileContent, errCont2 := os.ReadFile(fmt.Sprintf("%s/cmd/console/main.go", projDir))
			envContent, errCont3 := os.ReadFile(fmt.Sprintf("%s/.env", projDir))
			modulesContent, errCont4 := os.ReadFile(fmt.Sprintf("%s/modules.json", projDir))

			t.Log("Given the tools.go file in the root of the project")
			t.Log("When install a migrator that has dependencies on pgx")
			t.Log("	The error should be nil")
			require.NoError(t, err)
			t.Log("	Both pgx and migrator packages should be added to the tools.go file")
			require.NoError(t, errCont)
			require.Contains(t, string(toolsFileContent), "github.com/go-modulus/modulus/db/pgx")
			require.Contains(t, string(toolsFileContent), "github.com/go-modulus/modulus/db/migrator")
			t.Log("	The entrypoint file should be updated with the new two modules")
			require.NoError(t, errCont2)
			require.Contains(t, string(entrypointFileContent), "github.com/go-modulus/modulus/db/pgx")
			require.Contains(t, string(entrypointFileContent), "pgx.NewModule()")
			require.Contains(t, string(entrypointFileContent), "migrator.NewModule()")
			t.Log("	The .env file should be changed with pgx env variables")
			require.NoError(t, errCont3)
			require.Contains(t, string(envContent), "DB_NAME=test")
			t.Log("	The modules.json file should be updated with the new 2 modules")
			require.NoError(t, errCont4)
			require.Contains(t, string(modulesContent), "github.com/go-modulus/modulus/db/pgx")
			require.Contains(t, string(modulesContent), "github.com/go-modulus/modulus/db/migrator")
		},
	)

	t.Run(
		"install module with local path", func(t *testing.T) {
			projDir := "/tmp/testproj"
			rb := initProject(t, projDir, goModFullPathFile)
			defer rb()

			err := os.Chdir(projDir)
			require.NoError(t, err)
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			set.Var(cli.NewStringSlice("gqlgen"), "modules", "doc")
			set.String("manifest", projDir+"/manifest/modules.json", "doc")
			ctx := cli.NewContext(app, set, nil)
			err = installModule.Invoke(ctx)

			entrypointFileContent, errCont2 := os.ReadFile(fmt.Sprintf("%s/cmd/console/main.go", projDir))

			t.Log("Given the tools.go file in the root of the project")
			t.Log("When install a graphql module that has local path of module")
			t.Log("	The error should be nil")
			require.NoError(t, err)
			t.Log("	The entrypoint file should be updated with local module package")
			require.NoError(t, errCont2)
			require.Contains(t, string(entrypointFileContent), "github.com/test/testproj/internal/graphql")
			require.Contains(t, string(entrypointFileContent), "graphql2.NewModule()")
		},
	)
}
