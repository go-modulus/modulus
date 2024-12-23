package module

import (
	"context"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/fx"
	"reflect"
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
	configs             map[string]interface{}
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
		configs:        make(map[string]interface{}),
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

//func (m *Module) FxSupply() fx.Option {
//	type test struct {
//		t string
//	}
//	supplies := make([]interface{}, 0, len(m.configs))
//	if len(m.configs) > 0 {
//		for _, config := range m.configs {
//			reflValue := reflect.ValueOf(config)
//			c := reflValue.Elem()
//			c2 := c.Elem()
//			c3 := c2.Interface()
//			err := envconfig.Process(context.Background(), &c2)
//			if err != nil {
//				panic(err)
//			}
//
//			supplies = append(supplies, c3)
//		}
//	}
//	return fx.Supply(supplies...)
//}

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
	m.configs[m.getConfigName(config)] = filledConfig

	return m
}

func (m *Module) getConfigName(config any) string {
	t := reflect.TypeOf(config)
	pckgPath := t.PkgPath()
	nameOfType := t.Name()

	return pckgPath + "." + nameOfType
}
