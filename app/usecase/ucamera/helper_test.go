package ucamera

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/response"
)

func TestParseMoveDuration(t *testing.T) {
	tests := []struct {
		value   string
		want    time.Duration
		wantErr bool
	}{
		{value: "500ms", want: 500 * time.Millisecond},
		{value: "3s", want: 3 * time.Second},
		{value: "10s", want: 10 * time.Second},
		{value: "10.001s", wantErr: true},
		{value: "10m", wantErr: true},
		{value: "1h", wantErr: true},
		{value: "0s", wantErr: true},
		{value: "-1s", wantErr: true},
		{value: "abc", wantErr: true},
		{value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := parseMoveDuration(tt.value)

			if tt.wantErr {
				var apiErr response.APIError
				if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
					t.Fatalf("err = %v, want 400 APIError", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseMoveDuration(%q): %v", tt.value, err)
			}

			if got != tt.want {
				t.Fatalf("parseMoveDuration(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestCameraTargetMapsEveryField(t *testing.T) {
	profile, channel, searchID, trackID := "profile_2", 3, 7, 201

	row := &model.Camera{
		Host:     "enc-host",
		Username: "enc-user",
		Password: "enc-pass",
		Protocol: model.Protocol{Name: model.ProtocolISAPI},
		Profile:  &profile,
		Channel:  &channel,
		SearchID: &searchID,
		TrackID:  &trackID,
	}

	got := cameraTarget(row)

	if got.Credential.Host != "enc-host" || got.Credential.Username != "enc-user" || got.Credential.Password != "enc-pass" {
		t.Errorf("credential = %+v", got.Credential)
	}

	if got.Protocol != model.ProtocolISAPI {
		t.Errorf("protocol = %q", got.Protocol)
	}

	if got.Profile != &profile || got.Channel != &channel || got.SearchID != &searchID || got.TrackID != &trackID {
		t.Error("optional fields must point at the row's values")
	}

	// Unset optional fields stay nil so the adapter applies its defaults.
	empty := cameraTarget(&model.Camera{})
	if empty.Profile != nil || empty.Channel != nil || empty.SearchID != nil || empty.TrackID != nil {
		t.Error("unset optional fields must stay nil")
	}
}
