package http

import (
	"context"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/go-chi/chi/v5"
	infraCli "github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/urfave/cli/v2"
	"log/slog"
	netHttp "net/http"
	"time"

	"go.uber.org/fx"
)

type ServeConfig struct {
	Address          string            `env:"HTTP_HOST, default=localhost:8001"`
	TTL              time.Duration     `env:"ROUTER_TTL, default=15s"` // 15 seconds
	RequestSizeLimit datasize.ByteSize `env:"ROUTER_REQUEST_SIZE_LIMIT, default=5mb"`
}

type Serve struct {
	runner        *infraCli.Runner
	router        chi.Router
	registrars    []HandlerRegistrar `group:"http.handlerRegistrars"`
	routes        []Route
	middlewares   []Middleware
	errorPipeline *errhttp.ErrorPipeline
	logger        *slog.Logger
	config        ServeConfig
}

type ServeParams struct {
	fx.In

	Runner     *infraCli.Runner
	Router     chi.Router
	Registrars []HandlerRegistrar `group:"http.handlerRegistrars"`
	Routes     []Route            `group:"http.routes"`
	Pipeline   *Pipeline
	// @todo: think on placing this in each route to be able to override it for specific routes
	ErrorPipeline *errhttp.ErrorPipeline
	Logger        *slog.Logger
	Config        ServeConfig
}

func NewServe(params ServeParams) *Serve {
	middlewares := make([]Middleware, 0)
	if params.Pipeline != nil {
		middlewares = params.Pipeline.Middlewares
	}
	return &Serve{
		runner:        params.Runner,
		router:        params.Router,
		registrars:    params.Registrars,
		routes:        params.Routes,
		logger:        params.Logger,
		config:        params.Config,
		middlewares:   middlewares,
		errorPipeline: params.ErrorPipeline,
	}
}

func NewServeCommand(s *Serve) *cli.Command {
	return &cli.Command{
		Name:   "serve",
		Action: s.Invoke,
	}
}

func (s *Serve) Invoke(cliCtx *cli.Context) error {
	ctx := cliCtx.Context

	logger := s.logger.With(slog.String("component", "http"))

	server := &netHttp.Server{
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         s.config.Address,
		Handler:      s.router,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	if len(s.middlewares) > 0 {
		for _, middleware := range s.middlewares {
			s.router.Use(middleware)
		}

		logger.Info("registering global Middlewares", slog.Int("count", len(s.middlewares)))
	}

	routes := NewRoutes()
	for _, registrar := range s.registrars {
		registrar.Register(routes)
	}
	for _, route := range s.routes {
		if route.Handler == nil || route.Path == "" {
			continue
		}
		routes.Add(route)
	}
	for _, route := range routes.List() {
		logger.Info("registering route", slog.String("method", route.Method), slog.String("path", route.Path))
		s.router.Method(route.Method, route.Path, errhttp.WrapHandler(s.errorPipeline, route.Handler))
	}

	return s.runner.Run(
		ctx, func(ctx context.Context) error {
			errChannel := make(chan error)
			go func() {
				logger.Info("http server is starting")

				err := server.ListenAndServe()
				if err != nil {
					errChannel <- err
				}
			}()

			timer := time.NewTimer(1 * time.Second)

			for {
				select {
				case <-ctx.Done():
					logger.Info("http server is stopping")

					err := server.Shutdown(ctx)
					if err != nil {
						return err
					}

					logger.Info("http server has stopped")
					return nil
				case err := <-errChannel:
					if err == netHttp.ErrServerClosed {
						return nil
					}

					return fmt.Errorf("http server has failed to run: %w", err)
				case <-timer.C:
					logger.Info(
						"http server has started",
						slog.String("address", s.config.Address),
					)
				}
			}
		},
	)
}
