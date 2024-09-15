package middleware

import (
	"go.uber.org/fx"
)

func NewModule() fx.Option {
	return fx.Module(
		"http/middleware",
		fx.Provide(
			NewCorsConfig,
			NewCors,
		),
	)
}
