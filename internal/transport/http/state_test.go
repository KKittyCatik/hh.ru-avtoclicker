package httptransport

import (
	"testing"
	"time"
)

func TestStateStoreValidateOneTime(t *testing.T) {
	store := newStateStore()
	store.Set("state-1")

	if !store.Validate("state-1") {
		t.Fatal("expected state to be valid")
	}
	if store.Validate("state-1") {
		t.Fatal("expected state to be one-time use")
	}
}

func TestStateStoreValidateExpired(t *testing.T) {
	store := newStateStore()
	store.states["expired"] = time.Now().Add(-time.Minute)

	if store.Validate("expired") {
		t.Fatal("expected expired state to be invalid")
	}
}
