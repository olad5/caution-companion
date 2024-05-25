package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/go-hackathon-starter-template/internal/infra"
	response "github.com/olad5/go-hackathon-starter-template/pkg/utils"
)

func (u UserHandler) GetLoggedInUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := u.userService.GetLoggedInUser(ctx)
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

	response.SuccessResponse(w, "user retrieved successfully", ToUserDTO(user), u.logger)
}
