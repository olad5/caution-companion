package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/olad5/go-hackathon-starter-template/internal/domain"
	"github.com/olad5/go-hackathon-starter-template/internal/infra"
)

type RedisAuthService struct {
	Cache     infra.Cache
	SecretKey string
}

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrExpiredToken    = errors.New("expired token")
	ErrGeneratingToken = errors.New("Error generating JWT token")
	ErrDecodingToken   = errors.New("error decoding JWT token")
)

const (
	JWT_HASH_NAME       = "go-starter-template-active-jwt-clients"
	SessionTTLInMinutes = 10
)

func NewRedisAuthService(ctx context.Context, cache infra.Cache, jwtSecretKey string) (*RedisAuthService, error) {
	if cache == nil {
		return nil, fmt.Errorf("failed to initialize auth service, cache is nil")
	}

	if err := cache.Ping(ctx); err != nil {
		return nil, err
	}

	return &RedisAuthService{cache, jwtSecretKey}, nil
}

func (r *RedisAuthService) GenerateJWT(ctx context.Context, user domain.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Minute * SessionTTLInMinutes).Unix(),
	})
	tokenString, err := token.SignedString([]byte(r.SecretKey))
	if err != nil {
		return "", ErrGeneratingToken
	}

	err = r.Cache.SetOne(ctx, constructUserIdKey(user.ID.String()), tokenString)
	if err != nil {
		return "", ErrGeneratingToken
	}
	return tokenString, nil
}

func (r *RedisAuthService) DecodeJWT(ctx context.Context, authHeader string) (JWTClaims, error) {
	const Bearer = "Bearer "
	var tokenString string
	if strings.HasPrefix(authHeader, Bearer) {
		tokenString = strings.TrimPrefix(authHeader, Bearer)
		if tokenString == "" {
			return JWTClaims{}, ErrInvalidToken
		}
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["sub"])
		}
		return []byte(r.SecretKey), nil
	})
	if err != nil {
		return JWTClaims{}, ErrDecodingToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return JWTClaims{}, ErrExpiredToken
		}

		var jwtClaims JWTClaims
		userId, ok := claims["sub"]
		if ok && userId != nil {
			jwtClaims.ID, err = uuid.Parse(userId.(string))
			if err != nil {
				return JWTClaims{}, ErrDecodingToken
			}
		}

		userEmail, ok := claims["email"]
		if ok && userEmail != nil {
			jwtClaims.Email = userEmail.(string)
		}

		return jwtClaims, nil
	}
	return JWTClaims{}, ErrInvalidToken
}

func (r *RedisAuthService) IsUserLoggedIn(ctx context.Context, authHeader, userId string) bool {
	token := strings.Split(authHeader, " ")[1]
	cachedToken, err := r.Cache.GetOne(ctx, constructUserIdKey(userId))
	if err != nil || cachedToken != token {
		return false
	}
	return true
}

func constructUserIdKey(key string) string {
	return JWT_HASH_NAME + key
}
