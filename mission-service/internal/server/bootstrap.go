package server

import (
	"github.com/MasLazu/dev-ops-porto/mission-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/MasLazu/dev-ops-porto/pkg/server"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"go.opentelemetry.io/otel"
)

func bootstrap(config config, db *database.Service) *server.HttpServer {
	tracer := otel.Tracer(config.serviceName)

	responseWriter := util.NewResponseWriter(tracer)
	requestDecoder := util.NewRequestBodyDecoder(tracer)
	validator := util.NewValidator(tracer)
	handlerTracer := util.NewHandlerTracer(tracer)

	repository := app.NewRepository(db)
	userRepository := app.NewUserRepository(db)
	userMissionRepository := app.NewUserMissionRepository(db)
	missionRepository := app.NewMissionRepository(db)

	handler := app.NewHandler(
		tracer,
		responseWriter,
		requestDecoder,
		validator,
		handlerTracer,
		repository,
		userRepository,
		userMissionRepository,
		missionRepository,
	)

	authMiddleware := middleware.NewAuthMiddleware(config.jwtSecret, responseWriter, handlerTracer)

	router := NewRouter(handler, authMiddleware)

	return server.NewHttpServer(server.HttpServerConfig{
		Port:        config.port,
		ServiceName: config.serviceName,
	}, router.setupRoutes, handlerTracer, responseWriter)
}
