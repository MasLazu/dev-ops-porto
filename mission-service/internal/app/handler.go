package app

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	tracer                trace.Tracer
	responseWriter        *util.ResponseWriter
	requestDecoder        *util.RequestBodyDecoder
	validator             *util.Validator
	handlerTracer         *util.HandlerTracer
	repository            *Repository
	userRepository        *UserRepository
	userMissionRepository *UserMissionRepository
	missionRepository     *MissionRepository
}

func NewHandler(
	tracer trace.Tracer,
	responseWriter *util.ResponseWriter,
	requestDecoder *util.RequestBodyDecoder,
	validator *util.Validator,
	handlerTracer *util.HandlerTracer,
	repository *Repository,
	userRepository *UserRepository,
	userMissionRepository *UserMissionRepository,
	missionRepository *MissionRepository,
) *Handler {
	return &Handler{
		tracer:                tracer,
		responseWriter:        responseWriter,
		requestDecoder:        requestDecoder,
		validator:             validator,
		handlerTracer:         handlerTracer,
		repository:            repository,
		userRepository:        userRepository,
		userMissionRepository: userMissionRepository,
		missionRepository:     missionRepository,
	}
}

const missionPeriod = time.Hour * 24

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HealthCheckHandler")
	defer span.End()

	response := h.repository.Health(ctx)

	h.responseWriter.WriteSuccessResponse(ctx, w, response)
}

func (h *Handler) GetUserMissions(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "GetUserMissionsHandler")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	tx, err := h.repository.BeginTransaction(ctx)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	user, err := h.createUserIfNotExists(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	if user.ExpirationDate.Before(time.Now()) {
		user, err = h.resetUserMissions(ctx, tx, user)
		if err != nil {
			tx.Rollback()
			h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	missions, err := h.missionRepository.GetUserMissions(ctx, user.ID)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, missions)
}

func (h *Handler) GetUserExpirationMissionDate(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "GetUserMissionExpirationDateHandler")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	tx, err := h.repository.BeginTransaction(ctx)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	user, err := h.createUserIfNotExists(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	if user.ExpirationDate.Before(time.Now()) {
		user, err = h.resetUserMissions(ctx, tx, user)
		if err != nil {
			tx.Rollback()
			h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	response := UserExpirationMissionDateResponse{
		ExpirationDate: user.ExpirationDate,
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, response)
}

func (h *Handler) createUserIfNotExists(ctx context.Context, tx *sql.Tx, userID string) (User, error) {
	user, err := h.userRepository.GetUserByID(ctx, userID)
	if err != nil && err != sql.ErrNoRows {
		return user, err
	}

	if err == nil {
		return user, nil
	}

	user, err = h.userRepository.InsertUserWithTransaction(ctx, tx, User{
		ID:             userID,
		ExpirationDate: time.Now().Add(missionPeriod),
	})
	if err != nil {
		return user, err
	}

	user, err = h.resetUserMissions(ctx, tx, user)

	return user, err
}

func (h *Handler) resetUserMissions(ctx context.Context, tx *sql.Tx, user User) (User, error) {
	user.ExpirationDate = time.Now().Add(missionPeriod)

	user, err := h.userRepository.UpdateUserWithTransaction(ctx, tx, user)
	if err != nil {
		return user, err
	}

	err = h.userMissionRepository.DeleteUserMissionByUserIDWithTransaction(ctx, tx, user.ID)
	if err != nil {
		return user, err
	}

	missionsIDs, err := h.missionRepository.GetTwoRandomMissionIDs(ctx)
	if err != nil {
		return user, err
	}

	if err := h.userMissionRepository.InsertUserMissionsWithTransaction(ctx, tx, user.ID, missionsIDs); err != nil {
		return user, err
	}

	return user, nil
}
