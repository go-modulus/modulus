package module_test

import (
	"flag"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
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
      "name": "urfave cli",
      "package": "github.com/go-modulus/modulus/cli",
      "description": "Adds HTTP server capabilities to the Modulus framework. Uses the CHI router.",
      "install": {},
      "version": "1.0.0"
    },
    {
      "name": "http router",
      "package": "github.com/go-modulus/modulus/http",
      "description": "Adds HTTP server capabilities to the Modulus framework. Uses the CHI router.",
      "install": {},
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
            "comment": ""
          },
          {
            "key": "HOST",
            "value": "localhost",
            "comment": ""
          },
          {
            "key": "PASSWORD",
            "value": "foobar",
            "comment": ""
          },
          {
            "key": "PGX_DSN",
            "value": "postgres://postgres:foobar@localhost:5432/test?sslmode=disable",
            "comment": "Use this variable to set the DSN for the PGX connection. It overwrites the other PG_* variables."
          },
          {
            "key": "PORT",
            "value": "5432",
            "comment": ""
          },
          {
            "key": "SSL_MODE",
            "value": "disable",
            "comment": ""
          },
          {
            "key": "USER",
            "value": "postgres",
            "comment": ""
          }
        ]
      },
      "version": "1.0.0"
    },
{
      "name": "Slog Logger with Zap Backend",
      "package": "github.com/go-modulus/modulus/logger",
      "description": "Adds a slog logger with a zap backend to the Modulus framework.",
      "install": {
        "envVars": [
          {
            "key": "LOGGER_APP",
            "value": "modulus",
            "comment": ""
          },
          {
            "key": "LOGGER_FX_EVENT_LEVEL",
            "value": "info",
            "comment": "Use one of \"debug\", \"info\", \"warn\", \"error\". It sets the maximum level of the fx events that should be logged"
          },
          {
            "key": "LOGGER_LEVEL",
            "value": "debug",
            "comment": "Use one of \"debug\", \"info\", \"warn\", \"error\". It sets the maximum level of the log messages that should be logged"
          },
          {
            "key": "LOGGER_TYPE",
            "value": "console",
            "comment": "Use either \"console\" or \"json\" value"
          }
        ]
      },
      "version": "1.0.0"
    }
]}`

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
	loggerOption := fx.WithLogger(
		func(logger *zap.Logger) fxevent.Logger {
			logger = logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))

			return &fxevent.ZapLogger{Logger: logger}
		},
	)
	// Add your project modules here
	// for example:
	// cli.NewModule(cli.ModuleConfig{}).BuildFx(),
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

func createFile(t *testing.T, projDir, filename, content string) {
	fn := fmt.Sprintf("%s/%s", projDir, filename)
	err := os.WriteFile(fn, []byte(content), 0644)
	if err != nil {
		t.Fatal("Cannot create "+fn+" file", err)
	}
}

func initProject(t *testing.T, projDir string) func() {
	if _, err := os.Stat(projDir); os.IsNotExist(err) {
		err = os.Mkdir(projDir, 0755)
		if err != nil {
			t.Fatal("Cannot create "+projDir+" dir", err)
		}
		createFile(t, projDir, "tools.go", localToolsGo)
		createFile(t, projDir, "modules.json", localModulesJson)
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
		os.RemoveAll(projDir)
	}
}

func TestAddModule_Invoke(t *testing.T) {
	t.Run(
		"update tools.go with new module", func(t *testing.T) {
			projDir := "/tmp/testproj"
			rb := initProject(t, projDir)
			defer rb()

			err := os.Chdir(projDir)
			require.NoError(t, err)
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			set.Var(cli.NewStringSlice("pgx"), "modules", "doc")
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
			require.Contains(t, string(entrypointFileContent), "pgx.NewModule().BuildFx()")
			t.Log("	The .env file should be changed with new env variables")
			require.NoError(t, errCont3)
			require.Contains(t, string(envContent), "PGX_DSN=")
			t.Log("	The old env variables should not be overwritten")
			require.Contains(t, string(envContent), "APP_ENV=local")
			require.Contains(t, string(envContent), "PG_HOST=myhost")
			t.Log("	The modules.json file should be updated with the new module")
			require.NoError(t, errCont4)
			require.Contains(t, string(modulesContent), "github.com/go-modulus/modulus/db/pgx")
		},
	)
}
