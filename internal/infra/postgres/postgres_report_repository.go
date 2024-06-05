package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/olad5/caution-companion/internal/domain"
	"github.com/olad5/caution-companion/internal/infra"
)

type PostgresReportRepository struct {
	connection *sqlx.DB
}

func NewPostgresReportRepo(ctx context.Context, connection *sqlx.DB) (*PostgresReportRepository, error) {
	if connection == nil {
		return &PostgresReportRepository{}, fmt.Errorf("Failed to create PostgresFileRepository: connection is nil")
	}

	return &PostgresReportRepository{connection: connection}, nil
}

func (p *PostgresReportRepository) CreateReport(ctx context.Context, report domain.Report) error {
	const query = `
    INSERT INTO reports
      (id, owner_id, incident_type, longitude, latitude, description, created_at, updated_at) 
    VALUES 
    (:id, :owner_id, :incident_type, :longitude, :latitude, :description, :created_at, :updated_at)
  `

	_, err := p.connection.NamedExec(query, toSqlxReport(report))
	if err != nil {
		return fmt.Errorf("error creating report in the db: %w", err)
	}
	return nil
}

func (p *PostgresReportRepository) GetReportsByUserId(ctx context.Context, userId uuid.UUID, pageNumber, rowsPerPage int) ([]domain.Report, error) {
	offset := (pageNumber - 1) * rowsPerPage
	var reports []SqlxReport

	query := fmt.Sprintf(`
    SELECT * FROM reports WHERE owner_id = $1
    OFFSET %d ROWS FETCH NEXT %d ROWS ONLY
	`, offset, rowsPerPage)

	err := p.connection.Select(&reports, query, userId)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return []domain.Report{}, infra.ErrReportNotFound
		}
		return []domain.Report{}, fmt.Errorf("error getting reports by userId: %w", err)
	}

	result := []domain.Report{}
	for _, element := range reports {
		result = append(result, toReport(element))
	}

	return result, nil
}

func (p *PostgresReportRepository) GetLatestReports(ctx context.Context, pageNumber, rowsPerPage int) ([]domain.Report, error) {
	offset := (pageNumber - 1) * rowsPerPage
	var reports []SqlxReport

	// TODO:TODO: fix the time sorting here
	query := fmt.Sprintf(`
    SELECT * FROM reports 
    OFFSET %d ROWS FETCH NEXT %d ROWS ONLY
	`, offset, rowsPerPage)

	err := p.connection.Select(&reports, query)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return []domain.Report{}, infra.ErrReportNotFound
		}
		return []domain.Report{}, fmt.Errorf("error getting latest reports: %w", err)
	}

	result := []domain.Report{}
	for _, element := range reports {
		result = append(result, toReport(element))
	}

	return result, nil
}

func (p *PostgresReportRepository) GetReportByReportId(ctx context.Context, reportId uuid.UUID) (domain.Report, error) {
	var report SqlxReport

	err := p.connection.Get(&report, "SELECT * FROM reports WHERE id = $1", reportId)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return domain.Report{}, infra.ErrReportNotFound
		}
		return domain.Report{}, fmt.Errorf("error getting report by reportId: %w", err)
	}
	return toReport(report), nil
}

func (p *PostgresReportRepository) Ping(ctx context.Context) error {
	if err := p.connection.Ping(); err != nil {
		return fmt.Errorf("failed to ping postgres database: %w", err)
	}
	return nil
}

type SqlxReport struct {
	ID           uuid.UUID `db:"id"`
	OwnerID      uuid.UUID `db:"owner_id"`
	IncidentType string    `db:"incident_type"`
	Longitude    string    `db:"longitude"`
	Latitude     string    `db:"latitude"`
	Description  string    `db:"description"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func toReport(r SqlxReport) domain.Report {
	return domain.Report{
		ID:           r.ID,
		OwnerID:      r.OwnerID,
		IncidentType: r.IncidentType,
		Longitude:    r.Longitude,
		Latitude:     r.Latitude,
		Description:  r.Description,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func toSqlxReport(r domain.Report) SqlxReport {
	return SqlxReport{
		ID:           r.ID,
		IncidentType: r.IncidentType,
		Longitude:    r.Longitude,
		Latitude:     r.Latitude,
		Description:  r.Description,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
