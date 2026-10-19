package ucamera

import (
	adaptercamera "camera_api/pkg/adapter/camera"
	repocamera "camera_api/pkg/adapter/database/repository/camera"
	"camera_api/pkg/adapter/database/repository/user"
	"camera_api/pkg/adapter/s3"
)

type usecase struct {
	repoUser   user.Repository
	repoCamera repocamera.Repository
	camera     adaptercamera.Camera
	storage    s3.Storage
}

type Usecase interface {
	Protocol
	Camera
	Configuration
	Device
	Command
	Direction
	Content
	Snapshot
}

func NewUsecase(repoUser user.Repository,
	repoCamera repocamera.Repository,
	camera adaptercamera.Camera,
	storage s3.Storage) Usecase {
	return &usecase{
		repoUser:   repoUser,
		repoCamera: repoCamera,
		camera:     camera,
		storage:    storage,
	}
}
