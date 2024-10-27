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
	r.Use(otelchi.Middleware("assignment-service", otelchi.WithChiRoutes(r)))
	r.Use(s.authMiddleware.Auth)
	r.NotFound(s.handler.NotFound)

	r.Get("/health", s.handler.HealthCheck)

	r.Post("/", s.handler.CreateAssignment)
	r.Get("/", s.handler.GetAssignments)
	r.Get("/{id}", s.handler.GetAssignmentByID)
	r.Put("/{id}", s.handler.UpdateAssignmentByID)

	return r
}
