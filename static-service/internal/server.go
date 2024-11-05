package internal

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
)

func Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config, err := getConfig(ctx)
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

	httpServer := bootstrap(config, logger)

	return httpServer.Run(ctx)
}
