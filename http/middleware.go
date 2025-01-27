package http

import (
	netHttp "net/http"
)

type Middleware func(handler netHttp.Handler) netHttp.Handler

type Pipeline struct {
	Middlewares []Middleware
}
