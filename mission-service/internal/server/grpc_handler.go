package server

import (
	"context"

	"github.com/MasLazu/dev-ops-porto/mission-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/missionservice"
	"go.opentelemetry.io/otel/trace"
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
		res.StatusCode = missionservice.StatusCode_STATUS_CODE_NOT_FOUND
		res.Message = "Unknown event"
		res.Error = &missionservice.Error{Message: "Unknown event"}
		return res, nil
	}

	err := h.service.TriggerMissionEvent(ctx, req.UserId, req.Event)
	if err != nil {
		res.StatusCode = missionservice.StatusCode_STATUS_CODE_INTERNAL_ERROR
		res.Message = "Internal error"
		res.Error = &missionservice.Error{Message: err.Error()}
		return res, nil
	}

	res.StatusCode = missionservice.StatusCode_STATUS_CODE_OK
	res.Message = "Event triggered successfully"
	return res, nil
}
