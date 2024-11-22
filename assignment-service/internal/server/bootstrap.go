package server

import (
	"github.com/MasLazu/dev-ops-porto/assignment-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/missionservice"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"github.com/MasLazu/dev-ops-porto/pkg/server"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"go.opentelemetry.io/otel"
)

func bootstrap(config config, db *database.Service, logger *monitoring.Logger, missionServiceClient missionservice.MissionServiceClient) *server.HttpServer {
	tracer := otel.Tracer(config.serviceName)
	responseWriter := util.NewResponseWriter(tracer)
	requestDecoder := util.NewRequestBodyDecoder(tracer)
	validator := util.NewValidator(tracer)
	handlerTracer := util.NewHandlerTracer(tracer)
	repository := app.NewRepository(db)
	assignmentRepository := app.NewAssignmentRepository(db, tracer)
	reminderRepository := app.NewReminderRepository(db, tracer)
	service := app.NewService(tracer, repository, assignmentRepository, reminderRepository, missionServiceClient)
	authMiddleware := middleware.NewAuthMiddleware(config.jwtSecret, responseWriter, handlerTracer)
	httpHandler := NewHttpHandler(service, authMiddleware, handlerTracer, responseWriter, requestDecoder, validator)

	return server.NewHttpServer(server.HttpServerConfig{
		Port:        config.port,
		ServiceName: config.serviceName,
	}, httpHandler.setupRoutes, handlerTracer, responseWriter, logger)
}
