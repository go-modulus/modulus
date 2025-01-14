package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/db/migrator"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/graphql"
	"github.com/go-modulus/modulus/http"
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
		module.NewManifestModule(
			http.NewModule(),
			"github.com/go-modulus/modulus/http",
			"HTTP module based on the Chi router.",
			"1.0.0",
		),
		getGraphqlModule(),
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

func getGraphqlModule() module.ManifestModule {
	graphqlModule := module.NewManifestModule(
		graphql.NewModule(),
		"github.com/go-modulus/modulus/graphql",
		"Graphql server and generator. It is based on the gqlgen library. It also provides a playground for the graphql server. You need to install the `chi http` module to use this module.",
		"1.0.0",
	)
	graphqlModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/schema.graphql",
			DestFile:  "internal/graphql/schema.graphql",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/gqlgen.yaml",
			DestFile:  "gqlgen.yaml",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/types/time.go",
			DestFile:  "internal/graphql/types/time.go",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/types/time.graphql",
			DestFile:  "internal/graphql/types/time.graphql",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/types/uuid.go",
			DestFile:  "internal/graphql/types/uuid.go",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/types/void.go",
			DestFile:  "internal/graphql/types/void.go",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/gqlgen.mk",
			DestFile:  "mk/gqlgen.mk",
		},
	).AppendPostInstallCommands(
		module.PostInstallCommand{
			CmdPackage: "github.com/99designs/gqlgen",
			Params:     []string{"init"},
		},
	)

	return graphqlModule
}
