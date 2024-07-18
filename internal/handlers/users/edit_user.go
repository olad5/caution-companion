package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/caution-companion/internal/usecases/users"
	appErrors "github.com/olad5/caution-companion/pkg/errors"
	response "github.com/olad5/caution-companion/pkg/utils"
	utils "github.com/olad5/caution-companion/pkg/utils/validation"
)

func (u UserHandler) EditUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}

	type requestDTO struct {
		Email     string `json:"email" validate:"required,email"`
		FirstName string `json:"first_name" validate:"required,alpha"`
		LastName  string `json:"last_name" validate:"required,alpha"`
		UserName  string `json:"user_name" validate:"required,lte=12"`
		Phone     string `json:"phone" validate:"required,len=11"`
		Location  string `json:"location" validate:"required,gt=3"`
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

	updatedUser, err := u.userService.EditUser(
		ctx,
		request.FirstName,
		request.LastName,
		request.UserName,
		request.Email,
		request.Location,
		request.Phone)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserAlreadyExists):
			response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		case errors.Is(err, users.ErrUserNameAlreadyExists):
			response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		default:
			response.InternalServerErrorResponse(w, err, u.logger)
			return
		}
	}

	response.SuccessResponse(w, "user profile updated successfully", ToUserDTO(updatedUser), u.logger)
}
