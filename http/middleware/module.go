package middleware

import (
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("http middleware").
		AddProviders(
			NewCors,
		).
		InitConfig(CorsConfig{})
}
