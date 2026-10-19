package uuser

import (
	"camera_api/pkg/adapter/database/repository/user"
	"camera_api/pkg/utils/token"
	"time"
)

type (
	Expire struct {
		Token   time.Duration
		Refresh time.Duration
	}
)

type usecase struct {
	exp      Expire
	token    token.Token
	repoUser user.Repository
}

type Usecase interface {
	GetExpire() Expire
	Auth
	Profile
	Management
	Activity
}

func NewUsecase(exp Expire,
	jwt token.Token,
	repoUser user.Repository) Usecase {
	return &usecase{
		exp:      exp,
		token:    jwt,
		repoUser: repoUser,
	}
}
