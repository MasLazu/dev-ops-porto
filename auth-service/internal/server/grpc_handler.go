package server

import (
	"context"

	"github.com/MasLazu/dev-ops-porto/auth-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/authservice"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	tracer  trace.Tracer
	service *app.Service
	authservice.UnimplementedAuthServiceServer
}

func NewGrpcHandler(tracer trace.Tracer, service *app.Service) *GrpcHandler {
	return &GrpcHandler{
		tracer:  tracer,
		service: service,
	}
}

func (h *GrpcHandler) AddUserCoins(ctx context.Context, req *authservice.UserCoinsRequest) (*authservice.EmptyResponse, error) {
	_, span := h.tracer.Start(ctx, "GrpcHandler.AddUserCoins")
	defer span.End()

	if req.Coins < 0 {
		return nil, status.Error(codes.InvalidArgument, "coins cannot be negative")
	}

	_, err := h.service.AddCoin(ctx, req.UserId, req.Coins)
	if err != nil {
		return nil, status.Error(err.GrpcCode(), err.ClientMessage())
	}

	return &authservice.EmptyResponse{}, nil
}

func (h *GrpcHandler) ReduceUserCoins(ctx context.Context, req *authservice.UserCoinsRequest) (*authservice.EmptyResponse, error) {
	_, span := h.tracer.Start(ctx, "GrpcHandler.ReduceUserCoins")
	defer span.End()

	if req.Coins < 0 {
		return nil, status.Error(codes.InvalidArgument, "coins cannot be negative")
	}

	_, err := h.service.AddCoin(ctx, req.UserId, -req.Coins)
	if err != nil {
		return nil, status.Error(err.GrpcCode(), err.ClientMessage())
	}

	return &authservice.EmptyResponse{}, nil
}
