package auth

import (
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/auth").
		AddProviders(
			NewMiddlewareConfig,
			NewMiddleware,
		)
}
