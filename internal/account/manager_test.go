package account

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts.json")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	mgr := NewManager(path)
	if err := mgr.Load(); err != nil {
		t.Fatalf("load empty accounts file: %v", err)
	}
	if len(mgr.GetAll()) != 0 {
		t.Fatalf("expected no accounts, got %d", len(mgr.GetAll()))
	}
}

func TestEncryptDecryptPassword(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	encrypted, err := EncryptPassword("secret", key)
	if err != nil {
		t.Fatalf("encrypt password: %v", err)
	}
	if encrypted == "secret" {
		t.Fatal("expected encrypted payload")
	}
	decrypted, err := DecryptPassword(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt password: %v", err)
	}
	if decrypted != "secret" {
		t.Fatalf("unexpected decrypted password: %s", decrypted)
	}
}

func TestLoadEncryptionKeyFromEnv(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	if err := os.Setenv("ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(key)); err != nil {
		t.Fatalf("set env: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("ENCRYPTION_KEY") })

	got, err := LoadEncryptionKeyFromEnv()
	if err != nil {
		t.Fatalf("load key: %v", err)
	}
	if string(got) != string(key) {
		t.Fatalf("unexpected key value")
	}
}

func TestUpdateMutatorCanAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts.json")
	if err := os.WriteFile(path, []byte("[]"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	mgr := NewManager(path)
	if err := mgr.Load(); err != nil {
		t.Fatalf("load accounts file: %v", err)
	}

	if err := mgr.Update(func(items *[]Account) error {
		*items = append(*items, Account{ID: "acc-1", Name: "Test"})
		return nil
	}); err != nil {
		t.Fatalf("update accounts: %v", err)
	}

	items := mgr.GetAll()
	if len(items) != 1 {
		t.Fatalf("expected one account, got %d", len(items))
	}
	if items[0].ID != "acc-1" {
		t.Fatalf("unexpected account id: %s", items[0].ID)
	}
}
