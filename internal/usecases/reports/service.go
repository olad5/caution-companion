package reports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/olad5/caution-companion/internal/domain"
	"github.com/olad5/caution-companion/internal/infra"
	"github.com/olad5/caution-companion/internal/services/auth"
	appErrors "github.com/olad5/caution-companion/pkg/errors"
	"go.uber.org/zap"
)

type ReportService struct {
	logger     *zap.Logger
	reportRepo infra.ReportRepository
}

var ErrInvalidIncidentType = errors.New("invalid incident_type")

func NewReportsService(l *zap.Logger, reportRepo infra.ReportRepository) (*ReportService, error) {
	if reportRepo == nil {
		return &ReportService{}, errors.New("ReportService failed to initialize, reportRepo is nil")
	}
	return &ReportService{logger: l, reportRepo: reportRepo}, nil
}

func (r *ReportService) CreateReport(
	ctx context.Context,
	incidentType, longitude, latitude, description string,
) (domain.Report, error) {
	jwtClaims, ok := auth.GetJWTClaims(ctx)
	if !ok {
		return domain.Report{}, fmt.Errorf("error parsing JWTClaims: %v", appErrors.ErrInvalidToken)
	}
	userId := jwtClaims.ID
	incidentTypes := []string{"robbery", "fire", "accident", "cult"}

	isIncidentTypeLegit := false
	for _, element := range incidentTypes {
		if element == incidentType {
			isIncidentTypeLegit = true
		}
	}

	if !isIncidentTypeLegit {
		return domain.Report{}, ErrInvalidIncidentType
	}

	newReport := domain.Report{
		ID:           uuid.New(),
		OwnerID:      userId,
		IncidentType: incidentType,
		Longitude:    longitude,
		Latitude:     latitude,
		Description:  description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := r.reportRepo.CreateReport(ctx, newReport)
	if err != nil {
		return domain.Report{}, err
	}
	return newReport, nil
}

func (r *ReportService) GetReportByReportId(
	ctx context.Context, reportId uuid.UUID,
) (domain.Report, error) {
	existingReport, err := r.reportRepo.GetReportByReportId(ctx, reportId)
	if err != nil {
		return domain.Report{}, err
	}
	return existingReport, nil
}

func (r *ReportService) GetLatestReports(
	ctx context.Context,
	pageNumber, rowsPerPage int,
) ([]domain.Report, error) {
	reports, err := r.reportRepo.GetLatestReports(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return []domain.Report{}, err
	}
	return reports, nil
}

func (r *ReportService) GetReportsByUserId(
	ctx context.Context, userId uuid.UUID, pageNumber, rowsPerPage int,
) ([]domain.Report, error) {
	reports, err := r.reportRepo.GetReportsByUserId(
		ctx, userId, pageNumber, rowsPerPage)
	if err != nil {
		return []domain.Report{}, err
	}
	return reports, nil
}
