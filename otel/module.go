package otel

import (
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
)

type Config struct {
	AppEnv      string `env:"APP_ENV, default=dev"`
	ServiceName string `env:"OTEL_SERVICE_NAME, default=server-app"`
	LogViaOtel  bool   `env:"OTEL_LOG_VIA_OTEL, default=false" comment:"If true, logs will be exported via OpenTelemetry. If false, logs will be exported via standard output."`
}

func NewModule() *module.Module {
	return module.NewModule("otel").
		AddDependencies(
			logger.NewModule(),
		).
		AddCliCommands().
		AddProviders(
			NewLogMiddlewareFactory,
		).
		SetOverriddenProvider("otel.LogProvider", NewLogProvider).
		SetOverriddenProvider("otel.MeterProvider", NewMeterProvider).
		SetOverriddenProvider("otel.TracerProvider", NewTracerProvider).
		SetOverriddenProvider("otel.Resource", NewResource).
		InitConfig(Config{}).
		AddInvokes(invokeProviders)
}

//
//// invokeLogMiddleware registers NewLogMiddleware on the modulus logger
//// pipeline so every log record is also emitted via OpenTelemetry, but only
//// when logging via OpenTelemetry is turned on in the config.
//func invokeLogMiddleware(pipeline *logger.Pipeline, config Config, provider *log.LoggerProvider) {
//	if !config.LogViaOtel {
//		return
//	}
//	pipeline.SetMiddleware(900, NewLogMiddleware(config.ServiceName, provider))
//}
