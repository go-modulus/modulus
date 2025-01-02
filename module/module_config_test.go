package module_test

import (
	"github.com/go-modulus/modulus/module"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetEnvVariablesFromConfig(t *testing.T) {
	t.Run(
		"get variables with default values", func(t *testing.T) {
			type s struct {
				A string `env:"A, default=a"`
			}

			vars := module.GetEnvVariablesFromConfig[s](s{})

			assert.Len(t, vars, 1)
			assert.Equal(t, "A", vars[0].Key)
			assert.Equal(t, "a", vars[0].Value)
		},
	)

	t.Run(
		"get variables with comment", func(t *testing.T) {
			type s struct {
				A string `env:"A, default=a" comment:"This is a comment"`
			}

			vars := module.GetEnvVariablesFromConfig[s](s{})

			assert.Len(t, vars, 1)
			assert.Equal(t, "A", vars[0].Key)
			assert.Equal(t, "a", vars[0].Value)
			assert.Equal(t, "This is a comment", vars[0].Comment)
		},
	)
}
