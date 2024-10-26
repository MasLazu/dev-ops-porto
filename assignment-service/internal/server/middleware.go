package server

import (
	"assignment-service/internal/util"
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type authMiddleware struct {
	responseWriter *util.ResponseWriter
	handlerTracer  *util.HandlerTracer
	jwtSecret      []byte
}

func NewAuthMiddleware(jwtSecret []byte) *authMiddleware {
	return &authMiddleware{jwtSecret: jwtSecret}
}

func (m *authMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.handlerTracer.TraceHttpHandler(r, "AuthMiddleware")
		defer span.End()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.responseWriter.WriteUnauthorizedResponse(ctx, w)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.responseWriter.WriteUnauthorizedResponse(ctx, w)
			return
		}
		tokenStr := parts[1]

		token, err := jwt.ParseWithClaims(tokenStr, jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})
		if err != nil {
			m.responseWriter.WriteUnauthorizedResponse(ctx, w)
			return
		}

		claims, ok := token.Claims.(jwt.RegisteredClaims)
		if !ok || !token.Valid {
			m.responseWriter.WriteUnauthorizedResponse(ctx, w)
			return
		}

		ctx = context.WithValue(ctx, util.AuthContextKey, claims.Subject)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
