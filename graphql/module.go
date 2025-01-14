package graphql

import (
	mHttp "github.com/go-modulus/modulus/http"
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("gqlgen").
		AddDependencies(
			mHttp.NewModule(),
		).
		AddProviders(
			NewGraphqlServer,
			NewLoadersInitializer,
			NewHandler,
			NewPlaygroundHandler,
			NewHandlerGetRoute,
			NewHandlerPostRoute,
			NewPlaygroundHandlerRoute,
		).InitConfig(Config{})
}
