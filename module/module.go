package module

import (
	"context"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/fx"
)

var builtModules = make(map[string]*Module)

func ConfigConstructor[T any](config T) interface{} {
	return func() (*T, error) {
		err := envconfig.Process(context.Background(), &config)
		return &config, err
	}
}

type Module struct {
	dependencies        []Module
	cliCommandProviders []interface{}
	providers           []interface{}
	name                string
	fxOptions           []fx.Option

	exposeCommands bool
}

func NewModule(name string) *Module {
	return &Module{
		name:           name,
		exposeCommands: true,
	}
}

func (m *Module) AddDependency(dependency Module) *Module {
	m.dependencies = append(m.dependencies, dependency)
	return m
}

// AddCliCommand adds a CLI command to the module.
// The commandConstructor is a constructor function that returns a CLI command.
// The depConstructors are optional providers that are used to provide dependencies to the commandConstructor.
func (m *Module) AddCliCommand(commandProvider interface{}, dependencyProviders ...any) *Module {
	m.cliCommandProviders = append(m.cliCommandProviders, commandProvider)
	for _, provider := range dependencyProviders {
		m.providers = append(m.providers, provider)
	}
	return m
}

func (m *Module) AddProviders(constructors ...interface{}) *Module {
	m.providers = append(m.providers, constructors...)
	return m
}

func (m *Module) AddFxOptions(option ...fx.Option) *Module {
	m.fxOptions = append(m.fxOptions, option...)
	return m

}

func (m *Module) BuildFx() fx.Option {
	opts := make([]fx.Option, 0, 2+len(m.dependencies))
	providers := make([]interface{}, 0, len(m.providers)+len(m.cliCommandProviders))
	providers = append(providers, m.providers...)
	if m.exposeCommands {
		for _, constructor := range m.cliCommandProviders {
			providers = append(providers, m.provideCommand(constructor))
		}
	}
	opts = append(opts, fx.Provide(providers...))
	builtModules[m.name] = m
	for _, dep := range m.dependencies {
		if _, ok := builtModules[dep.name]; !ok {
			opts = append(opts, dep.BuildFx())
		}
	}

	opts = append(opts, m.fxOptions...)

	return fx.Module(
		m.name,
		opts...,
	)
}

func (m *Module) HideCommands() *Module {
	m.exposeCommands = false
	return m
}

func (m *Module) provideCommand(command interface{}) interface{} {
	return fx.Annotate(command, fx.ResultTags(`group:"cli.commands"`))
}
