package module

import (
	"context"
	"testing"

	"github.com/go-modulus/modulus/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestModule_InitConfig(t *testing.T) {
	t.Parallel()
	t.Run(
		"Test InitConfig", func(t *testing.T) {
			type SubConfig struct {
				SubHost string `env:"SUB_HOST, default=subhost" comment:"SubHost comment"`
			}
			type PrefixedConfig struct {
				Host string `env:"HOST, default=prefhost" comment:"Prefixed SubHost comment"`
			}
			type Config struct {
				Host string `env:"HOST, default=localhost" comment:"Host comment"`

				PrefixedSubconfig *PrefixedConfig `env:",prefix=PREF_"`
				Subconfig         *SubConfig      `env:""`
			}
			module := NewModule("test")
			module = module.InitConfig(Config{})

			assert.Len(t, module.configs, 1)
			assert.Len(t, module.envVars, 3)

			varsMap := make(map[string]ConfigEnvVariable)
			keys := make([]string, 0)
			for _, envVar := range module.envVars {
				keys = append(keys, envVar.Key)
				varsMap[envVar.Key] = envVar
			}

			assert.Contains(t, keys, "HOST")
			assert.Contains(t, keys, "PREF_HOST")
			assert.Contains(t, keys, "SUB_HOST")

			assert.Equal(t, "localhost", varsMap["HOST"].Value)
			assert.Equal(t, "Host comment", varsMap["HOST"].Comment)
			assert.Equal(t, "prefhost", varsMap["PREF_HOST"].Value)
			assert.Equal(t, "Prefixed SubHost comment", varsMap["PREF_HOST"].Comment)
			assert.Equal(t, "subhost", varsMap["SUB_HOST"].Value)
			assert.Equal(t, "SubHost comment", varsMap["SUB_HOST"].Comment)
		},
	)
}
func TestModule_AddDependencies(t *testing.T) {
	t.Parallel()
	t.Run(
		"test build dependency graph when all dependencies mentioned in root", func(t *testing.T) {
			type T1 struct{}
			type T2 struct{}
			type T3 struct{}
			type T4 struct{}
			m1 := func() *Module {
				return NewModule("m1").AddProviders(
					func() T1 {
						return T1{}
					},
				)
			}
			m2 := func() *Module {
				return NewModule("m2").AddDependencies(m1()).AddProviders(
					func(t1 T1) T2 {
						return T2{}
					},
				)
			}

			m3 := func() *Module {
				return NewModule("m3").AddDependencies(m1()).AddProviders(
					func(t1 T1) T3 {
						return T3{}
					},
				)
			}
			m4 := NewModule("m4").AddDependencies(m2(), m3()).AddProviders(
				func(t1 T1, t2 T2, t3 T3) T4 {
					return T4{}
				},
			)

			app := fx.New(
				BuildFx([]*Module{m1(), m2(), m3(), m4}...),
				fx.Invoke(
					func(t2 T2, t3 T3, t4 T4) {
						assert.NotNil(t, t2)
						assert.NotNil(t, t3)
						assert.NotNil(t, t4)
					},
				),
			)

			err := app.Start(context.Background())
			require.NoError(t, err)
		},
	)

	t.Run(
		"test build dependency graph when firs dependency mentioned only in deps", func(t *testing.T) {
			type T1 struct{}
			type T2 struct{}
			type T3 struct{}
			type T4 struct{}
			m1 := func() *Module {
				return NewModule("m1").AddProviders(
					func() T1 {
						return T1{}
					},
				)
			}
			m2 := func() *Module {
				return NewModule("m2").AddDependencies(m1()).AddProviders(
					func(t1 T1) T2 {
						return T2{}
					},
				)
			}

			m3 := func() *Module {
				return NewModule("m3").AddDependencies(m1()).AddProviders(
					func(t1 T1) T3 {
						return T3{}
					},
				)
			}
			m4 := NewModule("m4").AddDependencies(m2(), m3()).AddProviders(
				func(t1 T1, t2 T2, t3 T3) T4 {
					return T4{}
				},
			)

			app := fx.New(
				BuildFx([]*Module{m2(), m3(), m4}...),
				fx.Invoke(
					func(t2 T2, t3 T3, t4 T4) {
						assert.NotNil(t, t2)
						assert.NotNil(t, t3)
						assert.NotNil(t, t4)
					},
				),
			)

			err := app.Start(context.Background())
			require.NoError(t, err)
		},
	)
}

func TestModule_WithOptions(t *testing.T) {
	t.Parallel()
	t.Run(
		"set overridden provider without overriding", func(t *testing.T) {
			t.Parallel()
			mod := NewModule("test").
				SetOverriddenProvider("InterfaceA", NewA)

			var intA InterfaceA
			err := test.Invoke(
				BuildFx(mod),
				fx.Populate(&intA),
			)
			require.NoError(t, err)
			require.Equal(t, "A", intA.MethodA())
		},
	)

	t.Run(
		"set overridden provider with overriding interface", func(t *testing.T) {
			t.Parallel()
			mod := NewModule("test").
				SetOverriddenProvider("InterfaceA", NewA).
				AddProviders(NewOverrideA).
				WithOptions(OverrideAInterfaceOption[*OverrideA])

			var intA InterfaceA
			err := test.Invoke(
				BuildFx(mod),
				fx.Populate(&intA),
			)
			require.NoError(t, err)
			require.Equal(t, "OverrideA", intA.MethodA())
		},
	)

	t.Run(
		"override obj constructor", func(t *testing.T) {
			t.Parallel()

			mod := NewModule("test").
				SetOverriddenProvider("A", NewAObj).
				AddProviders(NewOverrideA).
				WithOptions(OverrideAOption[*OverrideA])

			var intA *A
			err := test.Invoke(
				BuildFx(mod),
				fx.Populate(&intA),
			)
			require.NoError(t, err)
			require.Equal(t, "OverriddenA", intA.MethodA())
		},
	)
}

type InterfaceA interface {
	MethodA() string
}

type A struct {
	val string
}

func NewA() InterfaceA {
	return &A{
		val: "A",
	}
}
func NewAObj() *A {
	return &A{
		val: "A",
	}
}
func NewOverriddenAObj() *A {
	return &A{
		val: "OverriddenA",
	}
}
func (a *A) MethodA() string {
	return a.val
}

type OverrideA struct {
}

func NewOverrideA() *OverrideA {
	return &OverrideA{}
}
func (o *OverrideA) NewA() *A {
	return NewOverriddenAObj()
}

func (o *OverrideA) MethodA() string {
	return "OverrideA"
}

func OverrideAInterfaceOption[T InterfaceA](m *Module) *Module {
	return m.SetOverriddenProvider(
		"InterfaceA", func(impl T) InterfaceA {
			return impl
		},
	)
}

type FactoryA interface {
	NewA() *A
}

func OverrideAOption[T FactoryA](m *Module) *Module {
	return m.SetOverriddenProvider(
		"A", T.NewA,
	)
}
