package cli

import (
	"github.com/go-modulus/modulus/module"
	"go.uber.org/fx"
)

func NewModule(options ...module.Option) *module.Module {
	return module.NewModule("cli").
		AddProviders(
			NewRunner,
		).
		SetOverriddenProvider("cli.App", NewApp).
		SetOverriddenProvider("cli.ErrorHandler", NewLogErrorHandler).
		InitConfig(ModuleConfig{}).
		WithOptions(options...)
}

func NewManifesto() module.Manifesto {
	return module.NewManifesto(
		NewModule(),
		"github.com/go-modulus/modulus/cli",
		"Cli applications module for the Modulus framework. It is based on github.com/urfave/cli library.",
		"1.0.0",
	)
}

func OverrideApp[T App](m *module.Module) *module.Module {
	return m.SetOverriddenProvider("cli.App", func(impl T) App { return impl })
}

func OverrideErrorHandler[T ErrorHandler](m *module.Module) *module.Module {
	return m.SetOverriddenProvider("cli.ErrorHandler", func(impl T) ErrorHandler { return impl })
}

func InvokeStartCli() fx.Option {
	return fx.Invoke(Start)
}

func SetConfig(config ModuleConfig) module.Option {
	return func(m *module.Module) *module.Module {
		return m.InitConfig(config)
	}
}
