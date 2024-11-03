package app

import (
	"context"
	"database/sql"
	"time"

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

func (s *Service) GetUserMissions(ctx context.Context, userID string) ([]Mission, error) {
	ctx, span := s.tracer.Start(ctx, "Service.GetUserMissions")
	defer span.End()

	missions := make([]Mission, 0)

	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		tx.Rollback()
		return missions, err
	}

	user, err := s.createUserIfNotExists(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return missions, err
	}

	if user.ExpirationDate.Before(time.Now()) {
		user, err = s.resetUserMissions(ctx, tx, user)
		if err != nil {
			tx.Rollback()
			return missions, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return missions, err
	}

	missions, err = s.missionRepository.GetUserMissions(ctx, user.ID)
	if err != nil {
		return missions, err
	}

	return missions, nil
}

func (s *Service) GetUserExpirationMissionDate(ctx context.Context, userID string) (time.Time, error) {
	ctx, span := s.tracer.Start(ctx, "Service.GetUserExpirationMissionDate")
	defer span.End()

	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		tx.Rollback()
		return time.Time{}, err
	}

	user, err := s.createUserIfNotExists(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return time.Time{}, err
	}

	if user.ExpirationDate.Before(time.Now()) {
		user, err = s.resetUserMissions(ctx, tx, user)
		if err != nil {
			tx.Rollback()
			return time.Time{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return time.Time{}, err
	}

	return user.ExpirationDate, nil
}

func (s *Service) createUserIfNotExists(ctx context.Context, tx *sql.Tx, userID string) (User, error) {
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
