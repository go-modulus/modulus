package pgx

import (
	"braces.dev/errtrace"
	"context"
	"fmt"
	"github.com/go-modulus/modulus/module"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	_ "golang.org/x/text/message"
	"log/slog"
)

type ModuleConfig struct {
	DSN string `env:"PGX_DSN"`

	Host     string `env:"PG_HOST, default=localhost"`
	Port     int    `env:"PG_PORT, default=5432"`
	User     string `env:"PG_USER, default=postgres"`
	Password string `env:"PG_PASSWORD, default=foobar"`
	Database string `env:"PG_DB_NAME, default=trustypay"`
	SslMode  string `env:"PG_SSL_MODE, default=disable"`
}

func (c ModuleConfig) Dsn() string {
	if c.DSN != "" {
		return c.DSN
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SslMode,
	)
}

func NewPgxPool(
	logger *slog.Logger,
	config *ModuleConfig,
) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(config.Dsn())
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger: tracelog.LoggerFunc(
			func(
				ctx context.Context,
				level tracelog.LogLevel,
				msg string,
				data map[string]any,
			) {
				attrs := make([]slog.Attr, 0, len(data))
				for k, v := range data {
					attrs = append(attrs, slog.Any(k, v))
				}

				var lvl slog.Level
				switch level {
				case tracelog.LogLevelTrace:
					lvl = slog.LevelDebug - 1
					attrs = append(attrs, slog.Any("PGX_LOG_LEVEL", level))
				case tracelog.LogLevelDebug:
					lvl = slog.LevelDebug
				case tracelog.LogLevelInfo:
					lvl = slog.LevelInfo
				case tracelog.LogLevelWarn:
					lvl = slog.LevelWarn
				case tracelog.LogLevelError:
					return
				default:
					lvl = slog.LevelError
					attrs = append(attrs, slog.Any("INVALID_PGX_LOG_LEVEL", level))
				}
				logger.LogAttrs(ctx, lvl, msg, attrs...)
			},
		),
		LogLevel: tracelog.LogLevelTrace,
	}

	return errtrace.Wrap2(pgxpool.NewWithConfig(context.Background(), cfg))
}

func NewModule(config ModuleConfig) *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/db/pgx").
		AddProviders(
			NewPgxPool,
			module.ConfigProvider[ModuleConfig](config),
		)
}