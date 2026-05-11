package otel

import (
	"context"

	"github.com/go-modulus/modulus/errors"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	otelMetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

func NewLogExporter(lc fx.Lifecycle) log.Exporter {
	var exporter log.Exporter
	var err error
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				exporter, err = autoexport.NewLogExporter(ctx)
				if err != nil {
					return errors.NewWithCause("failed to create log exporter", err)
				}
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if exporter == nil {
					return nil
				}
				return exporter.Shutdown(ctx)
			},
		},
	)

	return exporter
}

func NewMeterProvider(lc fx.Lifecycle) otelMetric.MeterProvider {
	var meterProvider *metric.MeterProvider
	var exporter metric.Reader
	var err error
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				exporter, err = autoexport.NewMetricReader(ctx)
				if err != nil {
					return errors.NewWithCause("failed to create metric exporter", err)
				}
				otelResource, err2 := resource.New(
					ctx,
					resource.WithFromEnv(),
					resource.WithTelemetrySDK(),
					resource.WithHost(),
				)

				if err2 != nil {
					return errors.NewWithCause("failed to create resource", err2)
				}
				meterProvider = metric.NewMeterProvider(
					metric.WithReader(exporter),
					metric.WithResource(otelResource),
				)

				otel.SetMeterProvider(meterProvider)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if exporter == nil {
					return nil
				}
				return exporter.Shutdown(ctx)
			},
		},
	)

	return meterProvider
}

func NewTracerProvider(lc fx.Lifecycle) otelTrace.TracerProvider {
	var tracerProvider *trace.TracerProvider
	var exporter trace.SpanExporter
	var err error
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				exporter, err = autoexport.NewSpanExporter(ctx)
				if err != nil {
					return errors.NewWithCause("failed to create metric exporter", err)
				}
				otelResource, err2 := resource.New(
					ctx,
					resource.WithFromEnv(),
					resource.WithTelemetrySDK(),
					resource.WithHost(),
				)

				if err2 != nil {
					return errors.NewWithCause("failed to create resource", err2)
				}
				tracerProvider = trace.NewTracerProvider(
					trace.WithBatcher(exporter),
					trace.WithResource(otelResource),
				)

				otel.SetTracerProvider(tracerProvider)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if exporter == nil {
					return nil
				}
				return exporter.Shutdown(ctx)
			},
		},
	)

	return tracerProvider
}

func invokeProviders(
	meterProvider otelMetric.MeterProvider,
	tracerProvider otelTrace.TracerProvider,
) {
	//invoke meter provider to initialize
}
