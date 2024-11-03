package server

import (
	"context"
	"fmt"
	"net"

	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/log"
	"google.golang.org/grpc"
)

type GrpcServerConfig struct {
	Port        int
	ServiceName string
}

type GrpcServer struct {
	config GrpcServerConfig
	Server *grpc.Server
	logger *monitoring.Logger
}

func NewGrpcServer(config GrpcServerConfig, logger *monitoring.Logger) *GrpcServer {
	return &GrpcServer{
		config: config,
		Server: grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler())),
		logger: logger,
	}
}

func (s *GrpcServer) RegisterService(sd *grpc.ServiceDesc, ss any) {
	s.Server.RegisterService(sd, ss)
}

func (s *GrpcServer) Run(ctx context.Context) (err error) {
	s.logger.Info(ctx, "Starting GRPC server", log.Int("port", s.config.Port))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		s.logger.Error(ctx, "Failed to listen on port", log.Int("port", s.config.Port))
		return
	}

	go func() {
		<-ctx.Done()
		s.Server.Stop()
	}()

	if err = s.Server.Serve(lis); err != nil {
		s.logger.Error(ctx, fmt.Sprintf("Failed to serve GRPC server: %v", err), log.Int("port", s.config.Port))
	}

	return
}
