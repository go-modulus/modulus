package validator_test

import (
	"context"
	"testing"

	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func TestValidateWithContext(t *testing.T) {
	t.Parallel()
	t.Run(
		"Success map validation", func(t *testing.T) {
			m := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			keyRules := make([]*validation.KeyRules, 0)
			keyRules = append(keyRules, validation.Key("key1", validation.Required.Error("Test error")))
			keyRules = append(keyRules, validation.Key("key2", validation.Required.Error("Test error 2")))
			mapRule := validation.Map(keyRules...)
			err := validator.ValidateWithContext(context.Background(), &m, "", mapRule)

			assert.Nil(t, err)
		},
	)

	t.Run(
		"Fail map validation", func(t *testing.T) {
			m := map[string]string{
				"key1": "",
				"key2": "",
			}
			keyRules := make([]*validation.KeyRules, 0)
			keyRules = append(keyRules, validation.Key("key1", validation.Required.Error("Test error")))
			keyRules = append(keyRules, validation.Key("key2", validation.Required.Error("Test error 2")))
			mapRule := validation.Map(keyRules...)
			err := validator.ValidateWithContext(context.Background(), &m, "", mapRule)

			assert.Error(t, err)
			assert.Equal(t, "invalid input", err.Error())
			assert.Equal(t, "Test error", errors.Hint(err))
		},
	)
}
