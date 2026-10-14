package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	Secret []byte
	Issuer string
}

type CustomClaims struct {
	User    string `json:"sub"`
	Session string `json:"sid"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	jwt.RegisteredClaims
}

type Claims struct {
	UserUUID string
	UserName string
	RoleUUID string
	RoleName string
}

func (t Token) GenerateToken(claim *Claims, exp time.Duration, session string) (string, error) {
	claims := CustomClaims{
		User:    claim.UserUUID,
		Session: session,
		Type:    "access",
		Role:    claim.RoleName,
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

func (t Token) VerifyToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
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

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.Issuer != t.Issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	if claims.Type != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}
