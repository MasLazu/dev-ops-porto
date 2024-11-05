package internal

import (
	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"github.com/MasLazu/dev-ops-porto/pkg/server"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.opentelemetry.io/otel"
)

func bootstrap(config config, logger *monitoring.Logger) *server.HttpServer {
	s3Client := s3.NewFromConfig(config.aws.awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(config.aws.s3.enpoint)
	})

	tracer := otel.Tracer(config.serviceName)
	responseWriter := util.NewResponseWriter(tracer)
	handlerTracer := util.NewHandlerTracer(tracer)
	httpHandler := NewHttpHandler(tracer, responseWriter, handlerTracer, s3Client, config.aws.s3.bucketNames)

	return server.NewHttpServer(server.HttpServerConfig{
		Port:        config.port,
		ServiceName: config.serviceName,
	}, httpHandler.setupRoutes, handlerTracer, responseWriter, logger)
}
