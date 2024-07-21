package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/olad5/caution-companion/internal/infra"
)

type FileService struct {
	fileStore infra.FileStore
}

var ErrInvalidIncidentType = errors.New("invalid incident_type")

func NewFileService(fileStore infra.FileStore) (*FileService, error) {
	if fileStore == nil {
		return &FileService{}, errors.New("FilesService failed to initialize, fileStore is nil")
	}
	return &FileService{fileStore}, nil
}

func (f *FileService) UploadFile(ctx context.Context, file io.Reader) (string, error) {
	filename := strings.ReplaceAll(uuid.New().String(), "-", "")
	fileUrl, err := f.fileStore.SaveToFileStore(ctx, filename, file)
	if err != nil {
		return "", fmt.Errorf("unable to save to file Store :%w", err)
	}
	return fileUrl, nil
}
