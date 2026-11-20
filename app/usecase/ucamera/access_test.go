package ucamera

import (
	"context"
	"testing"

	"camera_api/pkg/adapter/database/model"
	repocamera "camera_api/pkg/adapter/database/repository/camera"
	"camera_api/pkg/api/request"

	"github.com/andriantp/camera"
	"github.com/google/uuid"
)

// fakeAccessRepo records which camera query a usecase used.
type fakeAccessRepo struct {
	repocamera.Repository

	calls  []string
	grants int
}

func (f *fakeAccessRepo) FindCameraByUUID(context.Context, string) (*model.Camera, error) {
	f.calls = append(f.calls, "all")
	return &model.Camera{UUID: uuid.New()}, nil
}

func (f *fakeAccessRepo) FindCameraAccessUser(context.Context, string, string) (*model.Camera, error) {
	f.calls = append(f.calls, "granted")
	return &model.Camera{UUID: uuid.New()}, nil
}

func (f *fakeAccessRepo) FindCameras(context.Context) ([]model.Camera, error) {
	f.calls = append(f.calls, "all")
	return nil, nil
}

func (f *fakeAccessRepo) FindCamerasByUser(context.Context, string) ([]model.Camera, error) {
	f.calls = append(f.calls, "granted")
	return nil, nil
}

func (f *fakeAccessRepo) FindCamerasByProtocol(context.Context, string) ([]model.CameraSummary, error) {
	f.calls = append(f.calls, "all")
	return nil, nil
}

func (f *fakeAccessRepo) FindCamerasByProtocolByUser(context.Context, string, string) ([]model.CameraSummary, error) {
	f.calls = append(f.calls, "granted")
	return nil, nil
}

func (f *fakeAccessRepo) FindProtocolByUUID(context.Context, string) (*model.Protocol, error) {
	return &model.Protocol{UUID: uuid.New(), Name: model.ProtocolISAPI}, nil
}

func (f *fakeAccessRepo) CreateCamera(_ context.Context, c *model.Camera) error {
	c.UUID = uuid.New()
	return nil
}

func (f *fakeAccessRepo) CreateCameraAccess(context.Context, *model.CameraAccess) error {
	f.grants++
	return nil
}

// fakeCredentialAdapter encrypts credentials for CreateCamera.
type fakeCredentialAdapter struct {
	fakeCameraAdapter
}

func (f *fakeCredentialAdapter) EncryptCredential(c camera.Credential) (*camera.Credential, error) {
	return &c, nil
}

var roleAccess = []struct {
	role string
	want string // "all" or "granted"
}{
	{role: "super_admin", want: "all"},
	{role: "admin", want: "all"},
	{role: "operator", want: "granted"},
	{role: "user", want: "granted"},
}

func TestCameraVisibilityByRole(t *testing.T) {
	usecases := map[string]func(*usecase, request.UserActor) error{
		"GetCamera": func(u *usecase, a request.UserActor) error {
			_, err := u.GetCamera(context.Background(), a, uuid.NewString())
			return err
		},
		"GetCameras": func(u *usecase, a request.UserActor) error {
			_, err := u.GetCameras(context.Background(), a)
			return err
		},
		"GetCamerasByProtocol": func(u *usecase, a request.UserActor) error {
			_, err := u.GetCamerasByProtocol(context.Background(), a, uuid.NewString())
			return err
		},
	}

	for name, call := range usecases {
		for _, tc := range roleAccess {
			t.Run(name+"/"+tc.role, func(t *testing.T) {
				repo := &fakeAccessRepo{}
				u := &usecase{repoCamera: repo}

				if err := call(u, request.UserActor{UserID: "user-1", Role: tc.role}); err != nil {
					t.Fatalf("%s: %v", name, err)
				}

				if len(repo.calls) != 1 || repo.calls[0] != tc.want {
					t.Fatalf("queries = %v, want [%s]", repo.calls, tc.want)
				}
			})
		}
	}
}

func TestCreateCameraGrantsAccessOnlyToRestrictedRoles(t *testing.T) {
	for _, tc := range roleAccess {
		t.Run(tc.role, func(t *testing.T) {
			repo := &fakeAccessRepo{}
			u := &usecase{
				repoUser:   &fakeTxUserRepo{},
				repoCamera: repo,
				camera:     &fakeCredentialAdapter{},
			}

			_, err := u.CreateCamera(
				context.Background(),
				request.UserActor{UserID: "user-1", SessionID: "session-1", Role: tc.role},
				request.CreateCamera{SN: "sn1", Protocol: uuid.NewString(), Host: "10.0.0.1", Username: "a", Password: "b"},
			)
			if err != nil {
				t.Fatalf("CreateCamera: %v", err)
			}

			wantGrants := 0
			if tc.want == "granted" {
				wantGrants = 1
			}

			if repo.grants != wantGrants {
				t.Fatalf("access grants = %d, want %d", repo.grants, wantGrants)
			}
		})
	}
}
