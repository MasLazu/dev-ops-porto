package server

import (
	"context"

	"github.com/MasLazu/dev-ops-porto/mission-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/missionservice"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	tracer  trace.Tracer
	service *app.Service
	missionservice.UnimplementedMissionServiceServer
}

func NewGrpcHandler(tracer trace.Tracer, service *app.Service) *GrpcHandler {
	return &GrpcHandler{
		tracer:  tracer,
		service: service,
	}
}

func (h *GrpcHandler) TriggerMissionEvent(ctx context.Context, req *missionservice.TriggerMissionEventRequest) (*missionservice.TriggerMissionEventResponse, error) {
	_, span := h.tracer.Start(ctx, "GrpcHandler.TriggerMissionEvent")
	defer span.End()

	res := &missionservice.TriggerMissionEventResponse{}

	if req.Event == missionservice.TriggerMissionEvent_MISSION_EVENT_UNKNOWN {
		return res, status.Error(codes.InvalidArgument, "unknown mission event")
	}

	err := h.service.TriggerMissionEvent(ctx, req.UserId, req.Event)
	if err != nil {
		return res, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}
