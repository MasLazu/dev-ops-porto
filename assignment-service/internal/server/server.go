package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/MasLazu/dev-ops-porto/pkg/genproto/missionservice"
	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"go.opentelemetry.io/otel/log"
)

func Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config, err := getConfig()
	if err != nil {
		return err
	}

	otelShutdown, err := monitoring.SetupOTelSDK(ctx, monitoring.Config{
		ServiceName: config.serviceName,
		OtlpDomain:  config.otlpDomain,
	})
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	logger := monitoring.NewLogger(config.serviceName)

	logger.Info(ctx, "Connecting to database", log.String("host", config.database.Host), log.Int("port", config.database.Port))
	db, err := database.New(config.database)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to connect to database: %v", err), log.String("host", config.database.Host), log.Int("port", config.database.Port))
		return err
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	missionServiceConn, err := util.NewGRPCClient(ctx, config.grpcMissionServiceDomain, logger)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to connect to gRPC server: %v", err), log.String("address", config.grpcMissionServiceDomain))
		return err
	}
	defer func() {
		err = errors.Join(err, missionServiceConn.Close())
	}()

	missionServiceClient := missionservice.NewMissionServiceClient(missionServiceConn)

	httpServer := bootstrap(config, db, logger, missionServiceClient)

	return httpServer.Run(ctx)
}
