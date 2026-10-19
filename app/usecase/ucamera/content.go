package ucamera

import (
	adaptercamera "camera_api/pkg/adapter/camera"
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/helper"
	"camera_api/pkg/utils/logger"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"path"
	"time"

	"github.com/andriantp/camera"
	cameraModel "github.com/andriantp/camera/model"
	"github.com/minio/minio-go/v7"
)

type Content interface {
	GetContents(ctx context.Context, actor request.UserActor, cameraUUID string, req request.SearchContent) (*adaptercamera.Playlist, error)
	FindContent(ctx context.Context, actor request.UserActor, cameraUUID, filename string) (*cameraModel.Record, error)
	DownloadContent(ctx context.Context, actor request.UserActor, cameraUUID, filename string) (*int64, io.ReadCloser, error)
	DownloadContentAtS3(ctx context.Context, actor request.UserActor, cameraUUID, filename string) (*response.Download, error)
}

func (u *usecase) GetContents(ctx context.Context, actor request.UserActor, cameraUUID string, req request.SearchContent) (*adaptercamera.Playlist, error) {
	start, end, loc, err := helper.ConvertTime(req.From, req.To)
	if err != nil {
		logger.Error("GetContents-ConvertTime", err.Error())
		return nil, err
	}
	req.From = *start
	req.To = *end

	row, err := u.findCamera(ctx, actor, "GetContents", cameraUUID)
	if err != nil {
		return nil, err
	}

	data, err := u.camera.GetContents(ctx,
		adaptercamera.CameraTarget{
			Credential: camera.Credential{
				Host:     row.Host,
				Username: row.Username,
				Password: row.Password,
			},
			Protocol: row.Protocol.Name,
			SearchID: row.SearchID,
			TrackID:  row.TrackID,
		}, req,
	)
	if err != nil {
		logger.Error("GetContents-Camera", err.Error())
		return nil, err
	}

	for i := range data.Content.Matches {
		match := &data.Content.Matches[i]

		match.Timeframe.Start = match.Timeframe.Start.In(loc)
		match.Timeframe.End = match.Timeframe.End.In(loc)

		match.PlaybackURI = nil
	}

	return data, nil
}

func (u *usecase) FindContent(ctx context.Context, actor request.UserActor, cameraUUID, filename string) (*cameraModel.Record, error) {
	start, end, err := helper.ParseContentFilename(filename)
	if err != nil {
		logger.Error(
			"FindContent-ParseContentFilename",
			fmt.Sprintf("filename[%s]:%s", filename, err.Error()),
		)

		return nil, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "filename is corrupted",
		}
	}

	row, err := u.findCamera(ctx, actor, "FindContent", cameraUUID)
	if err != nil {
		return nil, err
	}

	target := adaptercamera.CameraTarget{
		Credential: camera.Credential{
			Host:     row.Host,
			Username: row.Username,
			Password: row.Password,
		},
		Protocol: row.Protocol.Name,
		SearchID: row.SearchID,
		TrackID:  row.TrackID,
	}
	content, err := u.camera.FindContent(ctx,
		target,
		request.SearchContent{
			From:   start,
			To:     end,
			Limit:  100,
			Offset: 0,
		}, filename,
	)
	if err != nil {
		logger.Error("FindContent-FindContent", err.Error())
		return nil, err
	}

	content.PlaybackURI = nil

	return content, nil
}

func (u *usecase) DownloadContent(ctx context.Context, actor request.UserActor, cameraUUID, filename string) (*int64, io.ReadCloser, error) {
	start, end, err := helper.ParseContentFilename(filename)
	if err != nil {
		logger.Error(
			"DownloadContent-ParseContentFilename",
			fmt.Sprintf("filename[%s]:%s", filename, err.Error()),
		)

		return nil, nil, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "filename is corrupted",
		}
	}

	row, err := u.findCamera(ctx, actor, "DownloadContent", cameraUUID)
	if err != nil {
		return nil, nil, err
	}

	target := adaptercamera.CameraTarget{
		Credential: camera.Credential{
			Host:     row.Host,
			Username: row.Username,
			Password: row.Password,
		},
		Protocol: row.Protocol.Name,
		SearchID: row.SearchID,
		TrackID:  row.TrackID,
	}
	content, err := u.camera.FindContent(ctx,
		target,
		request.SearchContent{
			From:   start,
			To:     end,
			Limit:  100,
			Offset: 0,
		}, filename,
	)
	if err != nil {
		logger.Error("DownloadContent-FindContent", err.Error())
		return nil, nil, err
	}

	if content.PlaybackURI == nil {
		return nil, nil, response.APIError{
			StatusCode: http.StatusBadGateway,
			Message:    "Playback URI is not available",
		}
	}

	size, body, err := u.camera.DownloadContent(ctx, target, *content.PlaybackURI)
	if err != nil {
		logger.Error("DownloadContent-Download", err.Error())
		return nil, nil, err
	}

	now := time.Now().UTC()

	msg := fmt.Sprintf(
		"Downloaded content %s from camera %s",
		filename,
		cameraUUID,
	)

	if err := u.createActivity(
		ctx,
		now,
		actor,
		cameraUUID,
		"DOWNLOAD_CAMERA_CONTENT",
		msg,
	); err != nil {
		return nil, nil, err
	}

	return size, body, nil
}

func (u *usecase) DownloadContentAtS3(ctx context.Context, actor request.UserActor, cameraUUID, filename string) (*response.Download, error) {
	row, err := u.findCamera(ctx, actor, "DownloadContentAtS3", cameraUUID)
	if err != nil {
		return nil, err
	}

	objectName := path.Join(row.UUID.String(), filename)

	file, err := u.storage.FindFile(ctx, objectName)
	if err != nil {
		file, err = u.downloadContentToS3(
			ctx,
			row,
			cameraUUID,
			filename,
			objectName,
		)
		if err != nil {
			return nil, err
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

	now := time.Now().UTC()
	msg := fmt.Sprintf(
		"Downloaded content s3 %s from camera %s",
		filename,
		cameraUUID,
	)

	if err := u.createActivity(
		ctx,
		now,
		actor,
		cameraUUID,
		"DOWNLOAD_CAMERA_CONTENT_S3",
		msg,
	); err != nil {
		return nil, err
	}

	return resp, nil
}

func (u *usecase) downloadContentToS3(ctx context.Context, row *model.Camera, cameraUUID, filename, objectName string) (*minio.ObjectInfo, error) {
	start, end, err := helper.ParseContentFilename(filename)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "filename is corrupted",
		}
	}

	target := adaptercamera.CameraTarget{
		Credential: camera.Credential{
			Host:     row.Host,
			Username: row.Username,
			Password: row.Password,
		},
		Protocol: row.Protocol.Name,
		SearchID: row.SearchID,
		TrackID:  row.TrackID,
	}

	content, err := u.camera.FindContent(
		ctx,
		target,
		request.SearchContent{
			From:   start,
			To:     end,
			Limit:  100,
			Offset: 0,
		},
		filename,
	)
	if err != nil {
		return nil, err
	}

	if content.PlaybackURI == nil {
		return nil, response.APIError{
			StatusCode: http.StatusBadGateway,
			Message:    "Playback URI is not available",
		}
	}

	size, body, err := u.camera.DownloadContent(
		ctx,
		target,
		*content.PlaybackURI,
	)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if err := u.storage.Upload(
		ctx,
		"video",
		objectName,
		*size,
		body,
	); err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to upload storage",
		}
	}

	return u.storage.FindFile(ctx, objectName)
}

func (u *usecase) downloadAtS3(ctx context.Context, filename, objectName string, fileExist *minio.ObjectInfo) (*response.Download, error) {
	url, expires, err := u.storage.GetURLDownload(
		ctx,
		objectName,
	)
	if err != nil {
		logger.Error(
			"DownloadContentAtS3-GetURLDownload",
			err.Error(),
		)

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate URL download",
		}
	}

	sizeMB := math.Round(float64(fileExist.Size)/(1024*1024)*100) / 100

	return &response.Download{
		Filename:    filename,
		Size:        sizeMB,
		ContentType: fileExist.ContentType,
		Expires:     *expires,
		URL:         *url,
	}, nil
}
