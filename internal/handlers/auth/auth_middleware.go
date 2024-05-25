package auth

import (
	"net/http"

	"github.com/olad5/go-hackathon-starter-template/internal/services/auth"
	appErrors "github.com/olad5/go-hackathon-starter-template/pkg/errors"
	response "github.com/olad5/go-hackathon-starter-template/pkg/utils"
)

func EnsureAuthenticated(authService auth.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")

			jwtClaims, err := authService.DecodeJWT(ctx, authHeader)
			if err != nil {
				response.ErrorResponse(w, appErrors.ErrUnauthorized, http.StatusUnauthorized)
				return
			}

			userId := jwtClaims.ID.String()
			if isUserLoggedIn := authService.IsUserLoggedIn(ctx, authHeader, userId); !isUserLoggedIn {
				response.ErrorResponse(w, appErrors.ErrUnauthorized, http.StatusUnauthorized)
				return
			}

			ctx = auth.SetJWTClaims(ctx, jwtClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
