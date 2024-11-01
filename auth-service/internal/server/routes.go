package server

import (
	"net/http"

	"github.com/MasLazu/dev-ops-porto/auth-service/internal/app"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	handler *app.Handler
}

func NewRouter(handler *app.Handler) *Router {
	return &Router{
		handler: handler,
	}
}

func (r *Router) setupRoutes(c *chi.Mux) http.Handler {
	c.Get("/health", r.handler.HealthCheck)
	c.Post("/register", r.handler.Register)
	c.Post("/login", r.handler.Login)

	return c
}
