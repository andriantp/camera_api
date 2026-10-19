package camera

import (
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/response"
	"context"
	"net/http"
	"time"

	"github.com/andriantp/camera/driver/isapi"
	"github.com/andriantp/camera/driver/onvif"
	cameraModel "github.com/andriantp/camera/model"
)

type Zoom struct {
	Current float64 `json:"current"`
	Range   Range   `json:"range"`
}

type Range struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type Command interface {
	GetPosition(ctx context.Context, target CameraTarget) (any, error)

	CalibrationMove(ctx context.Context, target CameraTarget, pantilt cameraModel.PanTilt, duration time.Duration) (*cameraModel.Status, error)
	Move(ctx context.Context, target CameraTarget, movement model.PanTilt) (*cameraModel.Status, error)

	GetPresets(ctx context.Context, target CameraTarget) ([]cameraModel.Preset, error)
	GotoPreset(ctx context.Context, target CameraTarget, id int) error

	GotoHome(ctx context.Context, target CameraTarget) error

	GetZoom(ctx context.Context, target CameraTarget) (*Zoom, error)
	SetZoom(ctx context.Context, target CameraTarget, zoom float64) (*float64, error)
}

func (c *icamera) GetPosition(ctx context.Context, target CameraTarget) (any, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	c.setDefaults(&target)
	switch target.Protocol {
	case model.ProtocolONVIF:
		status, err := c.clientOnvif.GetStatus(ctx,
			onvif.Camera{
				Credential: *credential,
			},
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		return status, nil

	case model.ProtocolISAPI:
		absolute, err := c.clientIsapi.GetAbsolute(ctx,
			isapi.Camera{
				Credential: *credential,
				Channel:    *target.Channel,
			},
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		return absolute, nil

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}

func (c *icamera) CalibrationMove(ctx context.Context, target CameraTarget, pantilt cameraModel.PanTilt, duration time.Duration) (*cameraModel.Status, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	c.setDefaults(&target)
	var resp *cameraModel.Status
	switch target.Protocol {
	case model.ProtocolONVIF:
		status, err := c.clientOnvif.RelativeMove(ctx,
			onvif.Camera{
				Credential: *credential,
				Profile:    *target.Profile,
			},
			pantilt,
			duration,
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}
		resp = status

	case model.ProtocolISAPI:
		status, err := c.clientIsapi.RelativeMove(ctx,
			isapi.Camera{
				Credential: *credential,
				Channel:    *target.Channel,
			},
			cameraModel.Speed{
				Pan:  int(pantilt.X),
				Tilt: int(pantilt.Y),
			},
			duration,
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		resp = status

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}

	return resp, nil
}

func (c *icamera) Move(ctx context.Context, target CameraTarget, movement model.PanTilt) (*cameraModel.Status, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	c.setDefaults(&target)
	var resp *cameraModel.Status
	switch target.Protocol {
	case model.ProtocolONVIF:
		status, err := c.clientOnvif.RelativeMove(ctx,
			onvif.Camera{
				Credential: *credential,
				Profile:    *target.Profile,
			},
			cameraModel.PanTilt{
				X: movement.Pan,
				Y: movement.Tilt,
			},
			movement.Duration,
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}
		resp = status

	case model.ProtocolISAPI:
		status, err := c.clientIsapi.RelativeMove(ctx,
			isapi.Camera{
				Credential: *credential,
				Channel:    *target.Channel,
			},
			cameraModel.Speed{
				Pan:  int(movement.Pan),
				Tilt: int(movement.Tilt),
			},
			movement.Duration,
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		resp = status

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}

	return resp, nil
}

func (c *icamera) GetPresets(ctx context.Context, target CameraTarget) ([]cameraModel.Preset, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	presets, err := c.clientIsapi.GetPresets(ctx,
		isapi.Camera{
			Credential: *credential,
			Channel:    *target.Channel,
		},
	)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusBadGateway,
			Message:    "Unable to access camera",
		}
	}

	return presets, nil
}

func (c *icamera) GotoPreset(ctx context.Context, target CameraTarget, id int) error {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	err = c.clientIsapi.GotoPreset(ctx,
		isapi.Camera{
			Credential: *credential,
			Channel:    *target.Channel,
		},
		id,
	)
	if err != nil {
		return response.APIError{
			StatusCode: http.StatusBadGateway,
			Message:    "Unable to access camera",
		}
	}

	return nil
}

func (c *icamera) GotoHome(ctx context.Context, target CameraTarget) error {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	err = c.clientIsapi.GotoHome(ctx,
		isapi.Camera{
			Credential: *credential,
			Channel:    *target.Channel,
		},
	)
	if err != nil {
		return response.APIError{
			StatusCode: http.StatusBadGateway,
			Message:    "Unable to access camera",
		}
	}

	return nil
}

func (c *icamera) GetZoom(ctx context.Context, target CameraTarget) (*Zoom, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	c.setDefaults(&target)
	switch target.Protocol {
	// case model.ProtocolONVIF:

	case model.ProtocolISAPI:
		capabilities, err := c.clientIsapi.GetPTZCapabilities(ctx,
			isapi.Camera{
				Credential: *credential,
				Channel:    *target.Channel,
			},
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		absolute, err := c.clientIsapi.GetAbsolute(ctx,
			isapi.Camera{
				Credential: *credential,
				Channel:    *target.Channel,
			},
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		return &Zoom{
			Current: absolute.Zoom,
			Range: Range{
				Min: capabilities.Zoom.Absolute.Min,
				Max: capabilities.Zoom.Absolute.Max,
			},
		}, nil

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}

func (c *icamera) SetZoom(ctx context.Context, target CameraTarget, zoom float64) (*float64, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	c.setDefaults(&target)
	switch target.Protocol {
	// case model.ProtocolONVIF:

	case model.ProtocolISAPI:
		absolute, err := c.clientIsapi.SetZoom(ctx,
			isapi.Camera{
				Credential: *credential,
				Channel:    *target.Channel,
			}, zoom,
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		return &absolute.Zoom, nil

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}
