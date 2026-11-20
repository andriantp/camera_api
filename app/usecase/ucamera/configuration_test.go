package ucamera

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"camera_api/pkg/adapter/database/model"
	repocamera "camera_api/pkg/adapter/database/repository/camera"
	"camera_api/pkg/adapter/database/repository/user"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"

	"github.com/andriantp/camera"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type txKey struct{}

// fakeTxUserRepo marks the context passed into a transaction so tests can
// check that queries run inside it.
type fakeTxUserRepo struct {
	user.Repository
}

func (f *fakeTxUserRepo) DoTransaction(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	return fn(context.WithValue(ctx, txKey{}, true), nil)
}

func (f *fakeTxUserRepo) UpdateSessionLastSeen(context.Context, string, time.Time, string, string) error {
	return nil
}

func (f *fakeTxUserRepo) CreateActivity(context.Context, *model.UserActivity) error {
	return nil
}

// fakeConfigRepo implements the camera repository calls used by the
// configuration usecases. Other methods panic on the nil embedded interface.
type fakeConfigRepo struct {
	repocamera.Repository

	camera    *model.Camera
	cameraErr error

	direction    *model.Direction
	directionErr error
	directionTx  bool

	updateErr error
}

func (f *fakeConfigRepo) FindCameraByUUID(context.Context, string) (*model.Camera, error) {
	return f.camera, f.cameraErr
}

func (f *fakeConfigRepo) GetDirectionByUUID(ctx context.Context, _ string) (*model.Direction, error) {
	f.directionTx = ctx.Value(txKey{}) != nil
	return f.direction, f.directionErr
}

func (f *fakeConfigRepo) UpdateCameraSN(context.Context, string, string) error { return f.updateErr }

func (f *fakeConfigRepo) UpdateCameraStatus(context.Context, string, bool) error { return f.updateErr }

func (f *fakeConfigRepo) UpdateCameraCredential(context.Context, string, camera.Credential) error {
	return f.updateErr
}

func (f *fakeConfigRepo) UpdateCameraDirection(context.Context, string, string) error {
	return f.updateErr
}

var superAdmin = request.UserActor{UserID: "user-1", SessionID: "session-1", Role: "super_admin"}

// configCalls runs each configuration usecase that updates a camera.
var configCalls = map[string]func(*usecase) error{
	"UpdateCameraSN": func(u *usecase) error {
		return u.UpdateCameraSN(context.Background(), superAdmin, "cam", request.UpdateCameraSN{SN: "abc"})
	},
	"UpdateCameraStatus": func(u *usecase) error {
		active := true
		return u.UpdateCameraStatus(context.Background(), superAdmin, "cam", request.UpdateCameraStatus{Active: &active})
	},
	"UpdateCameraCredential": func(u *usecase) error {
		host := "10.0.0.1"
		return u.UpdateCameraCredential(context.Background(), superAdmin, "cam", request.UpdateCameraCredential{Host: &host})
	},
	"UpdateCameraDirection": func(u *usecase) error {
		return u.UpdateCameraDirection(context.Background(), superAdmin, "cam", request.UpdateCameraDirection{Direction: uuid.NewString()})
	},
}

func assertAPIStatus(t *testing.T, err error, want int, wantMessage string) {
	t.Helper()

	var apiErr response.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != want {
		t.Fatalf("err = %v, want APIError with status %d", err, want)
	}

	if wantMessage != "" && apiErr.Message != wantMessage {
		t.Fatalf("message = %q, want %q", apiErr.Message, wantMessage)
	}
}

func TestConfigurationCameraNotFound(t *testing.T) {
	for name, call := range configCalls {
		t.Run(name, func(t *testing.T) {
			u := &usecase{
				repoUser:   &fakeTxUserRepo{},
				repoCamera: &fakeConfigRepo{cameraErr: gorm.ErrRecordNotFound},
			}

			assertAPIStatus(t, call(u), http.StatusNotFound, "Camera not found")
		})
	}
}

func TestConfigurationDatabaseError(t *testing.T) {
	for name, call := range configCalls {
		t.Run(name, func(t *testing.T) {
			u := &usecase{
				repoUser: &fakeTxUserRepo{},
				repoCamera: &fakeConfigRepo{
					camera:    &model.Camera{UUID: uuid.New()},
					direction: &model.Direction{UUID: uuid.New(), Name: "D1"},
					updateErr: errors.New("connection refused"),
				},
				camera: &fakeEncrypter{},
			}

			assertAPIStatus(t, call(u), http.StatusInternalServerError, "Internal server error")
		})
	}
}

func TestUpdateCameraDirectionNotFound(t *testing.T) {
	repo := &fakeConfigRepo{
		camera:       &model.Camera{UUID: uuid.New()},
		directionErr: gorm.ErrRecordNotFound,
	}
	u := &usecase{repoUser: &fakeTxUserRepo{}, repoCamera: repo}

	err := configCalls["UpdateCameraDirection"](u)
	assertAPIStatus(t, err, http.StatusNotFound, "Direction not found")

	if !repo.directionTx {
		t.Fatal("direction must be looked up inside the transaction")
	}
}

func TestUpdateCameraDirectionSuccess(t *testing.T) {
	repo := &fakeConfigRepo{
		camera:    &model.Camera{UUID: uuid.New()},
		direction: &model.Direction{UUID: uuid.New(), Name: "D1"},
	}
	u := &usecase{repoUser: &fakeTxUserRepo{}, repoCamera: repo}

	if err := configCalls["UpdateCameraDirection"](u); err != nil {
		t.Fatalf("UpdateCameraDirection: %v", err)
	}

	if !repo.directionTx {
		t.Fatal("direction must be looked up inside the transaction")
	}
}

// fakeEncrypter stands in for the camera adapter's Encrypt.
type fakeEncrypter struct {
	fakeCameraAdapter
}

func (f *fakeEncrypter) Encrypt(plain string) (string, error) {
	return "enc:" + plain, nil
}
