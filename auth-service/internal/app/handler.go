package app

import (
	"auth-service/internal/util"
	"database/sql"
	"log"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "Register")
	defer span.End()

	var req registerUserRequest
	if err := h.requestDecoder.Decode(ctx, r, &req); err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if err := h.validator.Validate(ctx, req); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	if _, err := h.repository.FindUserByEmail(ctx, req.Email); err != sql.ErrNoRows {
		h.responseWriter.WriteErrorResponse(ctx, w, http.StatusConflict, "user with this email already exists")
		return
	}

	_, hashSpan := h.tracer.Start(ctx, "hashing password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w)
		return
	}
	req.Password = string(hashedPassword)
	hashSpan.End()

	user, err := h.repository.InsertUser(ctx, req.toUser())
	if err != nil {
		log.Printf("failed to insert user: %v", err)
		h.responseWriter.WriteErrorResponse(ctx, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, user)
}
