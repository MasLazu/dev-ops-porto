package server

import (
	"net/http"

	"github.com/MasLazu/dev-ops-porto/assignment-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	handler        *app.Handler
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(handler *app.Handler, authMiddleware *middleware.AuthMiddleware) *Router {
	return &Router{
		handler:        handler,
		authMiddleware: authMiddleware,
	}
}

func (r *Router) setupRoutes(c *chi.Mux) http.Handler {
	c.Use(r.authMiddleware.Auth)

	c.Get("/health", r.handler.HealthCheck)

	c.Post("/", r.handler.CreateAssignment)
	c.Get("/", r.handler.GetAssignments)
	c.Get("/{id}", r.handler.GetAssignmentByID)
	c.Put("/{id}", r.handler.UpdateAssignmentByID)

	return c
}
