package handlers

import (
	"errors"

	"github.com/olad5/caution-companion/internal/services/auth"
	"github.com/olad5/caution-companion/internal/usecases/users"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger      *zap.Logger
	userService users.UserService
	authService auth.AuthService
}

func NewUserHandler(logger *zap.Logger, userService users.UserService, authService auth.AuthService) (*UserHandler, error) {
	if userService == (users.UserService{}) {
		return nil, errors.New("user service cannot be empty")
	}
	if authService == nil {
		return nil, errors.New("auth service cannot be empty")
	}

	return &UserHandler{logger, userService, authService}, nil
}
