package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/MasLazu/dev-ops-porto/mission-service/util"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/attribute"
)

type authMiddleware struct {
	responseWriter *util.ResponseWriter
	handlerTracer  *util.HandlerTracer
	jwtSecret      []byte
}

func NewAuthMiddleware(
	jwtSecret []byte,
	responseWriter *util.ResponseWriter,
	handlerTracer *util.HandlerTracer,
) *authMiddleware {
	return &authMiddleware{
		jwtSecret:      jwtSecret,
		responseWriter: responseWriter,
		handlerTracer:  handlerTracer,
	}
}

func (m *authMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.handlerTracer.TraceHttpHandler(r, "AuthMiddleware")
		defer span.End()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			span.SetAttributes(attribute.String("clientError.message", string("Authorization header is missing")))
			m.responseWriter.WriteUnauthorizedResponse(ctx, w)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			span.SetAttributes(attribute.String("clientError.message", string("Invalid Authorization header")))
			m.responseWriter.WriteUnauthorizedResponse(ctx, w)
			return
		}
		tokenStr := parts[1]

		token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})
		if err != nil {
			span.SetAttributes(attribute.String("clientError.message", err.Error()))
			m.responseWriter.WriteUnauthorizedResponse(ctx, w)
			return
		}

		ctx = context.WithValue(ctx, util.AuthContextKey, token.Claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
