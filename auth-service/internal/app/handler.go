package app

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/golang-jwt/jwt/v5"
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
	jwtSecret      []byte
}

func NewHandler(
	tracer trace.Tracer,
	responseWriter *util.ResponseWriter,
	requestDecoder *util.RequestBodyDecoder,
	validator *util.Validator,
	handlerTracer *util.HandlerTracer,
	repository *Repository,
	jwtSecret []byte,
) *Handler {
	return &Handler{
		tracer:         tracer,
		responseWriter: responseWriter,
		requestDecoder: requestDecoder,
		validator:      validator,
		handlerTracer:  handlerTracer,
		repository:     repository,
		jwtSecret:      jwtSecret,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HealthCheckHandler")
	defer span.End()

	response := h.repository.Health(ctx)

	h.responseWriter.WriteSuccessResponse(ctx, w, response)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "RegisterHandler")
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
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}
	req.Password = string(hashedPassword)
	hashSpan.End()

	user, err := h.repository.InsertUser(ctx, req.toUser())
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "LoginHandler")
	defer span.End()

	var req loginUserRequest
	if err := h.requestDecoder.Decode(ctx, r, &req); err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if err := h.validator.Validate(ctx, req); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	user, err := h.repository.FindUserByEmail(ctx, req.Email)
	if err == sql.ErrNoRows {
		h.responseWriter.WriteErrorResponse(ctx, w, http.StatusUnauthorized, "email or password is invalid")
		return
	}

	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	_, hashSpan := h.tracer.Start(ctx, "comparing password")
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.responseWriter.WriteErrorResponse(ctx, w, http.StatusUnauthorized, "email or password is invalid")
		return
	}
	hashSpan.End()

	_, tokenSpan := h.tracer.Start(ctx, "generating jwt token")
	claims := &jwt.RegisteredClaims{
		Subject:   user.ID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30 * 12)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}
	tokenSpan.End()

	h.responseWriter.WriteSuccessResponse(ctx, w, loginResponse{AccessToken: signedToken})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "MeHandler")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	user, err := h.repository.FindUserByID(ctx, userID)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, user)
}
