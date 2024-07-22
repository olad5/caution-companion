package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/caution-companion/internal/infra"
	"github.com/olad5/caution-companion/internal/services/auth"
	"github.com/olad5/caution-companion/internal/usecases/users"
	appErrors "github.com/olad5/caution-companion/pkg/errors"
	response "github.com/olad5/caution-companion/pkg/utils"
	utils "github.com/olad5/caution-companion/pkg/utils/validation"
)

func (u UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}

	type requestDTO struct {
		Token           string `json:"token" validate:"required"`
		Password        string `json:"password" validate:"required,gt=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,gt=8"`
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

	if request.Password != request.ConfirmPassword {
		response.ErrorResponse(w, "passwords are not the same", http.StatusBadRequest)
		return
	}

	err = u.userService.ResetPassword(ctx, request.Token, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrUserNotFound):
			response.ErrorResponse(w, err.Error(), http.StatusNotFound)
			return
		case errors.Is(err, auth.ErrRetrievingPasswordResetToken):
			response.ErrorResponse(w, users.ErrInvalidToken.Error(), http.StatusUnauthorized)
			return
		default:
			response.InternalServerErrorResponse(w, err, u.logger)
			return
		}
	}

	response.SuccessResponse(w, "password reset successfully", nil, u.logger)
}
