package app

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/MasLazu/dev-ops-porto/pkg/genproto/missionservice"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	tracer               trace.Tracer
	repository           *Repository
	assignmentRepository *AssignmentRepository
	reminderRepository   *ReminderRepository
	missionServiceClient missionservice.MissionServiceClient
}

func NewService(
	tracer trace.Tracer,
	repository *Repository,
	assignmentRepository *AssignmentRepository,
	reminderRepository *ReminderRepository,
	missionServiceClient missionservice.MissionServiceClient,
) *Service {
	return &Service{
		tracer:               tracer,
		repository:           repository,
		assignmentRepository: assignmentRepository,
		reminderRepository:   reminderRepository,
		missionServiceClient: missionServiceClient,
	}
}

func (s *Service) triggerMissionEvent(ctx context.Context, userID string, event missionservice.TriggerMissionEvent) error {
	_, triggerMissionEventSpan := s.tracer.Start(ctx, "Service.triggerMissionEvent")
	defer triggerMissionEventSpan.End()

	_, err := s.missionServiceClient.TriggerMissionEvent(ctx, &missionservice.TriggerMissionEventRequest{
		UserId: userID,
		Event:  event,
	})
	if err != nil {
		triggerMissionEventSpan.RecordError(err)
	}
	return err
}

func (s *Service) getAuthorizedAssignmentByID(ctx context.Context, userID string, assignmentID int) (Assignment, ServiceError) {
	assignment, err := s.assignmentRepository.FindAssignmentByIDJoinReminders(ctx, assignmentID)
	if err == sql.ErrNoRows {
		return Assignment{}, NewClientError(http.StatusNotFound, err)
	}
	if err != nil {
		return Assignment{}, NewInternalServiceError(err)
	}
	if assignment.UserID != userID {
		return Assignment{}, NewClientError(http.StatusForbidden, errors.New("assignment does not belong to user"))
	}

	return assignment, nil
}

func (s *Service) HealthCheck(ctx context.Context) map[string]map[string]string {
	ctx, span := s.tracer.Start(ctx, "Service.HealthCheck")
	defer span.End()

	res := make(map[string]map[string]string)
	res["database"] = s.repository.Health(ctx)
	// res["mission-service"] = s.missionServiceClient.Health(ctx)

	return res
}

func (s *Service) CreateAssignment(ctx context.Context, userID string, request CreateAssignmentRequest) (Assignment, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.CreateAssignment")
	defer span.End()

	assignment, reminders := request.toAssignmentAndReminders(userID)

	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		tx.Rollback()
		return Assignment{}, NewInternalServiceError(err)
	}

	assignment, err = s.assignmentRepository.InsertAssignmentWithTransaction(ctx, tx, assignment)
	if err != nil {
		tx.Rollback()
		return Assignment{}, NewInternalServiceError(err)
	}

	if reminders == nil {
		tx.Commit()
		return assignment, NewInternalServiceError(err)
	}

	for i := range reminders {
		reminders[i].AssignmentID = assignment.ID
	}

	assignment.Reminders, err = s.reminderRepository.InsertRemindersWithTransaction(ctx, tx, reminders)
	if err != nil {
		tx.Rollback()
		return Assignment{}, NewInternalServiceError(err)
	}
	tx.Commit()

	err = s.triggerMissionEvent(ctx, userID, missionservice.TriggerMissionEvent_MISSION_EVENT_CREATE_ASSIGNMENT)
	if err != nil {
		return Assignment{}, NewInternalServiceError(err)
	}

	return assignment, nil
}

func (s *Service) GetAssignments(ctx context.Context, userID string) ([]Assignment, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.GetAssignments")
	defer span.End()

	assignments, err := s.assignmentRepository.FindAssignmentsByUserIDJoinReminders(ctx, userID)
	if err != nil {
		return nil, NewInternalServiceError(err)
	}

	return assignments, nil
}

func (s *Service) GetAssignmentByID(ctx context.Context, userID string, assignmentID int) (Assignment, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.GetAssignmentByID")
	defer span.End()

	assignment, err := s.getAuthorizedAssignmentByID(ctx, userID, assignmentID)
	if err != nil {
		return Assignment{}, err
	}

	return assignment, nil
}

func (s *Service) DeleteAssignmentByID(ctx context.Context, userID string, assignmentID int) ServiceError {
	ctx, span := s.tracer.Start(ctx, "Service.DeleteAssignmentByID")
	defer span.End()

	assignment, serviceErr := s.getAuthorizedAssignmentByID(ctx, userID, assignmentID)
	if serviceErr != nil {
		return serviceErr
	}

	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		return NewInternalServiceError(err)
	}

	err = s.reminderRepository.DeleteRemindersByAssignmentIDWithTransaction(ctx, tx, assignment.ID)
	if err != nil {
		tx.Rollback()
		return NewInternalServiceError(err)
	}

	err = s.assignmentRepository.DeleteAssignmentByIDWithTransaction(ctx, tx, assignmentID)
	if err != nil {
		tx.Rollback()
		return NewInternalServiceError(err)
	}
	tx.Commit()

	err = s.triggerMissionEvent(ctx, userID, missionservice.TriggerMissionEvent_MISSION_EVENT_DELETE_ASSIGNMENT)
	if err != nil {
		return NewInternalServiceError(err)
	}

	return nil
}

func (s *Service) UpdateAssignmentByID(ctx context.Context, userID string, assignmentID int, request UpdateAssignmentRequest) (Assignment, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.UpdateAssignmentByID")
	defer span.End()

	assignment, serviceErr := s.getAuthorizedAssignmentByID(ctx, userID, assignmentID)
	if serviceErr != nil {
		return Assignment{}, serviceErr
	}

	assignmentRequest := request.toAssignment()

	assignment.Title = assignmentRequest.Title
	assignment.Note = assignmentRequest.Note
	assignment.DueDate = assignmentRequest.DueDate
	assignment.IsCompleted = assignmentRequest.IsCompleted
	assignment.IsImportant = assignmentRequest.IsImportant

	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		return Assignment{}, NewInternalServiceError(err)
	}

	assignment, err = s.assignmentRepository.UpdateAssignmentWithTransaction(ctx, tx, assignment)
	if err != nil {
		tx.Rollback()
		return Assignment{}, NewInternalServiceError(err)
	}

	err = s.reminderRepository.DeleteRemindersByAssignmentIDWithTransaction(ctx, tx, assignment.ID)
	if err != nil {
		tx.Rollback()
		return Assignment{}, NewInternalServiceError(err)
	}

	reminders, err := s.reminderRepository.InsertRemindersWithTransaction(ctx, tx, assignmentRequest.Reminders)
	if err != nil {
		tx.Rollback()
		return Assignment{}, NewInternalServiceError(err)
	}
	assignment.Reminders = reminders
	tx.Commit()

	return assignment, nil
}

func (s *Service) ChangeIsCompletedByID(ctx context.Context, userID string, assignmentID int, isCompleted bool) (Assignment, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.ChangeIsCompletedByID")
	defer span.End()

	assignment, serviceErr := s.getAuthorizedAssignmentByID(ctx, userID, assignmentID)
	if serviceErr != nil {
		return Assignment{}, serviceErr
	}

	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		return Assignment{}, NewInternalServiceError(err)
	}

	assignment.IsCompleted = isCompleted

	assignment, err = s.assignmentRepository.UpdateAssignmentWithTransaction(ctx, tx, assignment)
	if err != nil {
		return Assignment{}, NewInternalServiceError(err)
	}

	var event missionservice.TriggerMissionEvent

	if assignment.IsCompleted != isCompleted && isCompleted {
		event = missionservice.TriggerMissionEvent_MISSION_EVENT_DONE_ASSIGNMENT
	}

	if assignment.IsCompleted != isCompleted && !isCompleted {
		event = missionservice.TriggerMissionEvent_MISSION_EVENT_UNDONE_ASSIGNMENT

	}

	if err := s.triggerMissionEvent(ctx, userID, event); err != nil {
		tx.Rollback()
		return Assignment{}, NewInternalServiceError(err)
	}

	tx.Commit()

	return assignment, nil
}
