package handlers

import (
	"errors"

	"github.com/olad5/caution-companion/internal/usecases/reports"
	"go.uber.org/zap"
)

type ReportsHandler struct {
	userService reports.ReportService
	logger      *zap.Logger
}

func NewReportsHandler(reportsService reports.ReportService, logger *zap.Logger) (*ReportsHandler, error) {
	if reportsService == (reports.ReportService{}) {
		return nil, errors.New("reports service cannot be empty")
	}

	return &ReportsHandler{reportsService, logger}, nil
}
