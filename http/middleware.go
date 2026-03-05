package http

import (
	netHttp "net/http"
	"sort"

	"github.com/go-modulus/modulus/http/middleware"
)

type Middleware func(handler netHttp.Handler) netHttp.Handler

type PipelineFactory interface {
	New() *Pipeline
}

func NewDefaultPipeline() *Pipeline {
	return &Pipeline{
		middlewares: map[int][]Middleware{
			100: {
				middleware.RequestID,
			},
			200: {
				middleware.IP,
			},
			300: {
				middleware.UserAgent,
			},
		},
	}
}

type Pipeline struct {
	// middlewares is a map of ranked Middleware functions executed in rank order.
	middlewares map[int][]Middleware
	cache       []Middleware
}

// SetMiddleware appends a middleware at the given rank. Middlewares with a
// lower rank are executed first. Multiple middlewares at the same rank are
// executed in insertion order.
func (p *Pipeline) SetMiddleware(rank int, middleware Middleware) {
	if p.middlewares == nil {
		p.middlewares = make(map[int][]Middleware)
	}
	p.middlewares[rank] = append(p.middlewares[rank], middleware)
	p.cache = nil
}

// GetMiddlewares returns the flat, rank-sorted slice of middlewares,
// rebuilding it from the ranked map when the cache is empty.
func (p *Pipeline) GetMiddlewares() []Middleware {
	if len(p.middlewares) == 0 {
		return nil
	}
	middlewares := p.cache
	if len(middlewares) == 0 {
		middlewares = p.getMiddlewaresList()
		p.cache = middlewares
	}
	return middlewares
}

func (p *Pipeline) getMiddlewaresList() []Middleware {
	result := make([]Middleware, 0, len(p.middlewares))

	var ranks []int
	for rank := range p.middlewares {
		ranks = append(ranks, rank)
	}
	sort.Ints(ranks)

	for _, rank := range ranks {
		result = append(result, p.middlewares[rank]...)
	}
	return result
}
