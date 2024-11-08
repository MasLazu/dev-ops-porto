package server

import (
	"io"
	"net/http"

	"github.com/MasLazu/dev-ops-porto/auth-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
)

type HttpHandler struct {
	service        *app.Service
	authMiddleware *middleware.AuthMiddleware
	tracer         trace.Tracer
	responseWriter *util.ResponseWriter
	requestDecoder *util.RequestBodyDecoder
	validator      *util.Validator
	handlerTracer  *util.HandlerTracer
}

func NewHttpHandler(
	service *app.Service,
	authMiddleware *middleware.AuthMiddleware,
	tracer trace.Tracer,
	responseWriter *util.ResponseWriter,
	requestDecoder *util.RequestBodyDecoder,
	validator *util.Validator,
	handlerTracer *util.HandlerTracer,
) *HttpHandler {
	return &HttpHandler{
		service:        service,
		authMiddleware: authMiddleware,
		tracer:         tracer,
		responseWriter: responseWriter,
		requestDecoder: requestDecoder,
		validator:      validator,
		handlerTracer:  handlerTracer,
	}
}

func (h *HttpHandler) setupRoutes(c *chi.Mux) http.Handler {
	c.Get("/health", h.HealthCheck)
	c.Post("/register", h.Register)
	c.Post("/login", h.Login)

	c.Group(func(c chi.Router) {
		c.Use(h.authMiddleware.Auth)
		c.Get("/me", h.Me)
		c.Post("/me/profile-picture", h.ChangeProfilePicture)
		c.Delete("/me/profile-picture", h.DeleteProfilePicture)
	})

	return c
}

func (h *HttpHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.HealthCheck")
	defer span.End()

	response := h.service.HealthCheck(ctx)

	h.responseWriter.WriteSuccessResponse(ctx, w, response)
}

func (h *HttpHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.Register")
	defer span.End()

	var req app.RegisterUserRequest
	if err := h.requestDecoder.Decode(ctx, r, &req); err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if err := h.validator.Validate(ctx, req); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	user, err := h.service.Register(ctx, req)
	if err != nil {
		h.responseWriter.WriteErrorResponse(ctx, w, int(err.HttpCode()), err.ClientMessage())
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, user)
}

func (h *HttpHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.Login")
	defer span.End()

	var req app.LoginUserRequest
	if err := h.requestDecoder.Decode(ctx, r, &req); err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if err := h.validator.Validate(ctx, req); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	loginResponse, err := h.service.Login(ctx, req)
	if err != nil {
		h.responseWriter.WriteErrorResponse(ctx, w, err.HttpCode(), err.ClientMessage())
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, loginResponse)
}

func (h *HttpHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.Me")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	user, serviceErr := h.service.Me(ctx, userID)
	if serviceErr != nil {
		h.responseWriter.WriteErrorResponse(ctx, w, serviceErr.HttpCode(), serviceErr.ClientMessage())
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, user)
}

func (h *HttpHandler) ChangeProfilePicture(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.ChangeProfilePicture")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	fileMultipart, _, err := r.FormFile("profile_picture")
	if err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	file, err := io.ReadAll(fileMultipart)
	if err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	user, serviceErr := h.service.ChangeProfilePicture(ctx, userID, file)
	if serviceErr != nil {
		h.responseWriter.WriteErrorResponse(ctx, w, serviceErr.HttpCode(), serviceErr.ClientMessage())
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, user)
}

func (h *HttpHandler) DeleteProfilePicture(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.DeleteProfilePicture")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	user, serviceErr := h.service.DeleteProfilePicture(ctx, userID)
	if serviceErr != nil {
		h.responseWriter.WriteErrorResponse(ctx, w, serviceErr.HttpCode(), serviceErr.ClientMessage())
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, user)
}
