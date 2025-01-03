package graphql

import (
	infraHttp "github.com/go-modulus/modulus/http"
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/graphql").
		AddDependencies(
			infraHttp.NewModule(),
		).
		AddProviders(
			NewConfig,
			NewGraphqlServer,
			NewLoadersInitializer,
			infraHttp.Provide[*Handler](NewHandler),
			infraHttp.Provide[*PlaygroundHandler](NewPlaygroundHandler),
		)
}
