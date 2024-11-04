package app

import (
	"context"
	"database/sql"
	"time"

	"github.com/MasLazu/dev-ops-porto/pkg/genproto/missionservice"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	tracer                trace.Tracer
	repository            *Repository
	userRepository        *UserRepository
	userMissionRepository *UserMissionRepository
	missionRepository     *MissionRepository
}

func NewService(
	tracer trace.Tracer,
	repository *Repository,
	userRepository *UserRepository,
	userMissionRepository *UserMissionRepository,
	missionRepository *MissionRepository,
) *Service {
	return &Service{
		tracer:                tracer,
		repository:            repository,
		userRepository:        userRepository,
		userMissionRepository: userMissionRepository,
		missionRepository:     missionRepository,
	}
}

const missionPeriod = time.Hour * 24

func (s *Service) HealthCheck(ctx context.Context) map[string]string {
	return s.repository.Health(ctx)
}

func (s *Service) GetUserMissions(ctx context.Context, userID string) ([]UserMission, error) {
	ctx, span := s.tracer.Start(ctx, "Service.GetUserMissions")
	defer span.End()

	userMissions := make([]UserMission, 0)

	if _, err := s.SyncUserAndMissions(ctx, userID); err != nil {
		return userMissions, err
	}

	userMissions, err := s.userMissionRepository.GetUserMissionsByUserIDJoinMission(ctx, userID)
	if err != nil {
		return userMissions, err
	}

	return userMissions, nil
}

func (s *Service) GetUserExpirationMissionDate(ctx context.Context, userID string) (time.Time, error) {
	ctx, span := s.tracer.Start(ctx, "Service.GetUserExpirationMissionDate")
	defer span.End()

	user, err := s.SyncUserAndMissions(ctx, userID)
	if err != nil {
		return time.Time{}, err
	}

	return user.ExpirationDate, nil
}

func (s *Service) TriggerMissionEvent(ctx context.Context, userID string, triggerMissionEvent missionservice.TriggerMissionEvent) error {
	_, span := s.tracer.Start(ctx, "Service.TriggerMissionEvent")
	defer span.End()

	if _, err := s.SyncUserAndMissions(ctx, userID); err != nil {
		return err
	}

	encrease, err := s.userMissionRepository.GetUserMissionsByUserIDAndEncreasorEventIDJoinMission(ctx, userID, int(triggerMissionEvent))
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	decrease, err := s.userMissionRepository.GetUserMissionsByUserIDAndDecreasorEventIDJoinMission(ctx, userID, int(triggerMissionEvent))
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	userMissions := make([]UserMission, 0, len(encrease)+len(decrease))
	for _, um := range encrease {
		if um.Progress < um.Mission.Goal {
			um.Progress++
		}
		userMissions = append(userMissions, um)
	}

	for _, um := range decrease {
		if um.Progress > 0 {
			um.Progress--
		}
		userMissions = append(userMissions, um)
	}

	return s.userMissionRepository.UpdateUserMissions(ctx, userMissions)
}

func (s *Service) SyncUserAndMissions(ctx context.Context, userID string) (User, error) {
	_, span := s.tracer.Start(ctx, "Service.SyncUserAndMissions")
	defer span.End()

	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		tx.Rollback()
		return User{}, err
	}

	user, err := s.createUserIfNotExists(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return User{}, err
	}

	if user.ExpirationDate.Before(time.Now()) {
		user, err = s.resetUserMissions(ctx, tx, user)
		if err != nil {
			tx.Rollback()
			return User{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *Service) createUserIfNotExists(ctx context.Context, tx *sql.Tx, userID string) (User, error) {
	ctx, span := s.tracer.Start(ctx, "Service.createUserIfNotExists")
	defer span.End()

	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil && err != sql.ErrNoRows {
		return user, err
	}

	if err == nil {
		return user, nil
	}

	user, err = s.userRepository.InsertUserWithTransaction(ctx, tx, User{
		ID:             userID,
		ExpirationDate: time.Now().Add(missionPeriod),
	})
	if err != nil {
		return user, err
	}

	user, err = s.resetUserMissions(ctx, tx, user)

	return user, err
}

func (s *Service) resetUserMissions(ctx context.Context, tx *sql.Tx, user User) (User, error) {
	ctx, span := s.tracer.Start(ctx, "Service.resetUserMissions")
	defer span.End()

	user.ExpirationDate = time.Now().Add(missionPeriod)

	user, err := s.userRepository.UpdateUserWithTransaction(ctx, tx, user)
	if err != nil {
		return user, err
	}

	err = s.userMissionRepository.DeleteUserMissionByUserIDWithTransaction(ctx, tx, user.ID)
	if err != nil {
		return user, err
	}

	missionsIDs, err := s.missionRepository.GetTwoRandomMissionIDs(ctx)
	if err != nil {
		return user, err
	}

	if err := s.userMissionRepository.InsertUserMissionsWithTransaction(ctx, tx, user.ID, missionsIDs); err != nil {
		return user, err
	}

	return user, nil
}
