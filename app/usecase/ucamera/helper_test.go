package ucamera

import (
	"errors"
	"net/http"
	"testing"
	"time"

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
