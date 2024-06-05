package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/olad5/caution-companion/internal/infra"
	appErrors "github.com/olad5/caution-companion/pkg/errors"
	response "github.com/olad5/caution-companion/pkg/utils"
)

func (rh ReportsHandler) GetReportByReportId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.ErrorResponse(w, appErrors.ErrInvalidID, http.StatusBadRequest)
	}

	report, err := rh.userService.GetReportByReportId(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrReportNotFound):
			response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		default:
			response.InternalServerErrorResponse(w, err, rh.logger)
			return
		}
	}

	response.SuccessResponse(w, "report retrieved successfully", ToReportDTO(report), rh.logger)
}
