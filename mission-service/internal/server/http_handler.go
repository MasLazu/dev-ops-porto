package server

import (
	"net/http"
	"time"

	"github.com/MasLazu/dev-ops-porto/mission-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
)

type HttpHandler struct {
	tracer         trace.Tracer
	responseWriter *util.ResponseWriter
	requestDecoder *util.RequestBodyDecoder
	validator      *util.Validator
	handlerTracer  *util.HandlerTracer
	service        *app.Service
	authMiddleware *middleware.AuthMiddleware
}

func NewHttpHandler(
	tracer trace.Tracer,
	responseWriter *util.ResponseWriter,
	requestDecoder *util.RequestBodyDecoder,
	validator *util.Validator,
	handlerTracer *util.HandlerTracer,
	service *app.Service,
	authMiddleware *middleware.AuthMiddleware,
) *HttpHandler {
	return &HttpHandler{
		tracer:         tracer,
		responseWriter: responseWriter,
		requestDecoder: requestDecoder,
		validator:      validator,
		handlerTracer:  handlerTracer,
		service:        service,
		authMiddleware: authMiddleware,
	}
}

func (h *HttpHandler) SetupRoutes(c *chi.Mux) http.Handler {
	c.Get("/health", h.HealthCheck)

	c.Group(func(c chi.Router) {
		c.Use(h.authMiddleware.Auth)
		c.Get("/", h.GetUserMissions)
		c.Get("/expiration", h.GetUserExpirationMissionDate)
	})

	return c
}

func (h *HttpHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.HealthCheck")
	defer span.End()

	response := h.service.HealthCheck(ctx)

	h.responseWriter.WriteSuccessResponse(ctx, w, response)
}

func (h *HttpHandler) GetUserMissions(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.GetUserMissions")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	missions, err := h.service.GetUserMissions(ctx, userID)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, missions)
}

func (h *HttpHandler) GetUserExpirationMissionDate(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.GetUserExpirationMissionDate")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	expirationDate, err := h.service.GetUserExpirationMissionDate(ctx, userID)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, map[string]time.Time{
		"expiration_date": expirationDate,
	})
}
