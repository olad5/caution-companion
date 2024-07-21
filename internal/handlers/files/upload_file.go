package handlers

import (
	"errors"
	"net/http"

	files "github.com/olad5/caution-companion/internal/usecases/files"
	response "github.com/olad5/caution-companion/pkg/utils"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 * 2 // 1MB * 2

func (f FilesHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		response.ErrorResponse(w, " file must not be more than 2MB.", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.ErrorResponse(w, "Error retrieving file, please try again", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ctx := r.Context()
	url, err := f.fileService.UploadFile(ctx, file)
	if err != nil {
		switch {
		case errors.Is(err, files.ErrInvalidFileType):
			response.ErrorResponse(w, "file is not an image", http.StatusBadRequest)
			return
		default:
			response.InternalServerErrorResponse(w, err, f.logger)
			return
		}
	}
	response.SuccessResponse(w, "image uploaded successfully",
		map[string]interface{}{
			"url": url,
		},
		f.logger)
}
