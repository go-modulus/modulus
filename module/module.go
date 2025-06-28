package module

import (
	"context"
	"fmt"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/fx"
	"reflect"
	"sort"
)

type Module struct {
	dependencies        []*Module
	cliCommandProviders []interface{}
	providers           []interface{}
	invokes             []interface{}
	configs             map[string]interface{}
	name                string
	envVars             []ConfigEnvVariable
	fxOptions           []fx.Option
	taggedProviders     map[string][]interface{}
	overriddenProviders map[string]interface{}

	exposeCommands bool
	hiddenTags     map[string]struct{}
}

func NewModule(name string) *Module {
	return &Module{
		name:           name,
		exposeCommands: true,
		configs:        make(map[string]interface{}),
	}
}

func (m *Module) AddDependencies(dependency ...*Module) *Module {
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

func (m *Module) AddTaggedProviders(tag string, providers ...interface{}) *Module {
	if m.taggedProviders == nil {
		m.taggedProviders = make(map[string][]interface{})
	}
	m.taggedProviders[tag] = append(m.taggedProviders[tag], providers...)
	return m
}

func (m *Module) SetOverriddenProvider(name string, provider interface{}) *Module {
	if m.overriddenProviders == nil {
		m.overriddenProviders = make(map[string]interface{})
	}
	m.overriddenProviders[name] = provider
	return m
}

func (m *Module) RemoveOverriddenProvider(name string) *Module {
	if m.overriddenProviders == nil {
		return m
	}
	delete(m.overriddenProviders, name)
	return m
}

func (m *Module) HideTags(tags ...string) *Module {
	if m.hiddenTags == nil {
		m.hiddenTags = make(map[string]struct{})
	}
	for _, tag := range tags {
		m.hiddenTags[tag] = struct{}{}
	}
	return m
}

func (m *Module) buildFx() fx.Option {
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
	}
	if len(m.taggedProviders) > 0 {
		for tag, taggedProviders := range m.taggedProviders {
			if _, ok := m.hiddenTags[tag]; !ok {
				opts = append(opts, fx.Provide(taggedProviders...))
			}
		}
	}
	if len(m.overriddenProviders) > 0 {
		for _, provider := range m.overriddenProviders {
			opts = append(opts, fx.Provide(provider))
		}
	}
	if len(m.configs) > 0 {
		supplies := make([]interface{}, 0, len(m.configs))
		for _, config := range m.configs {
			supplies = append(supplies, config)
		}
		opts = append(opts, fx.Supply(supplies...))
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

// InitConfig fills the config struct with the default values
// and adds it to the module if it doesn't exist.
// It can be called multiple times with different config structs.
// The last value added before the BuildFx() call will be used.
// Note: After the BuildFx() call, the config struct will be immutable.
// Note: Passed values of a struct have the highest priority. Env variables can overwrite only default values.
func (m *Module) InitConfig(config any) *Module {
	val := reflect.ValueOf(config)
	if val.Kind() != reflect.Ptr {
		vp := reflect.New(val.Type())
		vp.Elem().Set(val)
		config = vp.Interface()
	}

	err := envconfig.Process(context.Background(), config)
	if err != nil {
		panic(err)
	}

	val = reflect.ValueOf(config)

	filledConfig := val.Elem().Interface()
	name := m.getConfigName(filledConfig)
	m.configs[name] = filledConfig

	vars := getVariables(config, false)
	for _, value := range vars {
		m.envVars = append(m.envVars, value)
	}
	sort.Slice(
		m.envVars, func(i, j int) bool {
			return m.envVars[i].Key < m.envVars[j].Key
		},
	)
	return m
}

func (m *Module) getConfigName(config any) string {
	t := reflect.TypeOf(config)
	pckgPath := t.PkgPath()
	nameOfType := t.Name()

	return pckgPath + "." + nameOfType
}

func BuildFx(modules ...*Module) fx.Option {
	var builtModules = make(map[string]struct{})
	return buildFx(modules, builtModules, 0)
}

func buildFx(
	modules []*Module,
	builtModules map[string]struct{},
	level int,
) fx.Option {
	opts := make([]fx.Option, 0, len(modules))
	for _, module := range modules {
		opts = append(opts, module.buildFx())
		builtModules[module.name] = struct{}{}
	}

	// Add dependencies
	deps := make([]*Module, 0, len(builtModules))
	for _, module := range modules {
		for _, dep := range module.dependencies {
			if _, ok := builtModules[dep.name]; !ok {
				deps = append(deps, dep)
				builtModules[dep.name] = struct{}{}
			}
		}
	}

	if len(deps) > 0 {
		opts = append(opts, buildFx(deps, builtModules, level+1))
	}

	return fx.Module(fmt.Sprintf("system-container-level-%d", level), opts...)
}
