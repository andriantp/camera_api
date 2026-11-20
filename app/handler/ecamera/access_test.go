package ecamera

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"camera_api/pkg/api/response"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func TestCameraAccessRejectsInvalidUUID(t *testing.T) {
	tests := []struct {
		name               string
		cameraUUID, userID string
		wantMessage        string
	}{
		{name: "bad camera", cameraUUID: "not-a-uuid", userID: uuid.NewString(), wantMessage: "invalid camera UUID"},
		{name: "bad user", cameraUUID: uuid.NewString(), userID: "1; DROP TABLE users", wantMessage: "invalid user UUID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			c := e.NewContext(httptest.NewRequest(http.MethodDelete, "/", nil), httptest.NewRecorder())
			c.SetParamNames("cameraUUID", "userUUID")
			c.SetParamValues(tt.cameraUUID, tt.userID)

			// ucaseCamera is nil: the handlers must reject before calling it.
			for name, call := range map[string]func(echo.Context) error{
				"PUT":    (&handler{}).GrantCameraAccess,
				"DELETE": (&handler{}).RevokeCameraAccess,
			} {
				err := call(c)

				var apiErr response.APIError
				if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest || apiErr.Message != tt.wantMessage {
					t.Fatalf("%s: err = %v, want 400 %q", name, err, tt.wantMessage)
				}
			}
		})
	}
}
