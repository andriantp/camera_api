package password_test

import (
	"golang.org/x/crypto/bcrypt"
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

func TestCreatePasswordUsesCost(t *testing.T) {
	hash, err := password.CreatePassword("p@ssw0rd-example")
	if err != nil {
		t.Fatalf("CreatePassword: %v", err)
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("bcrypt.Cost: %v", err)
	}

	if cost != password.Cost {
		t.Fatalf("cost = %d, want %d", cost, password.Cost)
	}

	if password.NeedsRehash(hash) {
		t.Fatal("a hash with the current cost must not need a rehash")
	}
}

func TestNeedsRehash(t *testing.T) {
	// MinCost keeps the test fast; any cost other than password.Cost counts.
	old, err := bcrypt.GenerateFromPassword([]byte("p@ssw0rd-example"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}

	if !password.NeedsRehash(string(old)) {
		t.Error("a hash with another cost must need a rehash")
	}

	if password.NeedsRehash("not-a-bcrypt-hash") {
		t.Error("an unparseable hash must not be reported for rehash")
	}

	// Old hashes keep verifying until they are upgraded.
	if ok, err := password.CheckPassword("p@ssw0rd-example", string(old)); !ok || err != nil {
		t.Errorf("old hash must still verify: ok=%v err=%v", ok, err)
	}
}
