package ucamera

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func revoke(u *usecase, cameraUUID, userUUID string) error {
	return u.RevokeCameraAccess(
		context.Background(),
		request.UserActor{UserID: "admin-1", SessionID: "session-1", Role: "admin"},
		cameraUUID,
		userUUID,
	)
}

func newRevoke(cams *fakeAssignCameraRepo, users *fakeAssignUserRepo) *usecase {
	if cams.camera == nil && cams.cameraErr == nil {
		cams.camera = &model.Camera{UUID: uuid.New(), SN: "CAM-001"}
	}

	if users.target == nil && users.targetErr == nil {
		users.target = &model.UserToken{UserUUID: "user-2", UserName: "budi", RoleName: "operator"}
	}

	return &usecase{repoCamera: cams, repoUser: users}
}

func TestRevokeCameraAccessErrors(t *testing.T) {
	tests := []struct {
		name        string
		cams        *fakeAssignCameraRepo
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "camera not found",
			cams:        &fakeAssignCameraRepo{cameraErr: gorm.ErrRecordNotFound},
			wantStatus:  http.StatusNotFound,
			wantMessage: "Camera not found",
		},
		{
			name:        "no grant",
			cams:        &fakeAssignCameraRepo{accessErr: gorm.ErrRecordNotFound},
			wantStatus:  http.StatusNotFound,
			wantMessage: "Camera access not found",
		},
		{
			name:        "already revoked",
			cams:        &fakeAssignCameraRepo{access: &model.CameraAccess{ID: 9, Active: false}},
			wantStatus:  http.StatusConflict,
			wantMessage: "Camera access already revoked",
		},
		{
			name:        "database error",
			cams:        &fakeAssignCameraRepo{accessErr: errors.New("connection refused")},
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &fakeAssignUserRepo{}
			u := newRevoke(tt.cams, users)

			assertAPIStatus(t, revoke(u, uuid.NewString(), "user-2"), tt.wantStatus, tt.wantMessage)

			if len(tt.cams.deactivated) != 0 || len(users.activities) != 0 {
				t.Fatal("a rejected revoke must not write anything")
			}
		})
	}
}

func TestRevokeCameraAccessDeactivatesGrant(t *testing.T) {
	cams := &fakeAssignCameraRepo{access: &model.CameraAccess{ID: 42, Active: true}}
	users := &fakeAssignUserRepo{}
	u := newRevoke(cams, users)

	if err := revoke(u, uuid.NewString(), "user-2"); err != nil {
		t.Fatalf("RevokeCameraAccess: %v", err)
	}

	if len(cams.deactivated) != 1 || cams.deactivated[0] != 42 || cams.created != 0 {
		t.Fatalf("deactivated=%v created=%d, want only grant 42 deactivated", cams.deactivated, cams.created)
	}

	if len(users.activities) != 1 {
		t.Fatalf("activities = %d, want 1", len(users.activities))
	}

	a := users.activities[0]
	if a.Activity != "REVOKED_CAMERA" {
		t.Errorf("activity = %q, want REVOKED_CAMERA", a.Activity)
	}

	if a.TargetUserID == nil || *a.TargetUserID != "user-2" {
		t.Errorf("target_user_id = %v, want user-2", a.TargetUserID)
	}

	for _, want := range []string{"Revoked", "CAM-001", "budi"} {
		if a.Message == nil || !strings.Contains(*a.Message, want) {
			t.Errorf("message = %v, want it to contain %q", a.Message, want)
		}
	}
}

func TestRevokeCameraAccessForDeletedUser(t *testing.T) {
	cams := &fakeAssignCameraRepo{access: &model.CameraAccess{ID: 42, Active: true}}
	users := &fakeAssignUserRepo{targetErr: gorm.ErrRecordNotFound}
	u := newRevoke(cams, users)

	if err := revoke(u, uuid.NewString(), "user-2"); err != nil {
		t.Fatalf("RevokeCameraAccess: %v", err)
	}

	if msg := users.activities[0].Message; msg == nil || !strings.HasSuffix(*msg, "for user user-2") {
		t.Errorf("message = %v, want it to fall back to the user UUID", msg)
	}
}
