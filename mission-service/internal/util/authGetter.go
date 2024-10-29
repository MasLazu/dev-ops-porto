package util

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const AuthContextKey = contextKey("auth")

func GetAuthFromContext(ctx context.Context) (jwt.Claims, bool) {
	var claim jwt.Claims
	var ok bool
	claim, ok = ctx.Value(AuthContextKey).(jwt.Claims)

	return claim, ok
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	claim, ok := GetAuthFromContext(ctx)
	if !ok {
		return "", errors.New("auth claim not found")
	}

	sub, err := claim.GetSubject()

	return sub, err
}
