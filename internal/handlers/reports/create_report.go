package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/caution-companion/internal/usecases/reports"
	appErrors "github.com/olad5/caution-companion/pkg/errors"
	response "github.com/olad5/caution-companion/pkg/utils"
	utils "github.com/olad5/caution-companion/pkg/utils/validation"
)

func (rh ReportsHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}

	type location struct {
		Longitude string `json:"longitude" validate:"required,latitude"`
		Latitude  string `json:"latitude" validate:"required,longitude"`
	}

	type requestDTO struct {
		IncidentType string   `json:"incident_type" validate:"required"`
		Location     location `json:"location" validate:"required"`
		Description  string   `json:"description" validate:"required"`
	}

	request, err := response.Decode[requestDTO](r)
	if err != nil {
		response.ErrorResponse(w, appErrors.ErrInvalidJson, http.StatusBadRequest)
		return
	}

	err = utils.Check(request)
	if err != nil {
		response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	newReport, err := rh.userService.CreateReport(ctx, request.IncidentType, request.Location.Longitude, request.Location.Latitude, request.Description)
	if err != nil {
		switch {

		case errors.Is(err, reports.ErrInvalidIncidentType):
			response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		default:
			response.InternalServerErrorResponse(w, err, rh.logger)
			return
		}
	}

	response.SuccessResponse(w, "report created successfully", ToReportDTO(newReport), rh.logger)
}
