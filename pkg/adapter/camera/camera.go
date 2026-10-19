package camera

import (
	"camera_api/pkg/utils/crypto"
	"time"

	"github.com/andriantp/camera"
	"github.com/andriantp/camera/driver/isapi"
	"github.com/andriantp/camera/driver/onvif"
)

type CameraTarget struct {
	Credential camera.Credential
	Protocol   string
	Profile    *string
	Channel    *int
	SearchID   *int
	TrackID    *int
}

type DefaultConfig struct {
	Profile  string
	Channel  int
	SearchID int
	TrackID  int
}

type Setting struct {
	Timeout  time.Duration
	Download time.Duration
	Key      string
}

type icamera struct {
	key         []byte
	clientOnvif *onvif.Client
	clientIsapi *isapi.Client
	defaults    DefaultConfig
}

type Camera interface {
	Encrypt(plainText string) (string, error)
	EncryptCredential(credential camera.Credential) (*camera.Credential, error)
	DecryptCredential(credential camera.Credential) (*camera.Credential, error)

	Device
	Command
	Content
}

func NewCamera(setting Setting) (Camera, error) {
	key, err := crypto.DecodeKey(setting.Key)
	if err != nil {
		return nil, err
	}

	cfg := camera.Config{
		Timeout:         setting.Timeout,
		DownloadTimeout: setting.Download,
	}

	client := camera.New(cfg)

	return &icamera{
		key:         key,
		clientOnvif: onvif.New(client),
		clientIsapi: isapi.New(client),
		defaults: DefaultConfig{
			Profile:  "profile_1",
			Channel:  1,
			SearchID: 1,
			TrackID:  101,
		},
	}, nil
}

func (c *icamera) setDefaults(target *CameraTarget) {
	if target.Profile == nil {
		target.Profile = &c.defaults.Profile
	}

	if target.Channel == nil {
		target.Channel = &c.defaults.Channel
	}

	if target.SearchID == nil {
		target.SearchID = &c.defaults.SearchID
	}

	if target.TrackID == nil {
		target.TrackID = &c.defaults.TrackID
	}
}
