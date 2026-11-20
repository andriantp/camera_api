package vault_test

import (
	"os"
	"testing"

	"camera_api/pkg/adapter/vault"
	"camera_api/pkg/utils/crypto"
)

/*
Integration test: needs a reachable Vault. Skipped when HOST or TOKEN is empty.

HOST= TOKEN= \
go test ./pkg/adapter/vault -run TestReadVault -v

Only non-secret values are logged; secrets are checked for presence, never printed.
*/
func TestReadVault(t *testing.T) {
	hc := vault.HashiCorp{
		Product: "camera",
		ENV:     "development",
		File:    "api",
		Host:    os.Getenv("HOST"),
		Token:   os.Getenv("TOKEN"),
	}

	if hc.Host == "" || hc.Token == "" {
		t.Skip("HOST/TOKEN not set; skipping Vault integration test")
	}

	setting, appSetting, err := vault.ReadVault(hc)
	if err != nil {
		t.Fatalf("error on ReadVault: %v", err)
	}

	required := map[string]string{
		"database.host":      setting.Database.Host,
		"database.name":      setting.Database.Name,
		"database.user":      setting.Database.User,
		"database.password":  setting.Database.Password,
		"token.secret":       setting.Token.Secret,
		"token.issuer":       setting.Token.Issuer,
		"camera.key":         setting.Camera.Key,
		"s3.endpoint":        setting.S3.Endpoint,
		"s3.bucketName":      setting.S3.BucketName,
		"s3.accessKeyID":     setting.S3.AccessKeyID,
		"s3.secretAccessKey": setting.S3.SecretAccessKey,
	}

	for field, value := range required {
		if value == "" {
			t.Errorf("%s is empty", field)
		}
	}

	if _, err := crypto.DecodeKey(setting.Camera.Key); err != nil {
		t.Errorf("camera.key: %v", err)
	}

	if appSetting.Server.Port == 0 {
		t.Error("server.port is empty")
	}

	if appSetting.Server.JWTEXPToken <= 0 || appSetting.Server.JWTEXPRefresh <= 0 {
		t.Error("server.jWTEXPToken and server.jWTEXPRefresh must be greater than 0")
	}

	t.Logf(
		"general: core=%d log=%s | server: port=%d cors=%t swagger=%t trustProxy=%t | db migrate=%t",
		setting.General.Core,
		setting.General.Log,
		appSetting.Server.Port,
		appSetting.Server.Cors,
		appSetting.Server.Swagger,
		appSetting.Server.TrustProxy,
		appSetting.Database.Migrate,
	)
}
