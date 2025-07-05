package google

import (
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/providers/google/action"
	"github.com/go-modulus/modulus/auth/providers/google/graphql"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("google auth").
		// Add all dependencies of a module here
		AddDependencies(
			pgx.NewModule(),
			auth.NewModule(),
		).
		// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
		AddProviders(
			graphql.NewResolver,
			action.NewRegister,
		).
		// Add all your CLI commands here
		AddCliCommands().
		SetOverriddenProvider("auth.google.UserCreator", action.NewDefaultUserCreator).
		// Add all your configs here
		InitConfig(action.GoogleConfig{})

}

func OverrideUserCreator[T action.UserCreator](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider("auth.google.UserCreator", func(impl T) action.UserCreator { return impl })
}

func NewManifestModule() module.ManifestModule {
	googleModule := module.NewManifestModule(
		NewModule(),
		"github.com/go-modulus/modulus/auth/providers/google",
		"Authentication provider for the auth module that helps auth using Google.",
		"1.0.0",
	)
	googleModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/providers/google/graphql/auth.graphql",
			DestFile:  "internal/auth/providers/google/graphql/auth.graphql",
		},
	)
	googleModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/providers/google/prompt.md",
			DestFile:  "internal/auth/providers/google/prompt.md",
		},
	)

	// docs
	googleModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/providers/google/README.md",
			DestFile:  "internal/auth/providers/google/README.md",
		},
	)
	googleModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/providers/google/img/name_of_google_client.png",
			DestFile:  "internal/auth/providers/google/img/name_of_google_client.png",
		},
	)

	googleModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/providers/google/img/register_new_project.png",
			DestFile:  "internal/auth/providers/google/img/register_new_project.png",
		},
	)
	googleModule.LocalPath = "internal/auth/providers/google"

	return googleModule
}
