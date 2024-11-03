package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MasLazu/dev-ops-porto/pkg/monitoring"
	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel/log"
)

type HttpServerConfig struct {
	Port        int
	ServiceName string
}

type HttpServer struct {
	config         HttpServerConfig
	router         func(handler *chi.Mux) http.Handler
	handlerTracer  *util.HandlerTracer
	responseWriter *util.ResponseWriter
	logger         *monitoring.Logger
}

func NewHttpServer(
	config HttpServerConfig,
	router func(handler *chi.Mux) http.Handler,
	handlerTracer *util.HandlerTracer,
	responseWriter *util.ResponseWriter,
	logger *monitoring.Logger,
) *HttpServer {
	return &HttpServer{
		config:         config,
		router:         router,
		handlerTracer:  handlerTracer,
		responseWriter: responseWriter,
		logger:         logger,
	}
}

func (s *HttpServer) setupRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(otelchi.Middleware(s.config.ServiceName, otelchi.WithChiRoutes(r)))
	r.NotFound(s.HandleNotFound)

	return s.router(r)
}

func (s *HttpServer) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.handlerTracer.TraceHttpHandler(r, "NotFoundHandler")
	defer span.End()

	s.responseWriter.WriteNotFoundResponse(ctx, w)
}

func (s *HttpServer) Run(ctx context.Context) (err error) {
	s.logger.Info(ctx, "Starting HTTP server", log.Int("port", s.config.Port))

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.setupRoutes(),
	}

	shutdownErrChan := make(chan error, 1)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			shutdownErrChan <- fmt.Errorf("failed to shutdown HTTP server: %w", shutdownErr)
			s.logger.Error(ctx, fmt.Sprintf("Failed to shutdown HTTP server: %v", shutdownErr), log.Int("port", s.config.Port))
		} else {
			shutdownErrChan <- nil
		}
	}()

	s.logger.Info(ctx, "HTTP server started", log.Int("port", s.config.Port))
	if err = server.ListenAndServe(); err != http.ErrServerClosed {
		err = fmt.Errorf("failed to start HTTP server: %w", err)
		s.logger.Error(ctx, fmt.Sprintf("Failed to start HTTP server: %v", err), log.Int("port", s.config.Port))
		return
	}

	if shutdownErr := <-shutdownErrChan; shutdownErr != nil {
		err = shutdownErr
	}

	return
}
