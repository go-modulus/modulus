package email

import (
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/providers/email/action"
	"github.com/go-modulus/modulus/auth/providers/email/graphql"
	"github.com/go-modulus/modulus/captcha"
	"github.com/go-modulus/modulus/module"
)

type ModuleConfig struct {
	// Add your module configuration here
	// e.g. Var1 string `env:"MYMODULE_VAR1, default=test"`
}
type AuthEmailProviderModule struct {
	*module.Module
}

func NewModule() *AuthEmailProviderModule {
	return &AuthEmailProviderModule{
		module.NewModule("auth email provider").
			// Add all dependencies of a module here
			AddDependencies(
				auth.NewModule(),
				captcha.NewModule(),
			).
			// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
			AddProviders(
				action.NewLogin,
				action.NewRegister,
				action.NewResetPassword,
				action.NewChangePassword,
				graphql.NewResolver,
			).
			SetOverriddenProvider("auth.email.UserCreator", action.NewDefaultUserCreator).
			SetOverriddenProvider("auth.email.VerifiedEmailChecker", action.NewDefaultVerifiedEmailChecker).
			SetOverriddenProvider("auth.email.MailSender", action.NewDefaultMailSender).
			// Add all your CLI commands here
			AddCliCommands().
			// Add all your configs here
			InitConfig(ModuleConfig{}).
			InitConfig(action.ResetPasswordConfig{}),
	}
}

func NewManifestModule() module.ManifestModule {
	emailModule := module.NewManifestModule(
		NewModule().Module,
		"github.com/go-modulus/modulus/auth/providers/email",
		"A provider for auth module to organize authentication via the email/password pair.",
		"1.0.0",
	)
	emailModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/providers/email/graphql/auth.graphql",
			DestFile:  "internal/auth/providers/email/graphql/auth.graphql",
		},
	)
	emailModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/auth/providers/email/prompt.md",
			DestFile:  "internal/auth/providers/email/prompt.md",
		},
	)
	emailModule.LocalPath = "internal/auth/providers/email"
	return emailModule
}

type Option func(*module.Module) *module.Module

func (m *AuthEmailProviderModule) WithOptions(opts ...Option) *AuthEmailProviderModule {
	for _, opt := range opts {
		m.Module = opt(m.Module)
	}
	return m
}

func OverrideUserCreator[T action.UserCreator](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider("auth.email.UserCreator", func(impl T) action.UserCreator { return impl })
}

func OverrideVerifiedEmailChecker[T action.VerifiedEmailChecker](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider(
		"auth.email.VerifiedEmailChecker",
		func(impl T) action.VerifiedEmailChecker { return impl },
	)
}

func OverrideMailSender[T action.MailSender](authModule *module.Module) *module.Module {
	return authModule.SetOverriddenProvider("auth.email.MailSender", func(impl T) action.MailSender { return impl })
}
