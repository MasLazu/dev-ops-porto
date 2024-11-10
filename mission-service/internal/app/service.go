package app

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	err "github.com/MasLazu/dev-ops-porto/pkg/errors"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/authservice"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/missionservice"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/status"
)

type Service struct {
	tracer                trace.Tracer
	repository            *Repository
	userRepository        *UserRepository
	userMissionRepository *UserMissionRepository
	missionRepository     *MissionRepository
	authServiceClient     authservice.AuthServiceClient
}

func NewService(
	tracer trace.Tracer,
	repository *Repository,
	userRepository *UserRepository,
	userMissionRepository *UserMissionRepository,
	missionRepository *MissionRepository,
	authServiceClient authservice.AuthServiceClient,
) *Service {
	return &Service{
		tracer:                tracer,
		repository:            repository,
		userRepository:        userRepository,
		userMissionRepository: userMissionRepository,
		missionRepository:     missionRepository,
		authServiceClient:     authServiceClient,
	}
}

const missionPeriod = time.Hour * 24

func (s *Service) newError(code code.Code, internalError error) err.ServiceError {
	return err.NewServiceError(code, internalError)
}

func (s *Service) newErrorWithCLientMessage(code code.Code, internalError error, clientMessage string) err.ServiceError {
	return err.NewServiceErrorWithClientMessage(code, internalError, clientMessage)
}

func (s *Service) newInternalError(internalError error) err.ServiceError {
	return s.newError(code.Code_INTERNAL, internalError)
}

func (s *Service) HealthCheck(ctx context.Context) map[string]string {
	ctx, span := s.tracer.Start(ctx, "Service.HealthCheck")
	defer span.End()

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

func (s *Service) CaimMissionReward(ctx context.Context, userMissionID int, userID string) err.ServiceError {
	_, span := s.tracer.Start(ctx, "Service.caimMissionReward")
	defer span.End()

	userMission, err := s.userMissionRepository.GetUserMissionByIDJoinMission(ctx, userMissionID)
	if err != nil {
		log.Printf("Error: %v", err)
		return s.newErrorWithCLientMessage(code.Code_NOT_FOUND, err, "Mission not found")
	}

	if userMission.UserID != userID {
		return s.newError(code.Code_PERMISSION_DENIED, errors.New("Mission does not belong to user"))
	}

	if userMission.Claimed {
		return s.newErrorWithCLientMessage(code.Code_INVALID_ARGUMENT, nil, "Mission already claimed")
	}

	if userMission.Progress < userMission.Mission.Goal {
		return s.newErrorWithCLientMessage(code.Code_INVALID_ARGUMENT, nil, "Mission not completed yet")
	}

	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		return s.newInternalError(err)
	}

	userMission.Claimed = true

	err = s.userMissionRepository.UpdateUserMissionWithTransaction(ctx, tx, userMission)
	if err != nil {
		tx.Rollback()
		return s.newInternalError(err)
	}

	req := &authservice.UserCoinsRequest{
		UserId: userMission.UserID,
		Coins:  int32(userMission.Mission.Reward),
	}

	_, addUserCoinSpan := s.tracer.Start(ctx, "Service.AddUserCoins")
	_, err = s.authServiceClient.AddUserCoins(ctx, req)
	if err != nil {
		tx.Rollback()
		addUserCoinSpan.RecordError(err)
		addUserCoinSpan.End()
		st, ok := status.FromError(err)
		if ok {
			return s.newInternalError(err)
		}
		return s.newErrorWithCLientMessage(code.Code_INTERNAL, err, st.Message())
	}
	addUserCoinSpan.End()

	err = tx.Commit()
	if err != nil {
		return s.newInternalError(err)
	}

	return nil
}
