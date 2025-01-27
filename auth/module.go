package auth

import (
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/go-modulus/modulus/module"
)

// NewModule creates a new module for the auth package.
// It works with the default storage implementation. It uses pgxpool for database connection.
//
// If you want to use a custom storage implementation, you should create all interfaces provided as overridden
// and call authModule := auth.OverrideRepository(auth.NewModule(), NewStorageImplementation).
func NewModule() *module.Module {
	return module.NewModule("modulus auth").
		AddProviders(
			NewMiddlewareConfig,
			NewMiddleware,
		).
		SetOverriddenProvider("DefaultRepository", storage.NewDefaultRepository).
		SetOverriddenProvider("IdentityRepository", NewIdentityRepository)
}

// OverrideRepository overrides the default storage implementation with the custom one.
func OverrideRepository(authModule *module.Module, repository interface{}) *module.Module {
	return authModule.SetOverriddenProvider("IdentityRepository", repository).
		RemoveOverriddenProvider("DefaultRepository")
}

func NewManifestModule() module.ManifestModule {
	graphqlModule := module.NewManifestModule(
		NewModule(),
		"github.com/go-modulus/modulus/auth",
		"Authentication module. Helps protect HTTP routes with tokens and sessions. If you want to use default storage for identities and tokens, please install pgx module first.",
		"1.0.0",
	)
	graphqlModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/storage/migration/20240214134322_auth.sql",
			DestFile:  "internal/auth/storage/migration/20240214134322_auth.sql",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/install/module.go.tmpl",
			DestFile:  "internal/auth/install/module.go.tmpl",
		},
	).AppendPostInstallCommands(
		module.PostInstallCommand{
			CmdPackage: "github.com/go-modulus/modulus/cmd/mtools",
			Params:     []string{"db", "migrate"},
		},
	)

	graphqlModule.LocalPath = "internal/auth"

	return graphqlModule
}
