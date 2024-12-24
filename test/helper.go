package test

import (
	"context"
	"github.com/go-modulus/modulus/config"
	"os"
	"sync"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var initOnce sync.Once

func TestMain(m *testing.M, options ...fx.Option) {
	_ = Invoke(options...)
	os.Exit(m.Run())
}

func LoadEnv(envFileDir string) {
	initOnce.Do(
		func() {
			// force UTC timezone, otherwise it will use local timezone
			// on unmarshalling/marshalling time from/to postgres
			os.Setenv("TZ", "UTC")
			os.Setenv("APP_ENV", "test")

			config.LoadEnv(envFileDir, "", false)
			config.LoadEnv(envFileDir, "test", true)
		},
	)
}

func Invoke(options ...fx.Option) error {
	opts := []fx.Option{
		fx.WithLogger(
			func() fxevent.Logger {
				cfg := zap.NewDevelopmentConfig()
				cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
				cfg.DisableCaller = true
				logger, _ := cfg.Build()
				return &fxevent.ZapLogger{Logger: logger}
			},
		),
	}
	app := fx.New(append(opts, options...)...)

	return app.Start(context.Background())
}
