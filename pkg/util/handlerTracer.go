package util

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type HandlerTracer struct {
	tracer trace.Tracer
}

func NewHandlerTracer(tracer trace.Tracer) *HandlerTracer {
	return &HandlerTracer{tracer: tracer}
}

func (ht *HandlerTracer) TraceHttpHandler(r *http.Request, spanName string) (context.Context, trace.Span) {
	ctx := r.Context()
	return ht.tracer.Start(ctx, spanName)
}
