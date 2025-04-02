package graphql

import (
	modulusHttp "github.com/go-modulus/modulus/http"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

func NewHandlerRoute(handler *handler.Server, config Config) (modulusHttp.RouteProvider, modulusHttp.RouteProvider) {
	return modulusHttp.ProvideRawRoute(http.MethodGet, config.Path, handler),
		modulusHttp.ProvideRawRoute(http.MethodPost, config.Path, handler)
}

func NewPlaygroundHandlerRoute(config Config) modulusHttp.RouteProvider {
	if config.Playground.Enabled {
		return modulusHttp.ProvideRawRoute(
			http.MethodGet,
			config.Playground.Path,
			playground.Handler("Graphql Playground", config.Path),
		)
	}

	return modulusHttp.RouteProvider{}
}
