package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	tracer                trace.Tracer
	repository            *Repository
	jwtSecret             []byte
	client                *s3.Client
	profilePicturesBucket string
	staticServiceEnpoint  string
}

func NewService(
	tracer trace.Tracer,
	repository *Repository,
	jwtSecret []byte,
	client *s3.Client,
	profilePicturesBucket string,
	staticServiceEnpoint string,
) *Service {
	return &Service{
		tracer:                tracer,
		repository:            repository,
		jwtSecret:             jwtSecret,
		client:                client,
		profilePicturesBucket: profilePicturesBucket,
		staticServiceEnpoint:  staticServiceEnpoint,
	}
}

func (s *Service) HealthCheck(ctx context.Context) map[string]string {
	ctx, span := s.tracer.Start(ctx, "Service.HealthCheck")
	defer span.End()

	response := s.repository.Health(ctx)

	return response
}

func (s *Service) Register(ctx context.Context, req RegisterUserRequest) (user, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.Register")
	defer span.End()

	if _, err := s.repository.FindUserByEmail(ctx, req.Email); err != sql.ErrNoRows {
		return user{}, NewClientError(http.StatusConflict, errors.New("user with this email already exists"))
	}

	_, hashSpan := s.tracer.Start(ctx, "hashing password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return user{}, NewInternalServiceError(err)
	}
	req.Password = string(hashedPassword)
	hashSpan.End()

	user, err := s.repository.InsertUser(ctx, req.toUser())
	if err != nil {
		return user, NewInternalServiceError(err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, req LoginUserRequest) (LoginResponse, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.Login")
	defer span.End()

	var res LoginResponse

	user, err := s.repository.FindUserByEmail(ctx, req.Email)
	if err == sql.ErrNoRows {
		return res, NewClientError(http.StatusUnauthorized, errors.New("email or password is invalid"))
	}

	if err != nil {
		return res, NewInternalServiceError(err)
	}

	_, hashSpan := s.tracer.Start(ctx, "comparing password")
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return res, NewClientError(http.StatusUnauthorized, errors.New("email or password is invalid"))
	}
	hashSpan.End()

	_, tokenSpan := s.tracer.Start(ctx, "generating jwt token")
	claims := &jwt.RegisteredClaims{
		Subject:   user.ID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30 * 12)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		tokenSpan.End()
		return res, NewInternalServiceError(err)
	}
	tokenSpan.End()

	return LoginResponse{AccessToken: signedToken}, nil
}

func (s *Service) Me(ctx context.Context, userID string) (user, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.Me")
	defer span.End()

	user, err := s.repository.FindUserByID(ctx, userID)
	if err != nil {
		return user, NewInternalServiceError(err)
	}

	user.addPrefixToProfilePictureURL(s.staticServiceEnpoint)

	return user, nil
}

func (s *Service) ChangeProfilePicture(ctx context.Context, userID string, file []byte) (user, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.ChangeProfilePicture")
	defer span.End()

	user, err := s.repository.FindUserByID(ctx, userID)
	if err != nil {
		return user, NewInternalServiceError(err)
	}

	if user.ProfilePicture != nil {
		_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &s.profilePicturesBucket,
			Key:    user.ProfilePicture,
		})
		if err != nil {
			return user, NewInternalServiceError(err)
		}
	}

	storeFileCtx, storeFileSpan := s.tracer.Start(ctx, "storing file")
	_, detectingContentTypeSpan := s.tracer.Start(storeFileCtx, "detecting content type")
	mimeType := http.DetectContentType(file)
	if !strings.HasPrefix(mimeType, "image/") {
		detectingContentTypeSpan.End()
		storeFileSpan.End()
		return user, nil
	}
	detectingContentTypeSpan.End()

	_, putFileToBucketSpan := s.tracer.Start(storeFileCtx, "putting file to bucket")
	key := userID + "/" + "profile_picture"
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.profilePicturesBucket,
		Key:         &key,
		Body:        strings.NewReader(string(file)),
		ContentType: &mimeType,
	})
	if err != nil {
		putFileToBucketSpan.End()
		storeFileSpan.End()
		return user, NewInternalServiceError(err)
	}
	putFileToBucketSpan.End()
	storeFileSpan.End()

	user.ProfilePicture = &key
	user, err = s.repository.UpdateUser(ctx, user)
	if err != nil {
		return user, NewInternalServiceError(err)
	}

	user.addPrefixToProfilePictureURL(s.staticServiceEnpoint)

	return user, nil
}

func (s *Service) DeleteProfilePicture(ctx context.Context, userID string) (user, ServiceError) {
	ctx, span := s.tracer.Start(ctx, "Service.DeleteProfilePicture")
	defer span.End()

	user, err := s.repository.FindUserByID(ctx, userID)
	if err != nil {
		return user, NewInternalServiceError(err)
	}

	if user.ProfilePicture == nil {
		return user, nil
	}

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.profilePicturesBucket,
		Key:    user.ProfilePicture,
	})
	if err != nil {
		return user, NewInternalServiceError(err)
	}

	user.ProfilePicture = nil
	err = s.repository.DeleteUserProfilePicture(ctx, user.ID)
	if err != nil {
		return user, NewInternalServiceError(err)
	}

	return user, nil
}
