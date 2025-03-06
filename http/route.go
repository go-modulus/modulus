package http

import (
	"github.com/go-modulus/modulus/http/errhttp"
	"go.uber.org/fx"
	"net/http"
)

type Route struct {
	Method  string
	Path    string
	Handler errhttp.Handler
}

type RouteProvider struct {
	fx.Out
	Route Route `group:"http.routes"`
}

func NewRouteProvider[B any](method, path string, handler InputHandler[B]) RouteProvider {
	h := WrapInputHandler(handler)
	return RouteProvider{
		Route: Route{
			Method:  method,
			Path:    path,
			Handler: h,
		},
	}
}

func NewRouteFromHandler(method, path string, handler errhttp.Handler) RouteProvider {
	return RouteProvider{
		Route: Route{
			Method:  method,
			Path:    path,
			Handler: handler,
		},
	}
}

type Routes struct {
	routes []Route
}

func NewRoutes() *Routes {
	return &Routes{routes: make([]Route, 0)}
}

func (r *Routes) Get(path string, handler errhttp.Handler) {
	r.routes = append(
		r.routes,
		Route{
			Method:  http.MethodGet,
			Path:    path,
			Handler: handler,
		},
	)
}
func (r *Routes) Post(path string, handler errhttp.Handler) {
	r.routes = append(
		r.routes,
		Route{
			Method:  http.MethodPost,
			Path:    path,
			Handler: handler,
		},
	)
}
func (r *Routes) Put(path string, handler errhttp.Handler) {
	r.routes = append(
		r.routes,
		Route{
			Method:  http.MethodPut,
			Path:    path,
			Handler: handler,
		},
	)
}
func (r *Routes) Patch(path string, handler errhttp.Handler) {
	r.routes = append(
		r.routes,
		Route{
			Method:  http.MethodPatch,
			Path:    path,
			Handler: handler,
		},
	)
}
func (r *Routes) Delete(path string, handler errhttp.Handler) {
	r.routes = append(
		r.routes,
		Route{
			Method:  http.MethodDelete,
			Path:    path,
			Handler: handler,
		},
	)
}
func (r *Routes) List() []Route {
	return r.routes
}
func (r *Routes) Add(route Route) {
	r.routes = append(r.routes, route)
}
