package server

import (
	"net/http"

	"github.com/MasLazu/dev-ops-porto/mission-service/internal/app"
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
	c.Get("/health", r.handler.HealthCheck)

	c.Group(func(c chi.Router) {
		c.Use(r.authMiddleware.Auth)
		c.Get("/", r.handler.GetUserMissions)
		c.Get("/expiration", r.handler.GetUserExpirationMissionDate)
	})

	return c
}
