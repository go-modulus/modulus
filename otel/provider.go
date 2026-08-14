package otel

import (
	"context"

	"github.com/go-modulus/modulus/errors"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	otelMetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

func NewResource(lc fx.Lifecycle, config Config) *resource.Resource {
	var res *resource.Resource
	var err error
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				res, err = resource.New(
					ctx,
					resource.WithFromEnv(),
					resource.WithTelemetrySDK(),
					resource.WithHost(),
					resource.WithAttributes(
						semconv.DeploymentEnvironmentKey.String(config.AppEnv),
						semconv.ServiceNameKey.String(config.ServiceName),
					),
				)
				if err != nil {
					return errors.NewWithCause("failed to create resource", err)
				}
				return nil
			},
		},
	)

	return res
}

func NewLogProvider(lc fx.Lifecycle, res *resource.Resource) (*log.LoggerProvider, error) {
	ctx := context.Background()
	exporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return nil, errors.NewWithCause("failed to create log exporter", err)
	}
	lp := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)
	lc.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				if exporter == nil {
					return nil
				}
				return exporter.Shutdown(ctx)
			},
		},
	)

	return lp, nil
}

func NewMeterProvider(lc fx.Lifecycle, res *resource.Resource) (otelMetric.MeterProvider, error) {
	ctx := context.Background()
	exporter, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, errors.NewWithCause("failed to create metric exporter", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)
	lc.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				if exporter == nil {
					return nil
				}
				return exporter.Shutdown(ctx)
			},
		},
	)

	return meterProvider, nil
}

func NewTracerProvider(lc fx.Lifecycle, res *resource.Resource) (otelTrace.TracerProvider, error) {
	ctx := context.Background()
	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, errors.NewWithCause("failed to create metric exporter", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	lc.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				if exporter == nil {
					return nil
				}
				return exporter.Shutdown(ctx)
			},
		},
	)

	return tracerProvider, nil
}

func invokeProviders(
	meterProvider otelMetric.MeterProvider,
	tracerProvider otelTrace.TracerProvider,
	logProvider *log.LoggerProvider,
) {
	otel.SetMeterProvider(meterProvider)
	otel.SetTracerProvider(tracerProvider)
	global.SetLoggerProvider(logProvider)
}
