package graphql

import (
	infraHttp "github.com/go-modulus/modulus/http"
	"go.uber.org/fx"
)

func NewModule() fx.Option {
	return fx.Module(
		"modulus/graphql",
		fx.Provide(
			NewConfig,
			NewGraphqlServer,
			NewLoadersInitializer,
		),
		infraHttp.Provide[*Handler](NewHandler),
		infraHttp.Provide[*PlaygroundHandler](NewPlaygroundHandler),
	)
}
