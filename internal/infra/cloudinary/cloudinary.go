package cloudinary

import (
	"context"
	"fmt"
	"io"

	"github.com/olad5/caution-companion/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryFileStore struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryFileStore(ctx context.Context, cfg *config.Configurations) (*CloudinaryFileStore, error) {
	cld, err := cloudinary.NewFromURL(cfg.CloudinaryUrl)
	if err != nil {
		return &CloudinaryFileStore{}, fmt.Errorf("failed to create a cloudinary client: %w", err)
	}
	return &CloudinaryFileStore{cld}, nil
}

func (c *CloudinaryFileStore) SaveToFileStore(ctx context.Context, filename string, file io.Reader) (string, error) {
	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:         "caution-companion" + "/avatars/" + filename,
		FilenameOverride: filename,
		Folder:           "caution-companion",
		UniqueFilename:   api.Bool(false),
		Overwrite:        api.Bool(true),
		Transformation:   "q_auto,c_fill,g_auto,h_200,w_200",
	})
	if err != nil {
		if err != nil {
			return "", fmt.Errorf("Unable to upload %s, %v", filename, err)
		}
	}

	return resp.SecureURL, nil
}
