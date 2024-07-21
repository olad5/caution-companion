package handlers

import (
	"github.com/olad5/caution-companion/internal/domain"
)

type UserDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserName  string `json:"user_name"`
	Location  string `json:"location"`
	Phone     string `json:"phone"`
}

func ToUserDTO(user domain.User) UserDTO {
	return UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		Avatar:    user.AvatarUrl,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserName:  user.UserName,
		Location:  user.Location,
		Phone:     user.Phone,
	}
}
