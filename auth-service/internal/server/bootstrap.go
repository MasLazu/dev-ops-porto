package server

import (
	"github.com/MasLazu/dev-ops-porto/auth-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/authservice"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"github.com/MasLazu/dev-ops-porto/pkg/server"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.opentelemetry.io/otel"
)

func bootstrap(config config, db *database.Service, logger *monitoring.Logger) (*server.HttpServer, *server.GrpcServer) {
	client := s3.NewFromConfig(config.aws.awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(config.aws.s3.enpoint)
	})

	tracer := otel.Tracer(config.serviceName)
	responseWriter := util.NewResponseWriter(tracer)
	requestDecoder := util.NewRequestBodyDecoder(tracer)
	validator := util.NewValidator(tracer)
	handlerTracer := util.NewHandlerTracer(tracer)
	authMiddleware := middleware.NewAuthMiddleware(config.jwtSecret, responseWriter, handlerTracer)
	repository := app.NewRepository(db, tracer)
	service := app.NewService(tracer, repository, config.jwtSecret, client, config.aws.s3.bucketNames.profilePictures, config.staticServiceEnpoint)
	handler := NewHttpHandler(service, authMiddleware, tracer, responseWriter, requestDecoder, validator, handlerTracer)
	grpcHandler := NewGrpcHandler(tracer, service)

	httpServer := server.NewHttpServer(server.HttpServerConfig{
		Port:        config.httpPort,
		ServiceName: config.serviceName,
	}, handler.setupRoutes, handlerTracer, responseWriter, logger)

	grpcServer := server.NewGrpcServer(server.GrpcServerConfig{
		Port:        config.grpcPort,
		ServiceName: config.serviceName,
	}, logger)

	authservice.RegisterAuthServiceServer(grpcServer.Server, grpcHandler)

	return httpServer, grpcServer
}
