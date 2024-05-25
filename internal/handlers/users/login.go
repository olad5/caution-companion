package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/go-hackathon-starter-template/internal/infra"
	"github.com/olad5/go-hackathon-starter-template/internal/usecases/users"
	appErrors "github.com/olad5/go-hackathon-starter-template/pkg/errors"
	response "github.com/olad5/go-hackathon-starter-template/pkg/utils"
	utils "github.com/olad5/go-hackathon-starter-template/pkg/utils/validation"
)

func (u UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}

	type requestDTO struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,gt=8"`
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

	accessToken, err := u.userService.LogUserIn(ctx, request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrUserNotFound):
			response.ErrorResponse(w, "user does not exist", http.StatusNotFound)
			return
		case errors.Is(err, users.ErrPasswordIncorrect):
			response.ErrorResponse(w, "invalid credentials", http.StatusUnauthorized)
			return
		default:
			response.InternalServerErrorResponse(w, err, u.logger)
			return
		}
	}
	response.SuccessResponse(w, "user logged in successfully",
		map[string]interface{}{"access_token": accessToken},
		u.logger)
}
