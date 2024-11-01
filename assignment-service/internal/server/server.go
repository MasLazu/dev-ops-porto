package server

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
)

func Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config, err := getConfig()
	if err != nil {
		return err
	}

	db, err := database.New(config.database)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

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

	httpServer := bootstrap(config, db)

	return httpServer.Run(ctx)
}
