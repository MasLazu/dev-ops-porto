package monitoring

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func newLoggerProvider(ctx context.Context, config Config, resource *resource.Resource) (*sdklog.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(ctx, otlploggrpc.WithInsecure(), otlploggrpc.WithEndpoint(config.OtlpDomain))
	if err != nil {
		return nil, err
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithResource(resource),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	), err
}

type Logger struct {
	otelLogger log.Logger
}

func NewLogger(serviceName string) *Logger {
	return &Logger{
		otelLogger: global.GetLoggerProvider().Logger(serviceName),
	}
}

func (l *Logger) Debug(ctx context.Context, body string, attributes ...log.KeyValue) {
	l.LogMessage(ctx, log.SeverityDebug, log.StringValue(body), attributes...)
}

func (l *Logger) Info(ctx context.Context, body string, attributes ...log.KeyValue) {
	l.LogMessage(ctx, log.SeverityInfo, log.StringValue(body), attributes...)
}

func (l *Logger) Warn(ctx context.Context, body string, attributes ...log.KeyValue) {
	l.LogMessage(ctx, log.SeverityWarn, log.StringValue(body), attributes...)
}

func (l *Logger) Error(ctx context.Context, body string, attributes ...log.KeyValue) {
	l.LogMessage(ctx, log.SeverityError, log.StringValue(body), attributes...)
}

func (l *Logger) Fatal(ctx context.Context, body string, attributes ...log.KeyValue) {
	l.LogMessage(ctx, log.SeverityFatal, log.StringValue(body), attributes...)
}

func (l *Logger) LogMessage(ctx context.Context, severity log.Severity, body log.Value, attributes ...log.KeyValue) {
	record := log.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.SetBody(body)
	record.AddAttributes(attributes...)

	l.otelLogger.Emit(ctx, record)
}
