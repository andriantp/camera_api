package vault

import (
	"camera_api/app/usecase/usystem"
	"camera_api/pkg/utils/token"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"camera_api/app"
	"camera_api/pkg/adapter/camera"
	"camera_api/pkg/adapter/database/postgres"
	"camera_api/pkg/adapter/s3"
	"camera_api/pkg/utils/logger"

	"github.com/hashicorp/vault/api"
)

type Setting struct {
	General  General
	Database Database
	Server   app.Server
	Token    Token
	Camera   camera.Setting
	S3       s3.Setting
}

type General struct {
	Core int
	Log  string
}

type Database struct {
	Migrate  bool
	Host     string
	Port     uint16
	User     string
	Password string
	Name     string
}

type Token struct {
	Secret string
	Issuer string
}

type HashiCorp struct {
	Product string
	ENV     string
	File    string

	Host  string
	Token string
	Path  string
}

func ReadVault(s HashiCorp) (*Setting, *app.Setting, error) {
	hostVault, err := Unlock(s.Host)
	if err != nil {
		return nil, nil, fmt.Errorf("unlock vault host: %w", err)
	}

	tokenVault, err := Unlock(s.Token)
	if err != nil {
		return nil, nil, fmt.Errorf("unlock vault token: %w", err)
	}

	path := fmt.Sprintf(
		"%s/data/%s/%s",
		s.Product,
		s.ENV,
		s.File,
	)

	// Vault client configuration.
	config := api.DefaultConfig()
	config.Address = string(hostVault)

	// Create Vault client.
	client, err := api.NewClient(config)
	if err != nil {
		return nil, nil, fmt.Errorf("create vault client: %w", err)
	}

	// Set token authentication.
	client.SetToken(string(tokenVault))

	// Read secret.
	secret, err := client.Logical().Read(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read vault secret: %w", err)
	}

	// Check secret.
	if secret == nil {
		return nil, nil, errors.New("vault secret not found")
	}

	data, ok := secret.Data["data"]
	if !ok || data == nil {
		return nil, nil, errors.New("vault secret data not found")
	}

	// Convert Vault data to JSON.
	js, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal vault secret: %w", err)
	}

	// Convert JSON to application setting.
	var setting Setting

	if err := json.Unmarshal(js, &setting); err != nil {
		return nil, nil, fmt.Errorf("unmarshal vault setting: %w", err)
	}

	// Configure application logger.
	logger.Load(setting.General.Log)

	// Build PostgreSQL DSN.
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		setting.Database.Host,
		setting.Database.Port,
		setting.Database.User,
		setting.Database.Password,
		setting.Database.Name,
	)

	appSetting := &app.Setting{
		Build: usystem.Build{
			ENV: s.ENV,
		},

		Database: postgres.Setting{
			Migrate: setting.Database.Migrate,
			DSN:     dsn,
		},

		Server: app.Server{
			Port:          setting.Server.Port,
			ReadTimeout:   setting.Server.ReadTimeout * time.Second,
			WriteTimeout:  setting.Server.WriteTimeout * time.Minute,
			IdleTimeout:   setting.Server.IdleTimeout * time.Minute,
			JWTEXPToken:   setting.Server.JWTEXPToken * time.Minute,
			JWTEXPRefresh: setting.Server.JWTEXPRefresh * time.Hour,
			Cors:          setting.Server.Cors,
			AllowOrigins:  setting.Server.AllowOrigins,
			Swagger:       setting.Server.Swagger,
		},

		Token: token.Token{
			Secret: []byte(setting.Token.Secret),
			Issuer: setting.Token.Issuer,
		},

		Camera: camera.Setting{
			Download: setting.Camera.Download * time.Minute,
			Timeout:  setting.Camera.Timeout * time.Second,
			Key:      setting.Camera.Key,
		},

		S3: s3.Setting{
			Endpoint:        setting.S3.Endpoint,
			Region:          setting.S3.Region,
			Secure:          setting.S3.Secure,
			Token:           setting.S3.Token,
			AccessKeyID:     setting.S3.AccessKeyID,
			SecretAccessKey: setting.S3.SecretAccessKey,
			BucketName:      setting.S3.BucketName,
			Expire:          setting.S3.Expire * time.Minute,
		},
	}

	return &setting, appSetting, nil
}
