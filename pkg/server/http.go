package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
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
}

func NewHttpServer(
	config HttpServerConfig,
	router func(handler *chi.Mux) http.Handler,
	handlerTracer *util.HandlerTracer,
	responseWriter *util.ResponseWriter,
) *HttpServer {
	return &HttpServer{
		config:         config,
		router:         router,
		handlerTracer:  handlerTracer,
		responseWriter: responseWriter,
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

func (s *HttpServer) Run(ctx context.Context) error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), s.setupRoutes())
}
