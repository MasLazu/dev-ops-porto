package server

import (
	"github.com/MasLazu/dev-ops-porto/assignment-service/internal/app"
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
	assignmentRepository := app.NewAssignmentRepository(db)
	reminderRepository := app.NewReminderRepository(db)
	handler := app.NewHandler(tracer, responseWriter, requestDecoder, validator, handlerTracer, repository, assignmentRepository, reminderRepository)
	authMiddleware := middleware.NewAuthMiddleware(config.jwtSecret, responseWriter, handlerTracer)
	router := NewRouter(handler, authMiddleware)

	return server.NewHttpServer(server.HttpServerConfig{
		Port:        config.port,
		ServiceName: config.serviceName,
	}, router.setupRoutes, handlerTracer, responseWriter)
}
