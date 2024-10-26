package app

import (
	"assignment-service/internal/util"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	tracer         trace.Tracer
	responseWriter *util.ResponseWriter
	requestDecoder *util.RequestBodyDecoder
	validator      *util.Validator
	handlerTracer  *util.HandlerTracer
	repository     *Repository
}

func NewHandler(
	tracer trace.Tracer,
	responseWriter *util.ResponseWriter,
	requestDecoder *util.RequestBodyDecoder,
	validator *util.Validator,
	handlerTracer *util.HandlerTracer,
	repository *Repository,
) *Handler {
	return &Handler{
		tracer:         tracer,
		responseWriter: responseWriter,
		requestDecoder: requestDecoder,
		validator:      validator,
		handlerTracer:  handlerTracer,
		repository:     repository,
	}
}

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HealthCheckHandler")
	defer span.End()

	response := h.repository.Health(ctx)

	h.responseWriter.WriteSuccessResponse(ctx, w, response)
}
