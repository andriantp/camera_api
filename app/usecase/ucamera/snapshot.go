package ucamera

import (
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/logger"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/andriantp/camera/driver/isapi"
)

type Snapshot interface {
	GetSnapshot(ctx context.Context, actor request.UserActor, cameraUUID string) (*string, *int64, io.ReadCloser, error)
	GetSnapshotS3(ctx context.Context, actor request.UserActor, cameraUUID string) (*response.Download, error)
}

func (u *usecase) GetSnapshot(ctx context.Context, actor request.UserActor, cameraUUID string) (*string, *int64, io.ReadCloser, error) {
	now := time.Now().UTC()
	_, size, body, err := u.getSnapshot(ctx, actor, cameraUUID)
	if err != nil {
		return nil, nil, nil, err
	}

	filename := now.Format("20060102T150405-0700") + ".jpg"
	msg := fmt.Sprintf(
		"Snapshot content %s from camera %s",
		filename,
		cameraUUID,
	)

	if err := u.createActivity(
		ctx,
		now,
		actor,
		cameraUUID,
		"SNAPSHOT_CAMERA",
		msg,
	); err != nil {
		body.Close()
		return nil, nil, nil, err
	}

	return &filename, size, body, nil
}

func (u *usecase) GetSnapshotS3(ctx context.Context, actor request.UserActor, cameraUUID string) (*response.Download, error) {
	now := time.Now().UTC()

	canonicalUUID, size, body, err := u.getSnapshot(ctx, actor, cameraUUID)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	filename := now.Format("20060102T150405-0700") + ".jpg"
	// Use the UUID from the database, not the raw path parameter, so the
	// object key is always canonical.
	objectName := path.Join(canonicalUUID, filename)
	if err := u.storage.Upload(
		ctx,
		"image",
		objectName,
		*size,
		body,
	); err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to upload storage",
		}
	}

	file, err := u.storage.FindFile(ctx, objectName)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "image not found in storage",
		}
	}

	resp, err := u.downloadAtS3(
		ctx,
		filename,
		objectName,
		file,
	)
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf(
		"Snapshot s3 %s from camera %s",
		filename,
		cameraUUID,
	)

	if err := u.createActivity(
		ctx,
		now,
		actor,
		cameraUUID,
		"SNAPSHOT_CAMERA_S3",
		msg,
	); err != nil {
		return nil, err
	}

	return resp, nil
}

// getSnapshot returns the camera's canonical UUID together with the image.
func (u *usecase) getSnapshot(ctx context.Context, actor request.UserActor, cameraUUID string) (string, *int64, io.ReadCloser, error) {
	row, err := u.findCamera(ctx, actor, "getSnapshot", cameraUUID)
	if err != nil {
		return "", nil, nil, err
	}

	target := cameraTarget(row)

	size, body, err := u.camera.Snapshot(ctx, target, isapi.SnapshotOptions{})
	if err != nil {
		logger.Error("getSnapshot-Snapshot", err.Error())
		return "", nil, nil, err
	}

	return row.UUID.String(), size, body, nil
}
