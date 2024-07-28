package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/caution-companion/internal/services/auth"
	"github.com/olad5/caution-companion/internal/usecases/users"
	appErrors "github.com/olad5/caution-companion/pkg/errors"
	response "github.com/olad5/caution-companion/pkg/utils"
	utils "github.com/olad5/caution-companion/pkg/utils/validation"
)

func (u UserHandler) VerifyResetPasswordToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}

	type requestDTO struct {
		Token string `json:"token" validate:"required,len=6"`
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

	err = u.userService.VerifyResetPasswordToken(ctx, request.Token)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrRetrievingPasswordResetToken):
			response.ErrorResponse(w, err.Error(), http.StatusUnauthorized)
			return
		case errors.Is(err, users.ErrInvalidToken):
			response.ErrorResponse(w, err.Error(), http.StatusUnauthorized)
			return
		default:
			response.InternalServerErrorResponse(w, err, u.logger)
			return
		}
	}

	response.SuccessResponse(w, "token verified successfully", nil, u.logger)
}
