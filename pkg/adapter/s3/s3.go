package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Setting struct {
	Endpoint        string
	Region          string
	Secure          bool
	Token           string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Expire          time.Duration
}

type store struct {
	client  *minio.Client
	setting Setting
}

func NewStore(ctx context.Context, setting Setting) (Storage, error) {
	creds := credentials.NewStaticV4(
		setting.AccessKeyID,
		setting.SecretAccessKey,
		setting.Token,
	)

	client, err := minio.New(setting.Endpoint, &minio.Options{
		Creds:  creds,
		Secure: setting.Secure,
		Region: setting.Region,
	})
	if err != nil {
		return nil, err
	}

	if err := ensure(ctx, client, setting.Region, setting.BucketName); err != nil {
		return nil, err
	}

	return &store{
		client:  client,
		setting: setting,
	}, nil
}

type Storage interface {
	FindFile(ctx context.Context, objectName string) (*minio.ObjectInfo, error)
	Upload(ctx context.Context, contentType, objectName string, size int64, reader io.Reader) error
	GetURLDownload(ctx context.Context, objectName string) (*string, *time.Time, error)
}

func ensure(ctx context.Context, client *minio.Client, region, bucketName string) error {
	exist, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("Ensure: bucket %s existence check failed: %w", bucketName, err)
	}

	if !exist {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
			Region: region,
		})
		if err != nil {
			return fmt.Errorf("connect: failed to create bucket %s: %w", bucketName, err)
		}

	}

	return nil
}
