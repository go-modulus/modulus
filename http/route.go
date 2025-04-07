package http

import (
	"github.com/go-modulus/modulus/http/errhttp"
	"go.uber.org/fx"
	"net/http"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.Handler
	ErrHandler errhttp.Handler
}

func (r *Route) IsEmpty() bool {
	return (r.ErrHandler == nil && r.Handler == nil) || r.Path == ""
}

type RouteProvider struct {
	fx.Out
	Route Route `group:"http.routes"`
}

func ProvideRawRoute(method, path string, handler http.Handler) RouteProvider {
	return RouteProvider{
		Route: Route{
			Method:  method,
			Path:    path,
			Handler: handler,
		},
	}
}

func ProvideInputRoute[T any](method, path string, handler InputHandler[T]) RouteProvider {
	return RouteProvider{
		Route: Route{
			Method:     method,
			Path:       path,
			ErrHandler: WrapInputHandler(handler),
		},
	}
}

func ProvideRoute(method, path string, handler errhttp.Handler) RouteProvider {
	return RouteProvider{
		Route: Route{
			Method:     method,
			Path:       path,
			ErrHandler: handler,
		},
	}
}
