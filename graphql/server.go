package graphql

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/fx"
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
	ReturnCause          bool   `env:"GQL_RETURN_CAUSE, default=false"`
	Playground           PlaygroundConfig
}

type ErrorPresenterParams struct {
	fx.In

	ErrorPipeline *errhttp.ErrorPipeline `optional:"true"`
	Config        Config
}

type ServerParams struct {
	fx.In

	Config             Config
	Schema             graphql.ExecutableSchema
	LoadersInitializer *LoadersInitializer `optional:"true"`
	Logger             *slog.Logger
	ErrorPresenter     graphql.ErrorPresenterFunc
}

func NewGraphqlServer(
	params ServerParams,
) *handler.Server {
	var mb int64 = 1 << 20

	config := params.Config
	srv := handler.New(params.Schema)

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
	if params.LoadersInitializer != nil {
		srv.Use(params.LoadersInitializer)
	}
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
		params.ErrorPresenter,
	)

	return srv
}

func NewErrorPresenter(params ErrorPresenterParams) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
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

		config := params.Config

		if params.ErrorPipeline != nil {
			for _, converter := range params.ErrorPipeline.Processors {
				err = converter(ctx, err)
			}
		}

		code := err.Error()
		message := infraErrors.Hint(err)
		if message == "" {
			message = code
		}

		extra := make(map[string]any)

		meta := infraErrors.Meta(err)
		if meta != nil {
			extra["meta"] = infraErrors.Meta(err)
		}

		if config.ReturnCause {
			cause := infraErrors.Cause(err)
			if cause != nil {
				causeMap := map[string]interface{}{
					"code": cause.Error(),
				}
				hint := infraErrors.Hint(cause)
				if hint != "" {
					causeMap["message"] = hint
				}
				metaCause := infraErrors.Meta(cause)
				if metaCause != nil {
					causeMap["meta"] = infraErrors.Meta(cause)
				}

				extra["cause"] = causeMap
			}
		}

		extra["code"] = code

		return &gqlerror.Error{
			Message:    message,
			Path:       path,
			Extensions: extra,
		}
	}
}
