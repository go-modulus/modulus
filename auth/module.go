package auth

import (
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
)

// NewModule creates a new module for the auth package.
// It works with the default storage implementation. It uses pgxpool for database connection.
//
// If you want to use a custom identity storage implementation, you should implement the IdentityRepository interface
// and call authModule := auth.OverrideIdentityRepository(auth.NewModule(), NewYourIdentityRepositoryImplementation).
// The same is for other storage implementations if you need it.
func NewModule() *module.Module {
	return module.NewModule("modulus auth").
		AddDependencies(pgx.NewModule()).
		AddProviders(
			NewMiddlewareConfig,
			NewMiddleware,
			NewPasswordAuthenticator,
		).
		SetOverriddenProvider("CredentialRepository", NewDefaultCredentialRepository).
		SetOverriddenProvider("IdentityRepository", NewDefaultIdentityRepository)
}

// OverrideIdentityRepository overrides the default identity storage implementation with the custom one.
func OverrideIdentityRepository(authModule *module.Module, repository interface{}) *module.Module {
	return authModule.SetOverriddenProvider("IdentityRepository", repository)
}

// OverrideCredentialRepository overrides the default credential storage implementation with the custom one.
func OverrideCredentialRepository(authModule *module.Module, repository interface{}) *module.Module {
	return authModule.SetOverriddenProvider("CredentialRepository", repository)
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
			DestFile:  "internal/auth/module.go",
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
