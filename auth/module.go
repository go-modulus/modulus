package auth

import (
	"github.com/go-modulus/modulus/auth/hash"
	"github.com/go-modulus/modulus/auth/locales"
	"github.com/go-modulus/modulus/auth/repository"
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
			locales.ProvideLocalesFs(),
		).
		SetOverriddenProvider("auth.CredentialRepository", storage.NewDefaultCredentialRepository).
		SetOverriddenProvider("auth.IdentityRepository", storage.NewDefaultIdentityRepository).
		SetOverriddenProvider("auth.TokenRepository", storage.NewDefaultTokenRepository).
		SetOverriddenProvider("auth.AccountRepository", storage.NewDefaultAccountRepository).
		SetOverriddenProvider("auth.ResetPasswordRequestRepository", storage.NewDefaultResetPasswordRequestRepository).
		SetOverriddenProvider("auth.TokenHashStrategy", hash.NewSha1).
		SetOverriddenProvider(
			"auth.MiddlewareAuthenticator", func(auth *PlainTokenAuthenticator) Authenticator {
				return auth
			},
		).
		InitConfig(&ModuleConfig{}).
		InitConfig(&storage.ResetPasswordConfig{})
}

// OverrideIdentityRepository overrides the default identity storage implementation with the custom one.
// repository should be a constructor returning the implementation of the IdentityRepository interface.
func OverrideIdentityRepository[T repository.IdentityRepository](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider(
		"auth.IdentityRepository",
		func(impl T) repository.IdentityRepository { return impl },
	)
}

// OverrideCredentialRepository overrides the default credential storage implementation with the custom one.
// repository should be a constructor returning the implementation of the CredentialRepository interface.
func OverrideCredentialRepository[T repository.CredentialRepository](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider(
		"auth.CredentialRepository",
		func(impl T) repository.CredentialRepository { return impl },
	)
}

// OverrideTokenRepository overrides the default token storage implementation with the custom one.
// repository should be a constructor returning the implementation of the TokenRepository interface.
func OverrideTokenRepository[T repository.TokenRepository](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider(
		"auth.TokenRepository",
		func(impl T) repository.TokenRepository { return impl },
	)
}

// OverrideAccountRepository overrides the default account storage implementation with the custom one.
// repository should be a constructor returning the implementation of the AccountRepository interface.
func OverrideAccountRepository[T repository.AccountRepository](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider(
		"auth.AccountRepository",
		func(impl T) repository.AccountRepository { return impl },
	)
}

// OverrideTokenHashStrategy overrides the default token hash strategy with the custom one.
// strategy should be a constructor returning the implementation of the hash.TokenHashStrategy interface.
// by default, the sha1 hash strategy is used.
// if you don't want to hash tokens, you can set the strategy to none, like this auth.OverrideTokenHashStrategy[*hash.None](authModule)
func OverrideTokenHashStrategy[T hash.TokenHashStrategy](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider(
		"auth.TokenHashStrategy",
		func(impl T) hash.TokenHashStrategy { return impl },
	)
}

// OverrideMiddlewareAuthenticator overrides the default middleware authenticator with the custom one.
// authenticator should be a constructor returning the implementation of the Authenticator interface.
func OverrideMiddlewareAuthenticator[T Authenticator](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider("auth.MiddlewareAuthenticator", func(impl T) Authenticator { return impl })
}

// OverrideResetPasswordRequestRepository overrides the default reset password request storage implementation with the custom one.
func OverrideResetPasswordRequestRepository[T repository.ResetPasswordRequestRepository](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider(
		"auth.ResetPasswordRequestRepository",
		func(impl T) repository.ResetPasswordRequestRepository { return impl },
	)
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
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/storage/migration/20240320084613_auth_account.sql",
			DestFile:  "internal/auth/storage/migration/20240320084613_auth_account.sql",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/storage/migration/20250508110252_add_reset_password_request_table.sql",
			DestFile:  "internal/auth/storage/migration/20250508110252_add_reset_password_request_table.sql",
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
