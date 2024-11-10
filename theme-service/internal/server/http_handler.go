package server

import (
	"net/http"

	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/MasLazu/dev-ops-porto/theme-service/internal/app"
	"github.com/go-chi/chi/v5"
)

type HttpHandler struct {
	service        *app.Service
	handlerTracer  *util.HandlerTracer
	authMiddleware *middleware.AuthMiddleware
	responseWriter *util.ResponseWriter
	requestDecoder *util.RequestBodyDecoder
	validator      *util.Validator
}

func NewHttpHandler(
	service *app.Service,
	authMiddleware *middleware.AuthMiddleware,
	handlerTracer *util.HandlerTracer,
	responseWriter *util.ResponseWriter,
	requestDecoder *util.RequestBodyDecoder,
	validator *util.Validator,
) *HttpHandler {
	return &HttpHandler{
		service:        service,
		handlerTracer:  handlerTracer,
		authMiddleware: authMiddleware,
		responseWriter: responseWriter,
		requestDecoder: requestDecoder,
		validator:      validator,
	}
}

func (h *HttpHandler) setupRoutes(c *chi.Mux) http.Handler {
	c.Get("/health", h.HealthCheck)

	c.Group(func(c chi.Router) {
		c.Use(h.authMiddleware.Auth)
		c.Get("/", h.GetOwnedTheme)
		c.Post("/", h.UnlockTheme)
	})

	return c
}

func (h *HttpHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.HealthCheck")
	defer span.End()

	response := h.service.HealthCheck(ctx)

	h.responseWriter.WriteSuccessResponse(ctx, w, response)
}

func (h *HttpHandler) GetOwnedTheme(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.GetOwnedTheme")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	theme, err := h.service.GetOwnedTheme(ctx, userID)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, theme)
}

func (h *HttpHandler) UnlockTheme(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.UnlockTheme")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	var req app.UnlockThemeRequest
	err = h.requestDecoder.Decode(ctx, r, &req)
	if err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if err := h.validator.Validate(ctx, req); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	err = h.service.UnlockTheme(ctx, userID, req.ThemeID)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, nil)
}
