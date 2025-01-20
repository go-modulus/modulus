package graphql

import (
	"github.com/go-modulus/modulus/http"
	oHttp "net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

type Handler struct {
	config  Config
	handler *handler.Server
}

func NewHandler(
	config Config,
	handler *handler.Server,
) *Handler {
	return &Handler{
		config:  config,
		handler: handler,
	}
}

func NewHandlerGetRoute(handler *Handler) http.RouteProvider {
	return http.NewRouteFromHandler(oHttp.MethodGet, handler.config.Path, handler.Handle)
}

func NewHandlerPostRoute(handler *Handler) http.RouteProvider {
	return http.NewRouteFromHandler(oHttp.MethodPost, handler.config.Path, handler.Handle)
}

func (h *Handler) Handle(w oHttp.ResponseWriter, req *oHttp.Request) error {
	h.handler.ServeHTTP(w, req)

	return nil
}

type PlaygroundHandler struct {
	config  Config
	handler *handler.Server
}

func NewPlaygroundHandler(config Config, handler *handler.Server) *PlaygroundHandler {
	return &PlaygroundHandler{config: config, handler: handler}
}

func NewPlaygroundHandlerRoute(handler *PlaygroundHandler) http.RouteProvider {
	if handler.config.Playground.Enabled {
		return http.NewRouteFromHandler(oHttp.MethodGet, handler.config.Playground.Path, handler.Handle)
	}

	return http.RouteProvider{}
}

func (h *PlaygroundHandler) Register(routes *http.Routes) {
	if h.config.Playground.Enabled {
		routes.Get(h.config.Playground.Path, h.Handle)
	}
}

func (h *PlaygroundHandler) Handle(w oHttp.ResponseWriter, req *oHttp.Request) error {
	playground.Handler("Graphql Playground", h.config.Path).ServeHTTP(w, req)

	return nil
}
