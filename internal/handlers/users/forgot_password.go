package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/caution-companion/internal/infra"
	appErrors "github.com/olad5/caution-companion/pkg/errors"
	response "github.com/olad5/caution-companion/pkg/utils"
	utils "github.com/olad5/caution-companion/pkg/utils/validation"
)

func (u UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}

	type requestDTO struct {
		Email string `json:"email" validate:"required,email"`
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

	err = u.userService.ForgotPassword(ctx, request.Email)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrUserNotFound):
			response.ErrorResponse(w, err.Error(), http.StatusNotFound)
			return
		default:
			response.InternalServerErrorResponse(w, err, u.logger)
			return
		}
	}

	response.SuccessResponse(w, "reset password email sent successfully", nil, u.logger)
}
