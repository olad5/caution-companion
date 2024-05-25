package handlers

import (
	"github.com/olad5/go-hackathon-starter-template/internal/domain"
)

type UserDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func ToUserDTO(user domain.User) UserDTO {
	return UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}
