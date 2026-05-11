package otel

import (
	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/go-modulus/modulus/module"
)

type Config struct {
	TracesExporter  errhttp.ErrorProcessor `env:"OTEL_TRACES_EXPORTER, default=console", comment:"See more env variables at https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables to configure the OpenTelemetry SDK."`
	MetricsExporter errhttp.ErrorProcessor `env:"OTEL_METRICS_EXPORTER, default=console"`
	LogsExporter    errhttp.ErrorProcessor `env:"OTEL_LOGS_EXPORTER, default=console"`
}

func NewModule() *module.Module {
	return module.NewModule("otel").
		AddCliCommands().
		AddProviders().
		SetOverriddenProvider("otel.LogExporter", NewLogExporter).
		SetOverriddenProvider("otel.MeterProvider", NewMeterProvider).
		SetOverriddenProvider("otel.TracerProvider", NewTracerProvider).
		InitConfig(Config{}).
		AddInvokes(invokeProviders)
}
