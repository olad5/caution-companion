package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	AvatarUrl string
	FirstName string
	LastName  string
	UserName  string
	Password  string
	Location  string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
