package tracing

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

// Tracer defines the tracing interface
type Tracer interface {
	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	Close(ctx context.Context) error
}