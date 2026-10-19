package camera

import (
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/logger"
	"context"
	"io"
	"net/http"

	"github.com/andriantp/camera/driver/isapi"
	cameraModel "github.com/andriantp/camera/model"
)

type Playlist struct {
	Protocol string                   `json:"protocol"`
	Content  cameraModel.SearchResult `json:"content"`
}
type Content interface {
	GetContents(ctx context.Context, target CameraTarget, req request.SearchContent) (*Playlist, error)
	FindContent(ctx context.Context, target CameraTarget, req request.SearchContent, filename string) (*cameraModel.Record, error)
	DownloadContent(ctx context.Context, target CameraTarget, playbackURI string) (*int64, io.ReadCloser, error)
	Snapshot(ctx context.Context, target CameraTarget, ss isapi.SnapshotOptions) (*int64, io.ReadCloser, error)
}

func (c *icamera) GetContents(ctx context.Context, target CameraTarget, req request.SearchContent) (*Playlist, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	c.setDefaults(&target)
	switch target.Protocol {
	case model.ProtocolISAPI:
		data, err := c.clientIsapi.ContentSearch(ctx,
			isapi.Camera{
				Credential: *credential,
			}, isapi.SearchRequest{
				SearchID:             *target.SearchID,
				TrackID:              *target.TrackID,
				StartTime:            req.From,
				EndTime:              req.To,
				MaxResults:           req.Limit,
				SearchResultPosition: req.Offset,
			},
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		return &Playlist{
			Protocol: model.ProtocolISAPI,
			Content:  *data,
		}, nil

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}

func (c *icamera) FindContent(ctx context.Context, target CameraTarget, req request.SearchContent, filename string) (*cameraModel.Record, error) {
	position := 0
	for {
		req.Offset = position
		data, err := c.GetContents(ctx, target, req)
		if err != nil {
			logger.Error("FindContent", err.Error())
			return nil, err
		}

		for _, content := range data.Content.Matches {
			if content.Filename != filename {
				continue
			}

			return &content, nil
		}

		position += len(data.Content.Matches)
		if position >= data.Content.TotalMatches {
			break
		}
	}

	return nil, response.APIError{
		StatusCode: http.StatusNotFound,
		Message:    "Content not found",
	}
}

func (c *icamera) DownloadContent(ctx context.Context, target CameraTarget, playbackURI string) (*int64, io.ReadCloser, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	c.setDefaults(&target)
	switch target.Protocol {
	case model.ProtocolISAPI:
		size, data, err := c.clientIsapi.ContentDownload(ctx,
			isapi.Camera{
				Credential: *credential,
			}, playbackURI,
		)
		if err != nil {
			return nil, nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		return &size, data, nil

	default:
		return nil, nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}
func (c *icamera) Snapshot(ctx context.Context, target CameraTarget, ss isapi.SnapshotOptions) (*int64, io.ReadCloser, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	c.setDefaults(&target)
	switch target.Protocol {
	case model.ProtocolISAPI:
		size, data, err := c.clientIsapi.GetSnapshot(ctx,
			isapi.Camera{
				Credential: *credential,
				Channel:    *target.Channel,
			}, ss,
		)
		if err != nil {
			return nil, nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		return &size, data, nil

	default:
		return nil, nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}
