package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"go.opentelemetry.io/otel/log"
)

func Run(ctx context.Context) (err error) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config, err := getConfig()
	if err != nil {
		return
	}

	otelShutdown, err := monitoring.SetupOTelSDK(ctx, monitoring.Config{
		ServiceName: config.serviceName,
		OtlpDomain:  config.otlpDomain,
	})
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	logger := monitoring.NewLogger(config.serviceName)

	logger.Info(ctx, "Connecting to database", log.String("host", config.database.Host), log.Int("port", config.database.Port))
	db, err := database.New(config.database)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to connect to database: %v", err), log.String("host", config.database.Host), log.Int("port", config.database.Port))
		return
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	httpServer, grpcServer := bootstrap(config, db, logger)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if httpError := httpServer.Run(ctx); httpError != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to run HTTP server: %v", err))
			err = errors.Join(err, httpError)
		}
	}()

	go func() {
		defer wg.Done()
		if grpcErr := grpcServer.Run(ctx); grpcErr != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to run GRPC server: %v", err))
			err = errors.Join(err, grpcErr)
		}
	}()

	wg.Wait()

	return
}
