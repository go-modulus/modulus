package graphql

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	config2 "github.com/go-modulus/modulus/config"
	"github.com/go-modulus/modulus/errors/errlog"
	httpContext "github.com/go-modulus/modulus/http/context"
	context2 "github.com/go-modulus/modulus/translation"
	"github.com/vektah/gqlparser/v2/ast"
	"log/slog"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	infraErrors "github.com/go-modulus/modulus/errors"
	"github.com/ravilushqa/otelgqlgen"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type PlaygroundConfig struct {
	Enabled bool   `env:"GQL_PLAYGROUND_ENABLED, default=true"`
	Path    string `env:"GQL_PLAYGROUND_URL, default=/playground"`
}

type Config struct {
	ComplexityLimit      int    `env:"GQL_COMPLEXITY_LIMIT, default=200"`
	Path                 string `env:"GQL_API_URL, default=/graphql"`
	IntrospectionEnabled bool   `env:"GQL_INTROSPECTION_ENABLED, default=true"`
	TracingEnabled       bool   `env:"GQL_TRACING_ENABLED, default=false"`
	Playground           PlaygroundConfig
}

type UserError interface {
	ToUserError() map[string]interface{}
}

func NewGraphqlServer(
	config Config,
	schema graphql.ExecutableSchema,
	loadersInitializer *LoadersInitializer,
	logger *slog.Logger,
) *handler.Server {
	var mb int64 = 1 << 20

	srv := handler.New(schema)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(
		transport.MultipartForm{
			MaxUploadSize: mb * 5,
			MaxMemory:     mb * 5,
		},
	)
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](1000)})
	if config.IntrospectionEnabled {
		srv.Use(extension.Introspection{})
	}

	srv.Use(extension.FixedComplexityLimit(config.ComplexityLimit))
	srv.Use(loadersInitializer)
	srv.Use(otelgqlgen.Middleware())

	if config.TracingEnabled {
		srv.Use(apollotracing.Tracer{})
	}

	srv.SetRecoverFunc(
		func(ctx context.Context, p any) error {
			return fmt.Errorf("panic: %v", p)
		},
	)

	srv.SetErrorPresenter(
		func(ctx context.Context, err error) *gqlerror.Error {
			var gqlErr *gqlerror.Error
			path := graphql.GetPath(ctx)
			if errors.As(err, &gqlErr) {
				if gqlErr.Path == nil {
					gqlErr.Path = path
				} else {
					path = gqlErr.Path
				}

				originalErr := gqlErr.Unwrap()
				if originalErr == nil {
					return gqlErr
				}

				err = originalErr
			}

			code := err.Error()
			message := infraErrors.Message(context2.GetPrinter(ctx), err)
			extra := make(map[string]any)
			meta := infraErrors.Meta(err)
			if meta != nil && !config2.IsProd() {
				extra["meta"] = meta
			}
			extra["code"] = code
			requestID := httpContext.GetRequestID(ctx)
			if requestID != "" {
				extra["requestId"] = requestID
			}

			level, logged := errlog.LogError(ctx, err, logger)
			if logged && level == slog.LevelError {
				message = fmt.Sprintf("%s (RID: %s)", message, requestID)
			}

			return &gqlerror.Error{
				Message:    message,
				Path:       path,
				Extensions: extra,
			}
		},
	)

	return srv
}
