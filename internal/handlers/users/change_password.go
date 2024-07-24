package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/caution-companion/internal/usecases/users"
	appErrors "github.com/olad5/caution-companion/pkg/errors"
	response "github.com/olad5/caution-companion/pkg/utils"
	utils "github.com/olad5/caution-companion/pkg/utils/validation"
)

func (u UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}

	type requestDTO struct {
		OldPassword string `json:"old_password" validate:"required,gt=8"`
		NewPassword string `json:"new_password" validate:"required,gt=8"`
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

	if request.OldPassword == request.NewPassword {
		response.ErrorResponse(w, "passwords must not be the same", http.StatusBadRequest)
		return
	}

	err = u.userService.ChangePassword(ctx, request.OldPassword, request.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrPasswordIncorrect):
			response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		default:
			response.InternalServerErrorResponse(w, err, u.logger)
			return
		}
	}

	response.SuccessResponse(w, "password changed successfully", nil, u.logger)
}
