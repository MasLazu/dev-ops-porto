package server

import (
	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/authservice"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"github.com/MasLazu/dev-ops-porto/pkg/server"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/MasLazu/dev-ops-porto/theme-service/internal/app"
	"go.opentelemetry.io/otel"
)

func bootstrap(config config, db *database.Service, logger *monitoring.Logger, authServiceClient authservice.AuthServiceClient) *server.HttpServer {
	tracer := otel.Tracer(config.serviceName)
	responseWriter := util.NewResponseWriter(tracer)
	requestDecoder := util.NewRequestBodyDecoder(tracer)
	validator := util.NewValidator(tracer)
	handlerTracer := util.NewHandlerTracer(tracer)
	repository := app.NewRepository(db, tracer)
	service := app.NewService(tracer, repository, authServiceClient)
	authMiddleware := middleware.NewAuthMiddleware(config.jwtSecret, responseWriter, handlerTracer)
	httpHandler := NewHttpHandler(service, authMiddleware, handlerTracer, responseWriter, requestDecoder, validator)

	return server.NewHttpServer(server.HttpServerConfig{
		Port:        config.port,
		ServiceName: config.serviceName,
	}, httpHandler.setupRoutes, handlerTracer, responseWriter, logger)
}
