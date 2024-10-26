package util

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const AuthContextKey = contextKey("auth")

func GetAuthFromContext(ctx context.Context) (jwt.RegisteredClaims, bool) {
	var claim jwt.RegisteredClaims
	var ok bool
	claim, ok = ctx.Value(AuthContextKey).(jwt.RegisteredClaims)

	return claim, ok
}
