package logger

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/errlog"
	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
	slogzap "github.com/samber/slog-zap/v2"
	"github.com/sethvargo/go-envconfig"
	"log/slog"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ModuleConfig struct {
	Level string `env:"LOGGER_LEVEL, default=debug"`
	Type  string `env:"LOGGER_TYPE, default=json"`
	App   string `env:"LOGGER_APP, default=trustypay"`
}

func NewLogger(config *ModuleConfig) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		return nil, errtrace.Errorf(
			`invalid logger level "%s". Use "debug", "info", "warn" or "error"`,
			config.Level,
		)
	}
	if config.Type != "json" && config.Type != "console" {
		return nil, errtrace.Errorf(
			`invalid logger type "%s". Use "json" or "console"`,
			config.Type,
		)
	}

	cfg := zap.NewProductionConfig()
	cfg.Sampling = nil
	cfg.Level = level
	cfg.Encoding = config.Type
	if config.Type == "console" {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	cfg.InitialFields = map[string]interface{}{
		"app": config.App,
	}
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.DisableStacktrace = true

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(logger)
	//nolint:errcheck
	_, _ = zap.RedirectStdLogAt(logger, zapcore.WarnLevel)

	return logger, nil
}

func NewModule(config ModuleConfig) fx.Option {
	return fx.Module(
		"logger",
		fx.Provide(
			NewLogger,
			func(
				zapLogger *zap.Logger,
			) *slog.Logger {
				handler := slogzap.Option{Logger: zapLogger.WithOptions(zap.AddCallerSkip(8))}.NewZapHandler()
				errorFormattingMiddleware := slogformatter.NewFormatterMiddleware(
					slogformatter.TimeFormatter(time.RFC3339Nano, time.UTC),
					errlog.Formatter(),
				)
				logger := slog.New(
					slogmulti.
						Pipe(slogmulti.NewHandleInlineMiddleware(Tags)).
						Pipe(
							slogmulti.NewHandleInlineMiddleware(
								func(
									ctx context.Context,
									record slog.Record,
									next func(context.Context, slog.Record) error,
								) error {
									// https://github.com/temporalio/sdk-go/blob/7fc12d37fe7fde6dcab6dfb4e0763db82b9991df/internal/internal_task_handlers.go#L2118
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
								},
							),
						).
						Pipe(errorFormattingMiddleware).
						Handler(handler),
				)

				return logger
			},

			func() (*ModuleConfig, error) {
				return &config, envconfig.Process(context.Background(), &config)
			},
		),
	)
}
