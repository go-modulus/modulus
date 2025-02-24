package module

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModule_InitConfig(t *testing.T) {
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
