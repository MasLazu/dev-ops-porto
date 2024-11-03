package server

import (
	"github.com/MasLazu/dev-ops-porto/mission-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"github.com/MasLazu/dev-ops-porto/pkg/server"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"go.opentelemetry.io/otel"
)

func bootstrap(config config, db *database.Service, logger *monitoring.Logger) (*server.HttpServer, *server.GrpcServer) {
	tracer := otel.Tracer(config.serviceName)

	responseWriter := util.NewResponseWriter(tracer)
	requestDecoder := util.NewRequestBodyDecoder(tracer)
	validator := util.NewValidator(tracer)
	handlerTracer := util.NewHandlerTracer(tracer)

	repository := app.NewRepository(db)
	userRepository := app.NewUserRepository(db)
	userMissionRepository := app.NewUserMissionRepository(db, tracer)
	missionRepository := app.NewMissionRepository(db, tracer)

	service := app.NewService(
		tracer,
		repository,
		userRepository,
		userMissionRepository,
		missionRepository,
	)

	authMiddleware := middleware.NewAuthMiddleware(config.jwtSecret, responseWriter, handlerTracer)

	httpHandler := NewHttpHandler(tracer, responseWriter, requestDecoder, validator, handlerTracer, service, authMiddleware)

	httpServer := server.NewHttpServer(server.HttpServerConfig{
		Port:        config.httpPort,
		ServiceName: config.serviceName,
	}, httpHandler.SetupRoutes, handlerTracer, responseWriter, logger)

	grpcServer := server.NewGrpcServer(server.GrpcServerConfig{
		Port:        config.grpcPort,
		ServiceName: config.serviceName,
	}, logger)

	return httpServer, grpcServer
}
