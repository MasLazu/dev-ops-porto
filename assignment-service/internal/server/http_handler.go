package server

import (
	"net/http"
	"strconv"

	"github.com/MasLazu/dev-ops-porto/assignment-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

		c.Post("/", h.CreateAssignment)
		c.Get("/", h.GetAssignments)
		c.Get("/{id}", h.GetAssignmentByID)
		c.Put("/{id}", h.UpdateAssignmentByID)
		c.Delete("/{id}", h.DeleteAssignmentByID)
		c.Post("/change-status", h.ChangeIsCompletedByID)
	})

	return c
}

func (h *HttpHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.HealthCheck")
	defer span.End()

	response := h.service.HealthCheck(ctx)

	h.responseWriter.WriteSuccessResponse(ctx, w, response)
}

func (h *HttpHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.CreateAssignment")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	var request app.CreateAssignmentRequest
	if err := h.requestDecoder.Decode(ctx, r, &request); err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if err := h.validator.Validate(ctx, request); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	assignment, err := h.service.CreateAssignment(ctx, uuid.MustParse(userID), request)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, assignment)
}

func (h *HttpHandler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.GetAssignments")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	assignments, err := h.service.GetAssignments(ctx, uuid.MustParse(userID))
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, assignments)
}

func (h *HttpHandler) GetAssignmentByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.GetAssignmentByID")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	assignmentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	assignment, serviceErr := h.service.GetAssignmentByID(ctx, uuid.MustParse(userID), int32(assignmentID))
	if serviceErr != nil {
		code := serviceErr.Code()
		h.responseWriter.WriteJSONResponseWithInternalError(ctx, w, code, http.StatusText(code), nil, serviceErr)
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, assignment)
}

func (h *HttpHandler) DeleteAssignmentByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.DeleteAssignmentByID")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	assignmentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if serviceErr := h.service.DeleteAssignmentByID(ctx, uuid.MustParse(userID), int32(assignmentID)); serviceErr != nil {
		code := serviceErr.Code()
		h.responseWriter.WriteJSONResponseWithInternalError(ctx, w, code, http.StatusText(code), nil, serviceErr)
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, nil)
}

func (h *HttpHandler) UpdateAssignmentByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.UpdateAssignmentByID")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	assignmentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	var request app.UpdateAssignmentRequest
	if err := h.requestDecoder.Decode(ctx, r, &request); err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if err := h.validator.Validate(ctx, request); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	assignment, serviceErr := h.service.UpdateAssignmentByID(ctx, uuid.MustParse(userID), int32(assignmentID), request)
	if serviceErr != nil {
		code := serviceErr.Code()
		h.responseWriter.WriteJSONResponseWithInternalError(ctx, w, code, http.StatusText(code), nil, serviceErr)
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, assignment)
}

func (h *HttpHandler) ChangeIsCompletedByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.ChangeIsCompletedByID")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	var request app.ChangeIsCompletedRequest
	if err := h.requestDecoder.Decode(ctx, r, &request); err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if err := h.validator.Validate(ctx, request); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	assignment, serviceErr := h.service.ChangeIsCompletedByID(ctx, uuid.MustParse(userID), int32(request.ID), request.IsCompleted)
	if serviceErr != nil {
		code := serviceErr.Code()
		h.responseWriter.WriteJSONResponseWithInternalError(ctx, w, code, http.StatusText(code), nil, serviceErr)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, assignment)
}
