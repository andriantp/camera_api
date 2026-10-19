package usystem

type (
	Build struct {
		ENV     string `json:"environment"`
		Version string `json:"version"`
		Release string `json:"release"`
	}
)

type usecase struct {
	build Build
}

type Usecase interface {
	GetVersion() Build
}

func NewUsecase(build Build) Usecase {
	return &usecase{
		build: build,
	}
}

func (u *usecase) GetVersion() Build {
	return u.build
}
