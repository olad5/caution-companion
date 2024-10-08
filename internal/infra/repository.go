package infra

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/olad5/caution-companion/internal/domain"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrReportNotFound = errors.New("report not found")
)

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByUserId(ctx context.Context, userId uuid.UUID) (domain.User, error)
	GetUserByUserName(ctx context.Context, userName string) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
	Ping(ctx context.Context) error
}

type ReportRepository interface {
	CreateReport(ctx context.Context, report domain.Report) error
	GetReportsByUserId(ctx context.Context, userId uuid.UUID, pageNumber, rowsPerPage int) ([]domain.Report, error)
	GetLatestReports(ctx context.Context, pageNumber, rowsPerPage int) ([]domain.Report, error)
	GetReportByReportId(ctx context.Context, reportId uuid.UUID) (domain.Report, error)
}

type FileStore interface {
	SaveToFileStore(ctx context.Context, filename string, file io.Reader) (string, error)
}
