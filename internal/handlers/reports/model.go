package handlers

import (
	"time"

	"github.com/olad5/caution-companion/internal/domain"
)

type location struct {
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
}

type ReportDTO struct {
	ID           string     `json:"id"`
	IncidentType string     `json:"incident_type"`
	Location     location   `json:"location"`
	Description  string     `json:"description"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

func ToReportDTO(report domain.Report) ReportDTO {
	return ReportDTO{
		ID:           report.ID.String(),
		IncidentType: report.IncidentType,
		Location: location{
			Longitude: report.Longitude,
			Latitude:  report.Latitude,
		},
		Description: report.Description,
		CreatedAt:   &report.CreatedAt,
		UpdatedAt:   &report.UpdatedAt,
	}
}

type ReportsPagedDTO struct {
	Rows  int         `json:"rows"`
	Page  int         `json:"page"`
	Items []ReportDTO `json:"items"`
}

func ToReportsPagedDTO(reports []domain.Report, page int) ReportsPagedDTO {
	items := []ReportDTO{}
	for _, report := range reports {
		items = append(items, ToReportDTO(report))
	}
	return ReportsPagedDTO{
		Page:  page,
		Rows:  len(items),
		Items: items,
	}
}
