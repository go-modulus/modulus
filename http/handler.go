package http

import (
	"go.uber.org/fx"
)

type HandlerRegistrar interface {
	Register(routes *Routes)
}

func Provide[T HandlerRegistrar](register interface{}) fx.Option {
	return fx.Provide(
		register,
		fx.Annotate(
			func(a T) T { return a },
			fx.As(new(HandlerRegistrar)),
			fx.ResultTags(`group:"http.handlerRegistrars"`),
		),
	)
}
