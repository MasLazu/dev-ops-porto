package app

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/MasLazu/dev-ops-porto/assignment-service/internal/util"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	tracer               trace.Tracer
	responseWriter       *util.ResponseWriter
	requestDecoder       *util.RequestBodyDecoder
	validator            *util.Validator
	handlerTracer        *util.HandlerTracer
	repository           *Repository
	assignmentRepository *AssignmentRepository
	reminderRepository   *ReminderRepository
}

func NewHandler(
	tracer trace.Tracer,
	responseWriter *util.ResponseWriter,
	requestDecoder *util.RequestBodyDecoder,
	validator *util.Validator,
	handlerTracer *util.HandlerTracer,
	repository *Repository,
	assignmentRepository *AssignmentRepository,
	reminderRepository *ReminderRepository,
) *Handler {
	return &Handler{
		tracer:               tracer,
		responseWriter:       responseWriter,
		requestDecoder:       requestDecoder,
		validator:            validator,
		handlerTracer:        handlerTracer,
		assignmentRepository: assignmentRepository,
		repository:           repository,
		reminderRepository:   reminderRepository,
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

func (h *Handler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "CreateAssignmentHandler")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	var request CreateAssignmentRequest
	if err := h.requestDecoder.Decode(ctx, r, &request); err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}

	if err := h.validator.Validate(ctx, request); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	assignment, reminders := request.toAssignmentAndReminders(userID)

	tx, err := h.repository.BeginTransaction(ctx)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	assignment, err = h.assignmentRepository.InsertAssignmentWithTransaction(ctx, tx, assignment)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	if reminders == nil {
		tx.Commit()
		h.responseWriter.WriteSuccessResponse(ctx, w, assignment)
		return
	}

	for i := range reminders {
		reminders[i].AssignmentID = assignment.ID
	}

	assignment.Reminders, err = h.reminderRepository.InsertRemindersWithTransaction(ctx, tx, reminders)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}
	tx.Commit()

	h.responseWriter.WriteSuccessResponse(ctx, w, assignment)
}

func (h *Handler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "GetAssignmentsHandler")
	defer span.End()

	userID, err := util.GetUserIDFromContext(ctx)
	if err != nil {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	assignments, err := h.assignmentRepository.FindAssignmentsByUserIDJoinReminders(ctx, userID)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, assignments)
}

func (h *Handler) GetAssignmentByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "GetAssignmentByIDHandler")
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

	assignment, err := h.assignmentRepository.FindAssignmentByIDJoinReminders(ctx, assignmentID)
	if err == sql.ErrNoRows {
		h.responseWriter.WriteNotFoundResponse(ctx, w)
		return
	}
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}
	if assignment.UserID != userID {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	h.responseWriter.WriteSuccessResponse(ctx, w, assignment)
}

func (h *Handler) DeleteAssignmentByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "DeleteAssignmentByIDHandler")
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

	assignment, err := h.assignmentRepository.FindAssignmentByID(ctx, assignmentID)
	if err == sql.ErrNoRows {
		h.responseWriter.WriteNotFoundResponse(ctx, w)
		return
	}
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}
	if assignment.UserID != userID {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	tx, err := h.repository.BeginTransaction(ctx)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	err = h.reminderRepository.DeleteRemindersByAssignmentIDWithTransaction(ctx, tx, assignment.ID)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	err = h.assignmentRepository.DeleteAssignmentByIDWithTransaction(ctx, tx, assignmentID)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}
	tx.Commit()

	h.responseWriter.WriteSuccessResponse(ctx, w, nil)
}

func (h *Handler) UpdateAssignmentByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "UpdateAssignmentByIDHandler")
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

	var request UpdateAssignmentRequest
	if err := h.requestDecoder.Decode(ctx, r, &request); err != nil {
		h.responseWriter.WriteBadRequestResponse(ctx, w)
		return
	}
	assignmentRequest := request.toAssignment()

	if err := h.validator.Validate(ctx, request); err != nil {
		h.responseWriter.WriteValidationErrorResponse(ctx, w, *err)
		return
	}

	assignment, err := h.assignmentRepository.FindAssignmentByID(ctx, assignmentID)
	if err == sql.ErrNoRows {
		h.responseWriter.WriteNotFoundResponse(ctx, w)
		return
	}
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}
	if assignment.UserID != userID {
		h.responseWriter.WriteUnauthorizedResponse(ctx, w)
		return
	}

	assignment.Title = assignmentRequest.Title
	assignment.Note = assignmentRequest.Note
	assignment.DueDate = assignmentRequest.DueDate
	assignment.IsCompleted = assignmentRequest.IsCompleted
	assignment.IsImportant = assignmentRequest.IsImportant

	tx, err := h.repository.BeginTransaction(ctx)
	if err != nil {
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	assignment, err = h.assignmentRepository.UpdateAssignmentWithTransaction(ctx, tx, assignment)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	err = h.reminderRepository.DeleteRemindersByAssignmentIDWithTransaction(ctx, tx, assignment.ID)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	reminders, err := h.reminderRepository.InsertRemindersWithTransaction(ctx, tx, assignmentRequest.Reminders)
	if err != nil {
		tx.Rollback()
		h.responseWriter.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}
	assignment.Reminders = reminders
	tx.Commit()

	h.responseWriter.WriteSuccessResponse(ctx, w, assignment)
}
