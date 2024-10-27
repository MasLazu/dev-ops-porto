package server

import (
	"assignment-service/internal/app"
	"assignment-service/internal/database"
	"assignment-service/internal/util"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"go.opentelemetry.io/otel"
)

type Server struct {
	db             *database.Service
	config         config
	handlerTracer  *util.HandlerTracer
	handler        *app.Handler
	responseWriter *util.ResponseWriter
	authMiddleware *authMiddleware
}

func NewServer(
	config config,
	db *database.Service,
	responseWriter *util.ResponseWriter,
	handlerTracer *util.HandlerTracer,
	handler *app.Handler,
	authauthMiddleware *authMiddleware,
) *Server {
	return &Server{
		db:             db,
		config:         config,
		responseWriter: responseWriter,
		handlerTracer:  handlerTracer,
		handler:        handler,
		authMiddleware: authauthMiddleware,
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

	tracer := otel.Tracer("assignment-service")
	responseWriter := util.NewResponseWriter(tracer)
	requestDecoder := util.NewRequestBodyDecoder(tracer)
	validator := util.NewValidator(tracer)
	handlerTracer := util.NewHandlerTracer(tracer)
	repository := app.NewRepository(db)
	assignmentRepository := app.NewAssignmentRepository(db)
	reminderRepository := app.NewReminderRepository(db)
	handler := app.NewHandler(tracer, responseWriter, requestDecoder, validator, handlerTracer, repository, assignmentRepository, reminderRepository)
	authMiddleware := NewAuthMiddleware(config.jwtSecret, responseWriter, handlerTracer)
	server := NewServer(config, db, responseWriter, handlerTracer, handler, authMiddleware)

	http.ListenAndServe(fmt.Sprintf(":%d", config.port), server.setupRoutes())

	return nil
}
