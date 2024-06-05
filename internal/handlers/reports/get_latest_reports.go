package handlers

import (
	"net/http"

	apiUtils "github.com/olad5/caution-companion/pkg/utils"
)

func (rh ReportsHandler) GetLatestReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pageInfo, err := apiUtils.ParseRequest(r)
	if err != nil {
		apiUtils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
	}

	reports, err := rh.userService.GetLatestReports(ctx, pageInfo.Number, pageInfo.RowsPerPage)
	if err != nil {
		switch {
		default:
			apiUtils.InternalServerErrorResponse(w, err, rh.logger)
			return
		}
	}

	apiUtils.SuccessResponse(w, "latest reports retrieved successfully", ToReportsPagedDTO(reports), rh.logger)
}
