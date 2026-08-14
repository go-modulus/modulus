package logger

import (
	"log/slog"

	"braces.dev/errtrace"
	slogmulti "github.com/samber/slog-multi"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(config ModuleConfig) (*zap.Logger, error) {
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

func NewSlog(
	zapLogger *zap.Logger,
	pipeline *Pipeline,
) *slog.Logger {
	handler := slogzap.Option{Logger: zapLogger.WithOptions(zap.AddCallerSkip(8))}.NewZapHandler()

	logger := slog.New(
		slogmulti.Pipe(pipeline.GetMiddlewares()...).Handler(handler),
	)

	return logger
}
