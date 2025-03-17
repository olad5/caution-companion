package handlers

import (
	"errors"

	"github.com/olad5/caution-companion/internal/usecases/reports"
	"go.uber.org/zap"
)

type ReportsHandler struct {
	logger      *zap.Logger
	userService reports.ReportService
}

func NewReportsHandler(logger *zap.Logger, reportsService reports.ReportService) (*ReportsHandler, error) {
	if reportsService == (reports.ReportService{}) {
		return nil, errors.New("reports service cannot be empty")
	}

	return &ReportsHandler{logger, reportsService}, nil
}
