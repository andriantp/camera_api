package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomRefresh struct {
	User    string `json:"sub"`
	Session string `json:"sid"`
	Type    string `json:"type"`
	jwt.RegisteredClaims
}

type Refresh struct {
	Refresh string `json:"refresh"`
}

func (t Token) GenerateRefresh(claim *Claims, exp time.Duration, session string) (string, error) {
	claims := CustomRefresh{
		User:    claim.UserUUID,
		Session: session,
		Type:    "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return jwtToken.SignedString(t.Secret)
}

func (t Token) VerifyRefresh(tokenString string) (*CustomRefresh, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomRefresh{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf(
					"unexpected signing method: %v",
					token.Header["alg"],
				)
			}

			return t.Secret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomRefresh)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token claims")
	}

	if claims.Issuer != t.Issuer {
		return nil, fmt.Errorf("invalid refresh token issuer")
	}

	if claims.Type != "refresh" {
		return nil, fmt.Errorf("invalid refresh token type")
	}

	return claims, nil
}
