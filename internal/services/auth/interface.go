package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/olad5/caution-companion/internal/domain"
)

type JWTClaims struct {
	ID    uuid.UUID
	Email string
}

type ctxKey int

const jwtKey ctxKey = 1

func SetJWTClaims(ctx context.Context, jwt JWTClaims) context.Context {
	return context.WithValue(ctx, jwtKey, jwt)
}

func GetJWTClaims(ctx context.Context) (JWTClaims, bool) {
	v, ok := ctx.Value(jwtKey).(JWTClaims)
	return v, ok
}

type AuthService interface {
	DecodeJWT(ctx context.Context, tokenString string) (JWTClaims, error)
	GenerateAuthTokens(ctx context.Context, user domain.User) (string, string, error)
	IsUserLoggedIn(ctx context.Context, authHeader, userId string) bool
	GetUserIdFromRefreshToken(ctx context.Context, refreshToken string) (uuid.UUID, error)
	LogUserOut(ctx context.Context, userId string) error
}
