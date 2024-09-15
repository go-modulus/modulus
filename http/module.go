package http

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	infraCli "github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/errhttp"
	"github.com/go-modulus/modulus/erruser"
	"github.com/go-modulus/modulus/errwrap"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"log/slog"
	"net/http"
)

var (
	ErrMethodNotAllowed = errwrap.Wrap(
		erruser.New("MethodNotAllowed", "Method not allowed"),
		errhttp.With(http.StatusMethodNotAllowed),
	)
	ErrNotFound = errwrap.Wrap(
		erruser.New("NotFound", "Not found"),
		errhttp.With(http.StatusNotFound),
	)
)

func NewRouter(logger *slog.Logger, config *ServeConfig) chi.Router {
	r := chi.NewRouter()
	r.MethodNotAllowed(
		errhttp.WrapHandler(
			logger,
			func(w http.ResponseWriter, req *http.Request) error {
				return ErrMethodNotAllowed
			},
		),
	)
	r.NotFound(
		errhttp.WrapHandler(
			logger,
			func(w http.ResponseWriter, req *http.Request) error {
				return ErrNotFound
			},
		),
	)
	r.Use(chiMiddleware.Timeout(config.TTL))
	r.Use(chiMiddleware.RequestSize(int64(config.RequestSizeLimit.Bytes())))
	return r
}

func NewModule() fx.Option {
	return fx.Module(
		"http",
		fx.Provide(
			NewRouter,
			NewServeConfig,
			NewServe,
			infraCli.ProvideCommand(
				func(serve *Serve) *cli.Command {
					return serve.Command()
				},
			),
		),
	)
}
