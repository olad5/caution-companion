package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/go-hackathon-starter-template/internal/usecases/users"
	appErrors "github.com/olad5/go-hackathon-starter-template/pkg/errors"
	response "github.com/olad5/go-hackathon-starter-template/pkg/utils"
	utils "github.com/olad5/go-hackathon-starter-template/pkg/utils/validation"
)

func (u UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}

	type requestDTO struct {
		Email     string `json:"email" validate:"required,email"`
		FirstName string `json:"first_name" validate:"required,alpha"`
		LastName  string `json:"last_name" validate:"required,alpha"`
		Password  string `json:"password" validate:"required,gt=8"`
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

	newUser, err := u.userService.CreateUser(ctx, request.FirstName, request.LastName, request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserAlreadyExists):
			response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		default:
			response.InternalServerErrorResponse(w, err, u.logger)
			return
		}
	}

	response.SuccessResponse(w, "user created successfully", ToUserDTO(newUser), u.logger)
}
