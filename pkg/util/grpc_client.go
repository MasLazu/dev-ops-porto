package util

import (
	"context"
	"fmt"

	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCClient(ctx context.Context, addr string, logger monitoring.Logger) (*grpc.ClientConn, error) {
	logger.Info(ctx, "Connecting to gRPC server", log.String("address", addr))
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to connect to gRPC server: %v", err), log.String("address", addr))
		return nil, err
	}

	return conn, nil
}
