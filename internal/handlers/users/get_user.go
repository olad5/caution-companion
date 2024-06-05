package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/caution-companion/internal/usecases/users"

	"github.com/olad5/caution-companion/internal/infra"
	response "github.com/olad5/caution-companion/pkg/utils"
)

func (u UserHandler) GetLoggedInUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := u.userService.GetLoggedInUser(ctx)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrUserNotFound):
			response.ErrorResponse(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, users.ErrInvalidToken):
			response.ErrorResponse(w, err.Error(), http.StatusUnauthorized)
			return
		default:
			response.InternalServerErrorResponse(w, err, u.logger)
			return
		}
	}

	response.SuccessResponse(w, "user retrieved successfully", ToUserDTO(user), u.logger)
}
