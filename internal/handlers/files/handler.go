package handlers

import (
	"errors"

	files "github.com/olad5/caution-companion/internal/usecases/files"
	"go.uber.org/zap"
)

type FilesHandler struct {
	fileService files.FileService
	logger      *zap.Logger
}

func NewFilesHandler(filesService files.FileService, logger *zap.Logger) (*FilesHandler, error) {
	if filesService == (files.FileService{}) {
		return nil, errors.New("filesService cannot be empty")
	}

	return &FilesHandler{filesService, logger}, nil
}
