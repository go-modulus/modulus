package test

import (
	"context"
	"github.com/go-modulus/modulus/config"
	"os"
	"path"
	"runtime"
	"sync"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var initOnce sync.Once

func TestMain(m *testing.M, options ...fx.Option) {
	initOnce.Do(
		func() {
			// force UTC timezone, otherwise it will use local timezone
			// on unmarshalling/marshalling time from/to postgres
			os.Setenv("TZ", "UTC")

			_, filename, _, _ := runtime.Caller(0)
			dir := path.Join(path.Dir(filename), "../../..") + "/"
			config.LoadEnv(dir, "", false)
			config.LoadEnv(dir, "test", true)
		},
	)

	_ = Invoke(options...)
	os.Exit(m.Run())
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
