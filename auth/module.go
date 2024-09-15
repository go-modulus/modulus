package auth

import (
	"go.uber.org/fx"
)

func NewModule() fx.Option {
	return fx.Module(
		"modulus/auth",
		fx.Provide(
			NewMiddlewareConfig,
			NewMiddleware,
		),
	)
}
