package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
)

func (s *Server) setupRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(otelchi.Middleware("auth-service", otelchi.WithChiRoutes(r)))

	r.Get("/health", s.handler.HealthCheckHandler)
	r.Post("/register", s.handler.Register)

	return r
}
