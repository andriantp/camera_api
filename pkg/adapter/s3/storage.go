package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
)

func (s *store) FindFile(ctx context.Context, objectName string) (*minio.ObjectInfo, error) {
	info, err := s.client.StatObject(
		ctx,
		s.setting.BucketName,
		objectName,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("StatObject: %w", err)
	}

	return &info, nil
}

func (s *store) Upload(ctx context.Context, contentType, objectName string, size int64, reader io.Reader) error {
	content := ""

	switch contentType {
	case "image":
		content = "image/png"
	case "video":
		content = "video/mp4"
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	_, err := s.client.PutObject(
		ctx,
		s.setting.BucketName,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: content,
		},
	)
	if err != nil {
		return fmt.Errorf("PutObject: %w", err)
	}

	return nil
}

func (s *store) GetURLDownload(ctx context.Context, objectName string) (*string, *time.Time, error) {
	fileName := path.Base(objectName)

	reqParams := make(url.Values)
	reqParams.Set(
		"response-content-disposition",
		`attachment; filename="`+fileName+`"`,
	)

	expires := time.Now().UTC().Add(s.setting.Expire)

	presignedURL, err := s.client.PresignedGetObject(
		ctx,
		s.setting.BucketName,
		objectName,
		s.setting.Expire,
		reqParams,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("PresignedGetObject: %w", err)
	}

	url := presignedURL.String()

	return &url, &expires, nil
}
