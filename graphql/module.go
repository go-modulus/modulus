package graphql

import (
	mHttp "github.com/go-modulus/modulus/http"
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("gqlgen").
		AddDependencies(
			mHttp.NewModule(),
		).
		AddProviders(
			NewGraphqlServer,
			NewLoadersInitializer,
			NewHandler,
			NewPlaygroundHandler,
			NewHandlerGetRoute,
			NewHandlerPostRoute,
			NewPlaygroundHandlerRoute,
		).
		SetOverriddenProvider("graphql.ErrorPresenter", NewErrorPresenter).
		InitConfig(Config{})
}

// OverrideErrorPresenter overrides the error presenter provider.
// ErrorPresenter is a function that converts any error to a graphql error.
func OverrideErrorPresenter(gqlModule *module.Module, presenter interface{}) *module.Module {
	return gqlModule.SetOverriddenProvider("graphql.ErrorPresenter", presenter)
}

// NewManifestModule creates a new graphql module with the manifest module.
func NewManifestModule() module.ManifestModule {
	graphqlModule := module.NewManifestModule(
		NewModule(),
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
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/types/uuid.graphql",
			DestFile:  "internal/graphql/types/uuid.graphql",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/types/void.go",
			DestFile:  "internal/graphql/types/void.go",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/types/void.graphql",
			DestFile:  "internal/graphql/types/void.graphql",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/gqlgen.mk",
			DestFile:  "mk/gqlgen.mk",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/module.go.tmpl",
			DestFile:  "internal/graphql/module.go",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/generated/tools.go",
			DestFile:  "internal/graphql/generated/tools.go",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/model/tools.go",
			DestFile:  "internal/graphql/model/tools.go",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/resolver/resolver.go.tmpl",
			DestFile:  "internal/graphql/resolver/resolver.go",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/graphql/install/resolver/schema.resolvers.go.tmpl",
			DestFile:  "internal/graphql/resolver/schema.resolvers.go",
		},
	).AppendPostInstallCommands(
		module.PostInstallCommand{
			CmdPackage: "github.com/99designs/gqlgen",
			Params:     []string{"generate", "--config", "gqlgen.yaml"},
		},
	)

	graphqlModule.LocalPath = "internal/graphql"

	return graphqlModule
}
