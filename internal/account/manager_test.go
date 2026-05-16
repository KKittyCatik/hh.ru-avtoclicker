package account

import (
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
