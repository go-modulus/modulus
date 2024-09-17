package http

import (
	"context"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/go-chi/chi/v5"
	infraCli "github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/errors/errhttp"
	"github.com/sethvargo/go-envconfig"
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

func NewServeConfig() (*ServeConfig, error) {
	config := ServeConfig{}
	return &config, envconfig.Process(context.Background(), &config)
}

type Serve struct {
	runner     *infraCli.Runner
	router     chi.Router
	registrars []HandlerRegistrar `group:"http.handlerRegistrars"`
	logger     *slog.Logger
	config     *ServeConfig
}

type ServeParams struct {
	fx.In

	Runner     *infraCli.Runner
	Router     chi.Router
	Registrars []HandlerRegistrar `group:"http.handlerRegistrars"`
	Logger     *slog.Logger
	Config     *ServeConfig
}

func NewServe(params ServeParams) *Serve {
	return &Serve{
		runner:     params.Runner,
		router:     params.Router,
		registrars: params.Registrars,
		logger:     params.Logger,
		config:     params.Config,
	}
}

func (s *Serve) Command() *cli.Command {
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

	routes := NewRoutes()
	for _, registrar := range s.registrars {
		registrar.Register(routes)
	}
	for _, route := range routes.List() {
		logger.Info("registering route", slog.String("method", route.Method), slog.String("path", route.Path))
		s.router.Method(route.Method, route.Path, errhttp.WrapHandler(logger, route.Handler))
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