package middleware

import "camera_api/pkg/utils/token"

type Middleware struct {
	token token.Token
}

func NewMiddleware(jwt token.Token) *Middleware {
	return &Middleware{
		token: jwt,
	}
}
