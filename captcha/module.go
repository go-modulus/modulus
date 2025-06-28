package captcha

import (
	"github.com/go-modulus/modulus/captcha/action"
	"github.com/go-modulus/modulus/module"
)

type ModuleConfig struct {
	// Add your module configuration here
	// e.g. Var1 string `env:"MYMODULE_VAR1, default=test"`
}

func NewModule() *module.Module {
	return module.NewModule("captcha").
		// Add all dependencies of a module here
		AddDependencies().
		// Add all your services here. DO NOT DELETE AddProviders call. It is used for code generation
		AddProviders(
			action.NewCheckCaptcha,
		).
		// Add all your CLI commands here
		AddCliCommands().
		// Add all your configs here
		InitConfig(ModuleConfig{}).
		InitConfig(action.RecaptchaConfig{})
}

func NewManifestModule() module.ManifestModule {
	temporalModule := module.NewManifestModule(
		NewModule(),
		"github.com/go-modulus/modulus/captcha",
		"Captcha processor that have to be integrated in auth queries to protect against bots registrations.",
		"1.0.0",
	)
	temporalModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/captcha/install/graphql/captcha.graphql",
			DestFile:  "internal/captcha/graphql/captcha.graphql",
		},
	)
	return temporalModule
}
