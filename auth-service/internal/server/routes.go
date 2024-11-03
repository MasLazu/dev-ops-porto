package server

import (
	"net/http"

	"github.com/MasLazu/dev-ops-porto/auth-service/internal/app"
	"github.com/MasLazu/dev-ops-porto/pkg/middleware"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	handler        *app.Handler
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(handler *app.Handler, authMiddleware *middleware.AuthMiddleware) *Router {
	return &Router{
		handler, authMiddleware,
	}
}

func (r *Router) setupRoutes(c *chi.Mux) http.Handler {
	c.Get("/health", r.handler.HealthCheck)
	c.Post("/register", r.handler.Register)
	c.Post("/login", r.handler.Login)

	c.Group(func(c chi.Router) {
		c.Use(r.authMiddleware.Auth)
		c.Get("/me", r.handler.Me)
	})

	return c
}
