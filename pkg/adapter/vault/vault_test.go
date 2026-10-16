package vault_test

import (
	"os"
	"testing"

	jsoniter "github.com/json-iterator/go"

	"camera_api/pkg/adapter/vault"
)

/*
HOST= TOKEN= \
go test ./pkg/adapter/vault/vault_test.go -v
*/
func TestReadVault(t *testing.T) {
	hc := vault.HashiCorp{
		Product: "camera",
		ENV:     "development",
		File:    "api",
		Host:    os.Getenv("HOST"),
		Token:   os.Getenv("TOKEN"),
	}

	if hc.Host == "" {
		t.Fatal("HOST environment variable is required")
	}

	if hc.Token == "" {
		t.Fatal("TOKEN environment variable is required")
	}

	setting, appSetting, err := vault.ReadVault(hc)
	if err != nil {
		t.Fatalf("error on ReadVault: %v", err)
	}

	t.Logf(
		"database: host=%s port=%d name=%s user=%s migrate=%t",
		setting.Database.Host,
		setting.Database.Port,
		setting.Database.Name,
		setting.Database.User,
		setting.Database.Migrate,
	)

	t.Logf(
		"general: core=%d log=%s",
		setting.General.Core,
		setting.General.Log,
	)

	t.Logf(
		"server: port=%d cors=%t",
		setting.Server.Port,
		setting.Server.Cors,
	)

	json := jsoniter.Config{
		EscapeHTML:             false,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
	}.Froze()
	
	js, err := json.MarshalIndent(appSetting, "", "  ")
	if err != nil {
		t.Fatalf("error marshal app setting: %v", err)
	}

	t.Logf("app setting:\n%s", string(js))
}
