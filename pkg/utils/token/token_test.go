package token_test

import (
	"testing"
	"time"

	"camera_api/pkg/utils/token"
)

func TestGenerateAndVerifyToken(t *testing.T) {
	jwt := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "camera",
	}

	user := &token.Claims{
		UserUUID: "user-123",
		RoleName: "admin",
	}

	sessionID := "session-123"

	tokenString, err := jwt.GenerateToken(
		user,
		time.Hour,
		sessionID,
	)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	if tokenString == "" {
		t.Fatal("token is empty")
	}

	claims, err := jwt.VerifyToken(tokenString)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
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

	if claims.Role != user.RoleName {
		t.Errorf(
			"unexpected role: got %q, want %q",
			claims.Role,
			user.RoleName,
		)
	}

	if claims.Type != "access" {
		t.Errorf(
			"unexpected token type: got %q, want %q",
			claims.Type,
			"access",
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

func TestVerifyTokenWrongSecret(t *testing.T) {
	jwt := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "camera",
	}

	user := &token.Claims{
		UserUUID: "user-123",
		RoleName: "admin",
	}

	tokenString, err := jwt.GenerateToken(
		user,
		time.Hour,
		"session-123",
	)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	wrongJWT := token.Token{
		Secret: []byte("wrong-secret"),
		Issuer: "camera",
	}

	_, err = wrongJWT.VerifyToken(tokenString)
	if err == nil {
		t.Fatal("expected error for invalid secret")
	}
}

func TestVerifyTokenWrongIssuer(t *testing.T) {
	jwt := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "camera",
	}

	user := &token.Claims{
		UserUUID: "user-123",
		RoleName: "admin",
	}

	tokenString, err := jwt.GenerateToken(
		user,
		time.Hour,
		"session-123",
	)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	wrongIssuer := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "another-product",
	}

	_, err = wrongIssuer.VerifyToken(tokenString)
	if err == nil {
		t.Fatal("expected error for invalid issuer")
	}
}
