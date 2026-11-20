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
	"camera_api/pkg/adapter/s3"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

// Fakes implement only what the tests use. Any other method call panics on
// the nil embedded interface, which flags unexpected access.

type fakeCameraRepo struct {
	repocamera.Repository
	camera *model.Camera
	calls  int
}

func (f *fakeCameraRepo) FindCameraByUUID(_ context.Context, _ string) (*model.Camera, error) {
	f.calls++
	return f.camera, nil
}

type fakeUserRepo struct {
	user.Repository
}

func (f *fakeUserRepo) DoTransaction(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	return fn(ctx, nil)
}

func (f *fakeUserRepo) UpdateSessionLastSeen(context.Context, string, time.Time, string, string) error {
	return nil
}

func (f *fakeUserRepo) CreateActivity(context.Context, *model.UserActivity) error {
	return nil
}

type fakeStorage struct {
	s3.Storage
	objects []string
}

func (f *fakeStorage) FindFile(_ context.Context, objectName string) (*minio.ObjectInfo, error) {
	f.objects = append(f.objects, objectName)
	return &minio.ObjectInfo{Size: 1024, ContentType: "video/mp4"}, nil
}

func (f *fakeStorage) GetURLDownload(_ context.Context, objectName string) (*string, *time.Time, error) {
	url := "https://s3.example.com/" + objectName
	expires := time.Now().Add(time.Minute)
	return &url, &expires, nil
}

func TestDownloadContentAtS3RejectsUnsafeFilename(t *testing.T) {
	filenames := []string{
		"../11111111-1111-1111-1111-111111111111/20260101T000000Z_20260101T001000Z.mp4",
		"20260101T000000Z_20260101T001000Z.mp4/../../secret.mp4",
		"..",
		"secret.mp4",
		"",
	}

	for _, filename := range filenames {
		t.Run(filename, func(t *testing.T) {
			cameras := &fakeCameraRepo{}
			storage := &fakeStorage{}
			u := &usecase{repoCamera: cameras, storage: storage}

			_, err := u.DownloadContentAtS3(
				context.Background(),
				request.UserActor{Role: "super_admin"},
				uuid.NewString(),
				filename,
			)

			var apiErr response.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
				t.Fatalf("err = %v, want 400 APIError", err)
			}

			if cameras.calls != 0 || len(storage.objects) != 0 {
				t.Fatal("unsafe filename must be rejected before touching DB or storage")
			}
		})
	}
}

func TestDownloadContentAtS3UsesCanonicalCameraPath(t *testing.T) {
	cameraUUID := uuid.New()
	filename := "20260101T000000Z_20260101T001000Z.mp4"

	storage := &fakeStorage{}
	u := &usecase{
		repoUser:   &fakeUserRepo{},
		repoCamera: &fakeCameraRepo{camera: &model.Camera{UUID: cameraUUID}},
		storage:    storage,
	}

	resp, err := u.DownloadContentAtS3(
		context.Background(),
		request.UserActor{Role: "super_admin"},
		cameraUUID.String(),
		filename,
	)
	if err != nil {
		t.Fatalf("DownloadContentAtS3: %v", err)
	}

	want := cameraUUID.String() + "/" + filename
	if len(storage.objects) != 1 || storage.objects[0] != want {
		t.Fatalf("object path = %v, want [%s]", storage.objects, want)
	}

	if resp.Filename != filename {
		t.Fatalf("filename = %q, want %q", resp.Filename, filename)
	}
}
