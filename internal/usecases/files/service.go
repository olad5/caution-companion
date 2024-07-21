package files

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/olad5/caution-companion/internal/infra"
)

type FileService struct {
	fileStore infra.FileStore
}

var ErrInvalidFileType = errors.New("invalid filetype")

func NewFileService(fileStore infra.FileStore) (*FileService, error) {
	if fileStore == nil {
		return &FileService{}, errors.New("FilesService failed to initialize, fileStore is nil")
	}
	return &FileService{fileStore}, nil
}

func (f *FileService) UploadFile(ctx context.Context, file io.Reader) (string, error) {
	b, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("unable to save to file Store :%w", err)
	}
	mtype := mimetype.Detect(b)

	isImageMimeType := false
	for mtype := mtype; mtype != nil; mtype = mtype.Parent() {
		if mtype.Is("image/png") {
			isImageMimeType = true
		}
		if mtype.Is("image/jpeg") {
			isImageMimeType = true
		}
	}
	if !isImageMimeType {
		return "", fmt.Errorf("invalid mime type :%w", ErrInvalidFileType)
	}

	if !filetype.IsImage(b) {
		return "", fmt.Errorf("unable to save to file Store :%w", ErrInvalidFileType)
	}

	file = bytes.NewReader(b)
	filename := strings.ReplaceAll(uuid.New().String(), "-", "")
	fileUrl, err := f.fileStore.SaveToFileStore(ctx, filename, file)
	if err != nil {
		return "", fmt.Errorf("unable to save to file Store :%w", err)
	}
	return fileUrl, nil
}
