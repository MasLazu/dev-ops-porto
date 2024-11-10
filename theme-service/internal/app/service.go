package app

import (
	"context"
	"database/sql"

	serviceErr "github.com/MasLazu/dev-ops-porto/pkg/errors"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/authservice"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/rpc/code"
)

type Service struct {
	tracer            trace.Tracer
	repository        *Repository
	authServiceClient authservice.AuthServiceClient
}

func NewService(
	tracer trace.Tracer,
	repository *Repository,
	authServiceClient authservice.AuthServiceClient,
) *Service {
	return &Service{
		tracer:            tracer,
		repository:        repository,
		authServiceClient: authServiceClient,
	}
}

const signatureThemeID = 1

func (s *Service) HealthCheck(ctx context.Context) map[string]string {
	ctx, span := s.tracer.Start(ctx, "Service.HealthCheck")
	defer span.End()

	response := s.repository.Health(ctx)

	return response
}

func (s *Service) GetOwnedTheme(ctx context.Context, userID string) ([]Theme, error) {
	ctx, span := s.tracer.Start(ctx, "Service.getOwnedTheme")
	defer span.End()

	themes, err := s.repository.FindOwnedThemes(ctx, userID)
	if err == sql.ErrNoRows {
		return s.addNewUser(ctx, userID)
	} else if err != nil {
		return nil, err
	}

	return themes, nil
}

func (s *Service) addNewUser(ctx context.Context, userID string) ([]Theme, error) {
	if err := s.repository.InserOwnedTheme(ctx, userID, signatureThemeID); err != nil {
		return nil, err
	}
	return s.repository.FindOwnedThemes(ctx, userID)
}

func (s *Service) UnlockTheme(ctx context.Context, userID string, themeID int) serviceErr.ServiceError {
	ctx, span := s.tracer.Start(ctx, "Service.UnlockTheme")
	defer span.End()

	theme, err := s.repository.FindThemeByID(ctx, themeID)
	if err != nil {
		return serviceErr.NewServiceErrorWithClientMessage(code.Code_NOT_FOUND, err, "Theme not found")
	}

	_, err = s.authServiceClient.ReduceUserCoins(ctx, &authservice.UserCoinsRequest{
		UserId: userID,
		Coins:  int32(theme.Price),
	})
	if err != nil {
		return serviceErr.NewServiceError(code.Code_INTERNAL, err)
	}

	err = s.repository.InserOwnedTheme(ctx, userID, themeID)
	if err != nil {
		return serviceErr.NewServiceError(code.Code_INTERNAL, err)
	}

	return nil
}
