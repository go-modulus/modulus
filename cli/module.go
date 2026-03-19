package cli

import (
	"context"

	"github.com/go-modulus/modulus/cli/internal"
	"github.com/go-modulus/modulus/module"
)

type ModuleConfig = internal.ModuleConfig
type Runner interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewModule(options ...module.Option) *module.Module {
	return module.NewModule("cli").
		SetOverriddenProvider("cli.App", internal.NewApp).
		SetOverriddenProvider("cli.Runner", internal.NewRunner).
		InitConfig(internal.ModuleConfig{}).
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

func OverrideApp[T internal.App](m *module.Module) *module.Module {
	return m.SetOverriddenProvider("cli.App", func(impl T) internal.App { return impl })
}

func OverrideRunner[T Runner](m *module.Module) *module.Module {
	return m.SetOverriddenProvider("cli.Runner", func(impl T) Runner { return impl })
}

func InvokeStartCli(m *module.Module) *module.Module {
	return m.AddInvokes(internal.Start)
}

func SetConfig(config ModuleConfig) module.Option {
	return func(m *module.Module) *module.Module {
		return m.InitConfig(config)
	}
}
