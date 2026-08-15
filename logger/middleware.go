package logger

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"braces.dev/errtrace"
	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
)

// Middleware transforms one slog.Handler into another. It is an alias of
// slogmulti.Middleware so logger middlewares can be plugged directly into the
// slogmulti pipeline built by NewSlog.
type Middleware = slogmulti.Middleware

type PipelineFactory interface {
	New() *Pipeline
}

type MiddlewareFactory interface {
	SlogMiddleware() Middleware
}

// NewDefaultPipeline builds the default, ranked set of logger middlewares.
func NewDefaultPipeline() *Pipeline {
	return &Pipeline{
		middlewares: map[int][]Middleware{
			100: {
				slogmulti.NewHandleInlineMiddleware(Tags),
			},
			200: {
				slogmulti.NewHandleInlineMiddleware(formatActivityError),
			},
			300: {
				slogformatter.NewFormatterMiddleware(
					slogformatter.TimeFormatter(time.RFC3339Nano, time.UTC),
				),
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

// formatActivityError rewrites the "Error" attribute of a Temporal
// "Activity error." log record into a formatted stack trace produced by
// errtrace.
//
// https://github.com/temporalio/sdk-go/blob/7fc12d37fe7fde6dcab6dfb4e0763db82b9991df/internal/internal_task_handlers.go#L2118
func formatActivityError(
	ctx context.Context,
	record slog.Record,
	next func(context.Context, slog.Record) error,
) error {
	if record.Message == "Activity error." {
		// TODO: test it
		record.Attrs(
			func(attr slog.Attr) bool {
				if attr.Key == "Error" {
					err, ok := attr.Value.Any().(error)
					if ok {
						attr.Value = slog.StringValue(errtrace.FormatString(err))
					}
					return false
				}
				return true
			},
		)
	}
	return next(ctx, record)
}
