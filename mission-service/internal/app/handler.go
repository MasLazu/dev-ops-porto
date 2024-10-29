package app

import (
	"mission-service/internal/util"
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

func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "NotFoundHandler")
	defer span.End()

	h.responseWriter.WriteNotFoundResponse(ctx, w)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HealthCheckHandler")
	defer span.End()

	response := h.repository.Health(ctx)

	h.responseWriter.WriteSuccessResponse(ctx, w, response)
}
