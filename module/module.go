package module

import (
	"go.uber.org/fx"
)

var builtModules = make(map[string]*Module)

type Module struct {
	dependencies           []Module
	cliCommandConstructors []interface{}
	constructors           []interface{}
	name                   string
	fxOptions              []fx.Option

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
// The depConstructors are optional constructors that are used to provide dependencies to the commandConstructor.
func (m *Module) AddCliCommand(commandConstructor interface{}, depConstructors ...any) *Module {
	m.cliCommandConstructors = append(m.cliCommandConstructors, commandConstructor)
	for _, constructor := range depConstructors {
		m.constructors = append(m.constructors, constructor)
	}
	return m
}

func (m *Module) AddConstructor(constructor interface{}) *Module {
	m.constructors = append(m.constructors, constructor)
	return m
}

func (m *Module) AddConstructors(constructors ...interface{}) *Module {
	m.constructors = append(m.constructors, constructors...)
	return m
}

func (m *Module) AddFxOption(option fx.Option) *Module {
	m.fxOptions = append(m.fxOptions, option)
	return m

}

func (m *Module) BuildFx() fx.Option {
	opts := make([]fx.Option, 0, 2+len(m.dependencies))
	constructors := make([]interface{}, 0, len(m.constructors)+len(m.cliCommandConstructors))
	constructors = append(constructors, m.constructors...)
	if m.exposeCommands {
		for _, constructor := range m.cliCommandConstructors {
			constructors = append(constructors, m.provideCommand(constructor))
		}
	}
	opts = append(opts, fx.Provide(constructors...))
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
