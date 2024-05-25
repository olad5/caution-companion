package infra

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/olad5/go-hackathon-starter-template/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByUserId(ctx context.Context, userId uuid.UUID) (domain.User, error)
	Ping(ctx context.Context) error
}
