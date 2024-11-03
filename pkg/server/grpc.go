package server

import (
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type GrpcServerConfig struct {
	Port        int
	ServiceName string
}

type GrpcServer struct {
	config GrpcServerConfig
	Server *grpc.Server
}

func NewGrpcServer(config GrpcServerConfig) *GrpcServer {
	return &GrpcServer{
		config: config,
		Server: grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler())),
	}
}

func (s *GrpcServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return err
	}

	return s.Server.Serve(lis)
}
