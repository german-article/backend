package logging

import (
	"context"
)

// Logger defines the logging interface
type Logger interface {
	Debug(ctx context.Context, payload interface{})
	Info(ctx context.Context, payload interface{})
	Warning(ctx context.Context, payload interface{})
	Error(ctx context.Context, payload interface{})
	Critical(ctx context.Context, payload interface{})
	Close(ctx context.Context) error
}