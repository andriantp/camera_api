package token_test

import (
	"camera_api/pkg/utils/token"
	"testing"
	"time"
)

func TestGenerateAndVerifyRefreshToken(t *testing.T) {
	jwt := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "camera",
	}

	user := &token.Claims{
		UserUUID: "user-123",
		RoleName: "admin",
	}

	sessionID := "session-123"

	tokenString, err := jwt.GenerateRefresh(
		user,
		time.Hour*24,
		sessionID,
	)
	if err != nil {
		t.Fatalf("GenerateRefresh: %v", err)
	}

	if tokenString == "" {
		t.Fatal("refresh token is empty")
	}

	claims, err := jwt.VerifyRefresh(tokenString)
	if err != nil {
		t.Fatalf("VerifyRefresh: %v", err)
	}

	if claims.User != user.UserUUID {
		t.Errorf(
			"unexpected user: got %q, want %q",
			claims.User,
			user.UserUUID,
		)
	}

	if claims.Session != sessionID {
		t.Errorf(
			"unexpected session: got %q, want %q",
			claims.Session,
			sessionID,
		)
	}

	if claims.Type != "refresh" {
		t.Errorf(
			"unexpected token type: got %q, want %q",
			claims.Type,
			"refresh",
		)
	}

	if claims.Issuer != jwt.Issuer {
		t.Errorf(
			"unexpected issuer: got %q, want %q",
			claims.Issuer,
			jwt.Issuer,
		)
	}
}

func TestAccessTokenCannotBeVerifiedAsRefresh(t *testing.T) {
	jwt := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "camera",
	}

	user := &token.Claims{
		UserUUID: "user-123",
		RoleName: "admin",
	}

	accessToken, err := jwt.GenerateToken(
		user,
		time.Hour,
		"session-123",
	)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	_, err = jwt.VerifyRefresh(accessToken)
	if err == nil {
		t.Fatal("expected access token to be rejected as refresh token")
	}
}

func TestRefreshTokenCannotBeVerifiedAsAccess(t *testing.T) {
	jwt := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "camera",
	}

	user := &token.Claims{
		UserUUID: "user-123",
		RoleName: "admin",
	}

	refreshToken, err := jwt.GenerateRefresh(
		user,
		time.Hour*24,
		"session-123",
	)
	if err != nil {
		t.Fatalf("GenerateRefresh: %v", err)
	}

	_, err = jwt.VerifyToken(refreshToken)
	if err == nil {
		t.Fatal("expected refresh token to be rejected as access token")
	}
}
