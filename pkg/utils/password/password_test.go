package password_test

import (
	"testing"

	"camera_api/pkg/utils/password"
)

func TestPassword(t *testing.T) {
	pass := "test123"

	hash, err := password.CreatePassword(pass)
	if err != nil {
		t.Fatalf("CreatePassword: %v", err)
	}

	ok, err := password.CheckPassword(pass, hash)
	if err != nil {
		t.Fatalf("CheckPassword: %v", err)
	}

	if !ok {
		t.Fatal("Password Miss Match")
	}
}

func TestPasswordMismatch(t *testing.T) {
	pass := "test123"
	wrongPass := "wrong123"

	hash, err := password.CreatePassword(pass)
	if err != nil {
		t.Fatalf("CreatePassword: %v", err)
	}

	ok, err := password.CheckPassword(wrongPass, hash)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if ok {
		t.Fatal("expected password mismatch")
	}
}
