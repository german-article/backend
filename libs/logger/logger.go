package logger

import (
	"context"
	"os"

	"cloud.google.com/go/logging"
	"go.opentelemetry.io/otel/trace"
)

type Log struct {
	gcp       bool
	projectID string
	client    *logging.Client
	logger    *logging.Logger
	level     logging.Severity
}

func New(client *logging.Client, logger *logging.Logger, level logging.Severity, projectID string, gcp bool) *Log {
	return &Log{
		gcp:       gcp,
		projectID: projectID,
		client:    client,
		logger:    logger,
		level:     level,
	}
}

func Init(ctx context.Context, pID, applicationName string, gcp bool, level logging.Severity, opts ...logging.LoggerOption) (*Log, error) {
	client, err := logging.NewClient(ctx, pID)
	if err != nil {
		return nil, err
	}

	if !gcp {
		opts = append(opts, logging.RedirectAsJSON(os.Stdout))
	}

	return New(client, client.Logger(applicationName, opts...), level, pID, gcp), nil
}

func (l *Log) Close(_ context.Context) error {
	if l.client != nil {
		if err := l.client.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Log Default means the log entry has no assigned severity level.
func (l *Log) Log(ctx context.Context, payload interface{}) {
	l.logger.Log(l.build(ctx, logging.Default, payload))
}

// Debug means debug or trace information.
func (l *Log) Debug(ctx context.Context, payload interface{}) {
	if l.level > logging.Debug {
		return
	}

	l.logger.Log(l.build(ctx, logging.Debug, payload))
}

// Info means routine information, such as ongoing status or performance.
func (l *Log) Info(ctx context.Context, payload interface{}) {
	if l.level > logging.Info {
		return
	}

	l.logger.Log(l.build(ctx, logging.Info, payload))
}

// Notice means normal but significant events, such as start up, shut down, or configuration.
func (l *Log) Notice(ctx context.Context, payload interface{}) {
	if l.level > logging.Notice {
		return
	}

	l.logger.Log(l.build(ctx, logging.Notice, payload))
}

// Warning means events that might cause problems.
func (l *Log) Warning(ctx context.Context, payload interface{}) {
	if l.level > logging.Warning {
		return
	}

	l.logger.Log(l.build(ctx, logging.Warning, payload))
}

// Error means events that are likely to cause problems.
func (l *Log) Error(ctx context.Context, payload interface{}) {
	if l.level > logging.Error {
		return
	}

	l.logger.Log(l.build(ctx, logging.Error, payload))
}

// Critical means events that cause more severe problems or brief outages.
func (l *Log) Critical(ctx context.Context, payload interface{}) {
	if l.level > logging.Critical {
		return
	}

	l.logger.Log(l.build(ctx, logging.Critical, payload))
}

// Alert means a person must take action immediately.
func (l *Log) Alert(ctx context.Context, payload interface{}) {
	if l.level > logging.Alert {
		return
	}

	l.logger.Log(l.build(ctx, logging.Alert, payload))
}

// Emergency means one or more systems are unusable.
func (l *Log) Emergency(ctx context.Context, payload interface{}) {
	if l.level > logging.Emergency {
		return
	}

	l.logger.Log(l.build(ctx, logging.Emergency, payload))
}

func (l *Log) build(ctx context.Context, severity logging.Severity, payload interface{}) logging.Entry {
	e := logging.Entry{
		Payload:  payload,
		Severity: severity,
	}

	if !l.gcp {
		return e
	}

	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		e.Trace = "projects/" + l.projectID + "/traces/" + sc.TraceID().String()
		e.SpanID = sc.SpanID().String()
		e.TraceSampled = sc.IsSampled()
	}
	return e
}
