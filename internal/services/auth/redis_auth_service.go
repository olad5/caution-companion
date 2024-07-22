package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/olad5/caution-companion/internal/domain"
	"github.com/olad5/caution-companion/internal/infra"
)

type RedisAuthService struct {
	Cache     infra.Cache
	SecretKey string
}

var (
	ErrInvalidToken                 = errors.New("invalid token")
	ErrExpiredToken                 = errors.New("expired token")
	ErrGeneratingToken              = errors.New("Error generating JWT token")
	ErrGeneratingPasswordResetToken = errors.New("Error generating password reset token")
	ErrRetrievingPasswordResetToken = errors.New("Error retreiving password reset token")
	ErrDeletingPasswordResetToken   = errors.New("Error deleting password reset token")
	ErrDecodingToken                = errors.New("error decoding JWT token")
)

const (
	JWT_HASH_NAME           = "jwt-clients"
	refreshPrefix           = "refresh-"
	resetPrefix             = "reset-"
	keyDelimiter            = "--"
	colonDelimiter          = ":"
	AuthSessionTTLInMinutes = time.Minute * 30
	ResetTTLInMinutes       = time.Minute * 10
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

func (r *RedisAuthService) GenerateAuthTokens(ctx context.Context, user domain.User) (string, string, error) {
	err := r.deleteTokensTiedToUserId(ctx, user.ID.String())
	if err != nil {
		return "", "", fmt.Errorf("unable to delete existing accessTokens: %w", ErrGeneratingToken)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(AuthSessionTTLInMinutes).Unix(),
	})
	accessToken, err := token.SignedString([]byte(r.SecretKey))
	if err != nil {
		return "", "", ErrGeneratingToken
	}

	refreshToken := uuid.New().String()
	err = r.Cache.SetOne(ctx, constructKey(user.ID.String(), refreshToken), accessToken, AuthSessionTTLInMinutes)
	if err != nil {
		return "", "", ErrGeneratingToken
	}
	return accessToken, refreshToken, nil
}

func (r *RedisAuthService) getKeysTiedToAUserId(ctx context.Context, userId string) ([]string, error) {
	results, err := r.Cache.GetAllKeysUsingWildCard(ctx, "*"+userId)
	if err != nil {
		return []string{""}, infra.ErrUserNotFound
	}

	return results, nil
}

func (r *RedisAuthService) LogUserOut(ctx context.Context, userId string) error {
	return r.deleteTokensTiedToUserId(ctx, userId)
}

func (r *RedisAuthService) GetUserIdFromRefreshToken(ctx context.Context, refreshToken string) (uuid.UUID, error) {
	match := "*" + refreshPrefix + refreshToken + ":*"
	results, err := r.Cache.GetAllKeysUsingWildCard(ctx, match)
	if err != nil {
		return uuid.New(), ErrInvalidToken
	}
	if len(results) != 1 {
		return uuid.New(), ErrInvalidToken
	}

	first := results[0]

	elements := strings.Split(first, JWT_HASH_NAME+keyDelimiter)

	id, err := uuid.Parse(elements[1])
	if err != nil {
		return uuid.New(), ErrInvalidToken
	}

	return id, nil
}

func (r *RedisAuthService) extractTokensFromExistingValueInCache(ctx context.Context, userId string) (string, string, error) {
	results, err := r.getKeysTiedToAUserId(ctx, userId)
	if err != nil {
		return "", "", fmt.Errorf("Error extracting tokens from cache: %w", err)
	}

	if len(results) == 0 {
		return "", "", fmt.Errorf("Error extracting tokens from cache: No token found")
	}
	if len(results) != 1 {
		return "", "", fmt.Errorf("Error extracting tokens from cache")
	}

	first := results[0]

	acccessToken, err := r.Cache.GetOne(ctx, first)
	if err != nil {
		return "", "", fmt.Errorf("Error extracting tokens from cache: %w", err)
	}
	refreshToken, err := extractRefreshTokenFromKey(first)
	if err != nil {
		return "", "", fmt.Errorf("Error extracting tokens from cache: %w", err)
	}
	return acccessToken, refreshToken, nil
}

func extractRefreshTokenFromKey(token string) (string, error) {
	tokenSplits := strings.Split(token, refreshPrefix)
	if len(tokenSplits) == 1 && tokenSplits[0] == token {
		return "", fmt.Errorf("Error extracting refresh token from key: %w", ErrInvalidToken)
	}

	refreshSplits := strings.Split(tokenSplits[1], colonDelimiter+JWT_HASH_NAME)
	if len(refreshSplits) == 1 && refreshSplits[0] == tokenSplits[1] {
		return "", fmt.Errorf("Error extracting refresh token from key: %w", ErrInvalidToken)
	}

	return refreshSplits[0], nil
}

func (r *RedisAuthService) deleteTokensTiedToUserId(ctx context.Context, userId string) error {
	keys, err := r.getKeysTiedToAUserId(ctx, userId)
	if err != nil {
		return fmt.Errorf("Error deleting tokens tied to user: %w", err)
	}
	for _, key := range keys {
		err := r.Cache.DeleteOne(ctx, key)
		if err != nil {
			return fmt.Errorf("Error deleting tokens tied to user: %w", err)
		}
	}
	return nil
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

func (r *RedisAuthService) AddPasswordResetTokenToCache(
	ctx context.Context, userId uuid.UUID, token string,
) error {
	err := r.Cache.SetOne(ctx, constructPasswordResetKey(token), userId.String(), ResetTTLInMinutes)
	if err != nil {
		return ErrGeneratingPasswordResetToken
	}
	return nil
}

func (r *RedisAuthService) GetUserIdFromPasswordResetToken(ctx context.Context, token string) (string, error) {
	// TODO:TODO: I need a logger here, things can go wrong
	token, err := r.Cache.GetOne(ctx, constructPasswordResetKey(token))
	if err != nil {
		return "", ErrRetrievingPasswordResetToken
	}
	return token, nil
}

func (r *RedisAuthService) DeletePasswordResetToken(ctx context.Context, token string) error {
	err := r.Cache.DeleteOne(ctx, constructPasswordResetKey(token))
	if err != nil {
		return ErrDeletingPasswordResetToken
	}
	return nil
}

func (r *RedisAuthService) IsUserLoggedIn(ctx context.Context, authHeader, userId string) bool {
	existingAccesstoken := strings.Split(authHeader, " ")[1]
	cachedAccessToken, _, err := r.extractTokensFromExistingValueInCache(ctx, userId)
	if err != nil || cachedAccessToken != existingAccesstoken {
		return false
	}
	return true
}

func constructKey(userId, refreshToken string) string {
	return refreshPrefix + refreshToken + colonDelimiter + JWT_HASH_NAME + keyDelimiter + userId
}

func constructPasswordResetKey(token string) string {
	return resetPrefix + token
}
