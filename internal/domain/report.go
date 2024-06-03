package domain

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID           uuid.UUID
	OwnerID      uuid.UUID
	IncidentType string
	Longitude    string
	Latitude     string
	Description  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
