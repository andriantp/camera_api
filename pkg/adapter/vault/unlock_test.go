package vault_test

import (
	"camera_api/pkg/adapter/vault"
	"testing"
)

func TestUnlock(t *testing.T) {
	expected := "http://localhost:8200"

	encode, err := vault.Lock(expected)
	if err != nil {
		t.Fatalf("lock failed: %v", err)
	}
	t.Logf("lock:%s", encode)

	decode, err := vault.Unlock(encode)
	if err != nil {
		t.Fatalf("unlock failed: %v", err)
	}
	t.Logf("unlock:%s", decode)

	if string(decode) != expected {
		t.Fatalf(
			"unexpected result: got %q, want %q",
			string(decode),
			expected,
		)
	}
}
