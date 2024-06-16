package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/caution-companion/internal/infra"
	"github.com/olad5/caution-companion/internal/services/auth"
	appErrors "github.com/olad5/caution-companion/pkg/errors"
	response "github.com/olad5/caution-companion/pkg/utils"
	utils "github.com/olad5/caution-companion/pkg/utils/validation"
)

func (u UserHandler) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}

	type requestDTO struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	request, err := response.Decode[requestDTO](r)
	if err != nil {
		response.ErrorResponse(w, appErrors.ErrInvalidJson, http.StatusBadRequest)
		return
	}

	err = utils.Check(request)
	if err != nil {
		response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := u.userService.RefreshUserAccessToken(ctx, request.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			response.ErrorResponse(w, appErrors.ErrUnauthorized, http.StatusUnauthorized)
			return
		case errors.Is(err, infra.ErrUserNotFound):
			response.ErrorResponse(w, appErrors.ErrUnauthorized, http.StatusUnauthorized)
			return
		default:
			response.InternalServerErrorResponse(w, err, u.logger)
			return
		}
	}

	response.SuccessResponse(w, "access token refreshed successfully",
		map[string]interface{}{"access_token": accessToken, "refresh_token": refreshToken},
		u.logger)
}
