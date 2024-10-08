package module

import (
	"context"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/fx"
)

var builtModules = make(map[string]*Module)

func ConfigProvider[T any](config T) interface{} {
	return func() (*T, error) {
		err := envconfig.Process(context.Background(), &config)
		return &config, err
	}
}

type Module struct {
	dependencies        []Module
	cliCommandProviders []interface{}
	providers           []interface{}
	invokes             []interface{}
	name                string
	fxOptions           []fx.Option
	//@TODO: Add route handlers implementation
	//routeHandlers []interface{}

	exposeCommands bool
}

func NewModule(name string) *Module {
	return &Module{
		name:           name,
		exposeCommands: true,
	}
}

func (m *Module) AddDependencies(dependency ...Module) *Module {
	m.dependencies = append(m.dependencies, dependency...)
	return m
}

// AddCliCommands adds a CLI command to the module.
// The commandProvider is a constructor function that returns a CLI command.
func (m *Module) AddCliCommands(commandProvider ...interface{}) *Module {
	m.cliCommandProviders = append(m.cliCommandProviders, commandProvider...)

	return m
}

func (m *Module) AddProviders(constructors ...interface{}) *Module {
	m.providers = append(m.providers, constructors...)
	return m
}

func (m *Module) AddInvokes(invokes ...interface{}) *Module {
	m.invokes = append(m.invokes, invokes...)
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
	if len(m.providers) > 0 {
		opts = append(opts, fx.Provide(providers...))
		builtModules[m.name] = m
		for _, dep := range m.dependencies {
			if _, ok := builtModules[dep.name]; !ok {
				opts = append(opts, dep.BuildFx())
			}
		}
	}

	if len(m.invokes) > 0 {
		opts = append(opts, fx.Invoke(m.invokes...))
	}

	if len(m.fxOptions) > 0 {
		opts = append(opts, m.fxOptions...)
	}

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
