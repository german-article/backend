package tracer

import (
	"context"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	tr trace.Tracer
	tp *sdktrace.TracerProvider
}

func New(tp *sdktrace.TracerProvider, tr trace.Tracer) *Tracer {
	return &Tracer{tr: tr, tp: tp}
}

func Init(ctx context.Context, projectID, applicationName string, gcp bool) (tr *Tracer, err error) {
	var (
		traceProvider *sdktrace.TracerProvider
		tracer        trace.Tracer
	)

	if gcp {
		exporter, err := texporter.New(texporter.WithProjectID(projectID))
		if err != nil {
			return tr, err
		}

		res, err := resource.New(ctx,
			// Use the GCP resource detector to detect information about the GCP platform
			//resource.WithDetectors(gcp.NewDetector()),
			// Keep the default detectors
			//resource.WithTelemetrySDK(),
			// Add your own custom attributes to identify your application
			resource.WithAttributes(semconv.ServiceNameKey.String(applicationName)),
		)
		if err != nil {
			return tr, err
		}

		traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)
	} else {
		//exporter, err := jaeger.New(jaeger.WithAgentEndpoint())
		exporter, err := otlptracegrpc.New(ctx)
		if err != nil {
			return tr, err
		}
		res, err := resource.New(ctx,
			resource.WithAttributes(semconv.ServiceNameKey.String(applicationName)),
		)
		if err != nil {
			return tr, err
		}
		traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)
	}

	otel.SetTracerProvider(traceProvider)
	tracer = otel.GetTracerProvider().Tracer(applicationName)
	return New(traceProvider, tracer), nil
}

func (t *Tracer) Close(ctx context.Context) error {
	if t.tp == nil {
		return nil
	}

	return t.tp.Shutdown(ctx)
}

func (t *Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tr.Start(ctx, spanName, opts...)
}
