package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/MasLazu/dev-ops-porto/auth-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/MasLazu/dev-ops-porto/pkg/util"

	"go.opentelemetry.io/otel"
)

type Server struct {
	db             *database.Service
	config         config
	handlerTracer  *util.HandlerTracer
	handler        *app.Handler
	responseWriter *util.ResponseWriter
}

func NewServer(
	config config,
	db *database.Service,
	responseWriter *util.ResponseWriter,
	handlerTracer *util.HandlerTracer,
	handler *app.Handler,
) *Server {
	return &Server{
		db:             db,
		config:         config,
		responseWriter: responseWriter,
		handlerTracer:  handlerTracer,
		handler:        handler,
	}
}

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

	otelShutdown, err := setupOTelSDK(ctx, config.otlpDomain)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	tracer := otel.Tracer("auth-service")
	responseWriter := util.NewResponseWriter(tracer)
	requestDecoder := util.NewRequestBodyDecoder(tracer)
	validator := util.NewValidator(tracer)
	handlerTracer := util.NewHandlerTracer(tracer)
	repository := app.NewRepository(db)
	handler := app.NewHandler(tracer, responseWriter, requestDecoder, validator, handlerTracer, repository, config.jwtSecret)
	server := NewServer(config, db, responseWriter, handlerTracer, handler)

	http.ListenAndServe(fmt.Sprintf(":%d", config.port), server.setupRoutes())

	return nil
}
