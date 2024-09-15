package graphql

import (
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/http"
	oHttp "net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

type Handler struct {
	config        *Config
	handler       *handler.Server
	authenticator auth.Authenticator
}

func NewHandler(
	config *Config,
	handler *handler.Server,
	authenticator auth.Authenticator,
) *Handler {
	return &Handler{
		config:        config,
		handler:       handler,
		authenticator: authenticator,
	}
}

func (h *Handler) Register(routes *http.Routes) {
	routes.Get(h.config.Path, h.Handle)
	routes.Post(h.config.Path, h.Handle)
}

func (h *Handler) Handle(w oHttp.ResponseWriter, req *oHttp.Request) error {
	h.handler.ServeHTTP(w, req)

	return nil
}

type PlaygroundHandler struct {
	config  *Config
	handler *handler.Server
}

func NewPlaygroundHandler(config *Config, handler *handler.Server) *PlaygroundHandler {
	return &PlaygroundHandler{config: config, handler: handler}
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
