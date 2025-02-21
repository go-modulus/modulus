package auth

import (
	"github.com/go-modulus/modulus/auth/hash"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
	"time"
)

type ModuleConfig struct {
	AccessTokenTTL  time.Duration `env:"AUTH_ACCESS_TOKEN_TTL, default=1h"`
	RefreshTokenTTL time.Duration `env:"AUTH_REFRESH_TOKEN_TTL, default=720h"`
}

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
			NewPlainTokenAuthenticator,
		).
		SetOverriddenProvider("CredentialRepository", storage.NewDefaultCredentialRepository).
		SetOverriddenProvider("IdentityRepository", storage.NewDefaultIdentityRepository).
		SetOverriddenProvider("TokenRepository", storage.NewDefaultTokenRepository).
		SetOverriddenProvider("TokenHashStrategy", hash.NewSha1).
		SetOverriddenProvider(
			"MiddlewareAuthenticator", func(auth *PlainTokenAuthenticator) Authenticator {
				return auth
			},
		).
		InitConfig(&ModuleConfig{})
}

// OverrideIdentityRepository overrides the default identity storage implementation with the custom one.
// repository should be a constructor returning the implementation of the IdentityRepository interface.
func OverrideIdentityRepository(authModule *module.Module, repository interface{}) *module.Module {
	return authModule.SetOverriddenProvider("IdentityRepository", repository)
}

// OverrideCredentialRepository overrides the default credential storage implementation with the custom one.
// repository should be a constructor returning the implementation of the CredentialRepository interface.
func OverrideCredentialRepository(authModule *module.Module, repository interface{}) *module.Module {
	return authModule.SetOverriddenProvider("CredentialRepository", repository)
}

// OverrideTokenRepository overrides the default token storage implementation with the custom one.
// repository should be a constructor returning the implementation of the TokenRepository interface.
func OverrideTokenRepository(authModule *module.Module, repository interface{}) *module.Module {
	return authModule.SetOverriddenProvider("TokenRepository", repository)
}

// OverrideTokenHashStrategy overrides the default token hash strategy with the custom one.
// strategy should be a constructor returning the implementation of the hash.TokenHashStrategy interface.
// by default, the sha1 hash strategy is used.
// if you don't want to hash tokens, you can set the strategy to none, like this auth.OverrideTokenHashStrategy(authModule, hash.NewNone)
func OverrideTokenHashStrategy(authModule *module.Module, strategy interface{}) *module.Module {
	return authModule.SetOverriddenProvider("TokenHashStrategy", strategy)
}

// OverrideMiddlewareAuthenticator overrides the default middleware authenticator with the custom one.
// authenticator should be a constructor returning the implementation of the Authenticator interface.
func OverrideMiddlewareAuthenticator(authModule *module.Module, authenticator interface{}) *module.Module {
	return authModule.SetOverriddenProvider("MiddlewareAuthenticator", authenticator)
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
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/install/graphql/auth.graphql",
			DestFile:  "internal/auth/graphql/auth.graphql",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/install/graphql/directive.go",
			DestFile:  "internal/auth/graphql/directive.go",
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
