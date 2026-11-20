package ucamera

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"camera_api/pkg/adapter/database/model"
	repocamera "camera_api/pkg/adapter/database/repository/camera"
	"camera_api/pkg/adapter/database/repository/user"
	"camera_api/pkg/api/request"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fakeAssignCameraRepo struct {
	repocamera.Repository

	camera    *model.Camera
	cameraErr error

	access    *model.CameraAccess
	accessErr error
	createErr error

	created     int
	activated   []int64
	deactivated []int64
}

func (f *fakeAssignCameraRepo) DoTransaction(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	return fn(ctx, nil)
}

func (f *fakeAssignCameraRepo) FindCameraByUUID(context.Context, string) (*model.Camera, error) {
	return f.camera, f.cameraErr
}

func (f *fakeAssignCameraRepo) FindCameraAccess(context.Context, string, string) (*model.CameraAccess, error) {
	return f.access, f.accessErr
}

func (f *fakeAssignCameraRepo) CreateCameraAccess(context.Context, *model.CameraAccess) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created++
	return nil
}

func (f *fakeAssignCameraRepo) ActivateCameraAccess(_ context.Context, id int64) error {
	f.activated = append(f.activated, id)
	return nil
}

func (f *fakeAssignCameraRepo) DeactivateCameraAccess(_ context.Context, id int64) error {
	f.deactivated = append(f.deactivated, id)
	return nil
}

type fakeAssignUserRepo struct {
	user.Repository

	target    *model.UserToken
	targetErr error

	activities []model.UserActivity
}

func (f *fakeAssignUserRepo) FindUserByUserUUID(context.Context, string) (*model.UserToken, error) {
	return f.target, f.targetErr
}

func (f *fakeAssignUserRepo) UpdateSessionLastSeen(context.Context, string, time.Time, string, string) error {
	return nil
}

func (f *fakeAssignUserRepo) CreateActivity(_ context.Context, a *model.UserActivity) error {
	f.activities = append(f.activities, *a)
	return nil
}

func newAssign(cams *fakeAssignCameraRepo, users *fakeAssignUserRepo) (*usecase, request.AssignCameraRequest) {
	if cams.camera == nil && cams.cameraErr == nil {
		cams.camera = &model.Camera{UUID: uuid.New(), SN: "CAM-001"}
	}

	if users.target == nil && users.targetErr == nil {
		users.target = &model.UserToken{UserUUID: "user-2", UserName: "budi", RoleName: "operator"}
	}

	if cams.access == nil && cams.accessErr == nil {
		cams.accessErr = gorm.ErrRecordNotFound
	}

	return &usecase{repoCamera: cams, repoUser: users}, request.AssignCameraRequest{
		CameraUUID: uuid.NewString(),
		UserUUID:   "user-2",
	}
}

func assign(u *usecase, req request.AssignCameraRequest) error {
	_, err := assignChanged(u, req)
	return err
}

func assignChanged(u *usecase, req request.AssignCameraRequest) (bool, error) {
	return u.AssignCamera(
		context.Background(),
		request.UserActor{UserID: "admin-1", SessionID: "session-1", Role: "admin"},
		req.CameraUUID,
		req.UserUUID,
	)
}

func TestAssignCameraValidation(t *testing.T) {
	tests := []struct {
		name        string
		cams        *fakeAssignCameraRepo
		users       *fakeAssignUserRepo
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "camera not found",
			cams:        &fakeAssignCameraRepo{cameraErr: gorm.ErrRecordNotFound},
			users:       &fakeAssignUserRepo{},
			wantStatus:  http.StatusNotFound,
			wantMessage: "Camera not found",
		},
		{
			name:        "user not found",
			cams:        &fakeAssignCameraRepo{},
			users:       &fakeAssignUserRepo{targetErr: gorm.ErrRecordNotFound},
			wantStatus:  http.StatusNotFound,
			wantMessage: "User not found",
		},
		{
			name:        "target is admin",
			cams:        &fakeAssignCameraRepo{},
			users:       &fakeAssignUserRepo{target: &model.UserToken{UserUUID: "user-2", RoleName: "admin"}},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Camera access can only be assigned to operator or user",
		},
		{
			name:        "target is super_admin",
			cams:        &fakeAssignCameraRepo{},
			users:       &fakeAssignUserRepo{target: &model.UserToken{UserUUID: "user-2", RoleName: "super_admin"}},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Camera access can only be assigned to operator or user",
		},
		{
			name:        "database error",
			cams:        &fakeAssignCameraRepo{cameraErr: errors.New("connection refused")},
			users:       &fakeAssignUserRepo{},
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, req := newAssign(tt.cams, tt.users)

			assertAPIStatus(t, assign(u, req), tt.wantStatus, tt.wantMessage)

			if tt.cams.created != 0 || len(tt.cams.activated) != 0 || len(tt.users.activities) != 0 {
				t.Fatal("a rejected assignment must not write anything")
			}
		})
	}
}

func TestAssignCameraCreatesGrantWithAudit(t *testing.T) {
	cams := &fakeAssignCameraRepo{}
	users := &fakeAssignUserRepo{}
	u, req := newAssign(cams, users)

	changed, err := assignChanged(u, req)
	if err != nil || !changed {
		t.Fatalf("AssignCamera: changed=%v err=%v, want true and nil", changed, err)
	}

	if cams.created != 1 {
		t.Fatalf("grants created = %d, want 1", cams.created)
	}

	if len(users.activities) != 1 {
		t.Fatalf("activities = %d, want 1", len(users.activities))
	}

	a := users.activities[0]
	if a.TargetUserID == nil || *a.TargetUserID != req.UserUUID {
		t.Errorf("target_user_id = %v, want %s", a.TargetUserID, req.UserUUID)
	}

	for _, want := range []string{"Assigned", "CAM-001", "budi"} {
		if a.Message == nil || !strings.Contains(*a.Message, want) {
			t.Errorf("message = %v, want it to contain %q", a.Message, want)
		}
	}
}

func TestAssignCameraReactivatesRevokedGrant(t *testing.T) {
	cams := &fakeAssignCameraRepo{access: &model.CameraAccess{ID: 42, Active: false}}
	users := &fakeAssignUserRepo{}
	u, req := newAssign(cams, users)

	if err := assign(u, req); err != nil {
		t.Fatalf("AssignCamera: %v", err)
	}

	if cams.created != 0 || len(cams.activated) != 1 || cams.activated[0] != 42 {
		t.Fatalf("created=%d activated=%v, want reactivation of grant 42 only", cams.created, cams.activated)
	}

	if msg := users.activities[0].Message; msg == nil || !strings.HasPrefix(*msg, "Reactivated") {
		t.Errorf("message = %v, want it to start with Reactivated", msg)
	}
}

func TestAssignCameraIsIdempotent(t *testing.T) {
	tests := []struct {
		name string
		cams *fakeAssignCameraRepo
	}{
		{name: "already active", cams: &fakeAssignCameraRepo{access: &model.CameraAccess{ID: 7, Active: true}}},
		{name: "concurrent duplicate insert", cams: &fakeAssignCameraRepo{createErr: gorm.ErrDuplicatedKey}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &fakeAssignUserRepo{}
			u, req := newAssign(tt.cams, users)

			changed, err := assignChanged(u, req)
			if err != nil {
				t.Fatalf("AssignCamera: %v", err)
			}

			if changed {
				t.Error("changed = true, want false for an access that already exists")
			}

			if len(tt.cams.activated) != 0 || len(users.activities) != 0 {
				t.Fatal("a no-op grant must not write anything")
			}
		})
	}
}
