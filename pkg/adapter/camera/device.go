package camera

import (
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/response"
	"context"
	"net/http"

	"github.com/andriantp/camera/driver/isapi"
	"github.com/andriantp/camera/driver/onvif"
	cameraModel "github.com/andriantp/camera/model"
)

type PTZCapability struct {
	Protocol string                       `json:"protocol"`
	Onvif    *cameraModel.Configuration   `json:"onvif,omitempty"`
	Isapi    *cameraModel.PTZCapabilities `json:"isapi,omitempty"`
}

type Storage struct {
	Protocol string               `json:"protocol"`
	Isapi    *cameraModel.Storage `json:"isapi,omitempty"`
}

type Configuration struct {
	Protocol string                `json:"protocol"`
	Onvif    []cameraModel.Profile `json:"onvif,omitempty"`
	Isapi    []cameraModel.Input   `json:"isapi,omitempty"`
}

type Device interface {
	GetCapability(ctx context.Context, target CameraTarget) (*cameraModel.Capabilities, error)
	GetPTZCapability(ctx context.Context, target CameraTarget) (*PTZCapability, error)

	GetConfiguration(ctx context.Context, target CameraTarget) (any, error)
	GetStorage(ctx context.Context, target CameraTarget) (*Storage, error)
}

func (c *icamera) GetCapability(ctx context.Context, target CameraTarget) (*cameraModel.Capabilities, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	switch target.Protocol {
	case model.ProtocolONVIF:
		capabilities, err := c.clientOnvif.GetCapabilities(ctx,
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
		return capabilities, nil

	case model.ProtocolISAPI:
		capabilities, err := c.clientIsapi.GetCapabilities(ctx,
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
		return capabilities, nil

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}

func (c *icamera) GetPTZCapability(ctx context.Context, target CameraTarget) (*PTZCapability, error) {
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
		capabilities, err := c.clientOnvif.GetConfigurations(ctx,
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
		return &PTZCapability{
			Protocol: model.ProtocolONVIF,
			Onvif:    capabilities,
		}, nil

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

		return &PTZCapability{
			Protocol: model.ProtocolISAPI,
			Isapi:    capabilities,
		}, nil

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}

func (c *icamera) GetConfiguration(ctx context.Context, target CameraTarget) (any, error) {
	credential, err := c.DecryptCredential(target.Credential)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unable to decrypt credential",
		}
	}

	switch target.Protocol {
	case model.ProtocolONVIF:
		profiles, err := c.clientOnvif.GetProfiles(ctx,
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

		return Configuration{
			Protocol: model.ProtocolONVIF,
			Onvif:    profiles,
		}, nil

	case model.ProtocolISAPI:
		channels, err := c.clientIsapi.GetInputs(ctx,
			isapi.Camera{
				Credential: *credential,
			},
		)
		if err != nil {
			return nil, response.APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Unable to access camera",
			}
		}

		return Configuration{
			Protocol: model.ProtocolISAPI,
			Isapi:    channels,
		}, nil

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}

func (c *icamera) GetStorage(ctx context.Context, target CameraTarget) (*Storage, error) {
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
		storage, err := c.clientIsapi.GetStorage(ctx,
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

		return &Storage{
			Protocol: model.ProtocolISAPI,
			Isapi:    storage,
		}, nil

	default:
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}
}
