package hh

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoJSONFastSkipsRandomDelay(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(srv.Client(), nil, slog.Default())
	client.baseURL = srv.URL

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.DoJSON(canceled, http.MethodGet, "/me", nil, nil); err == nil || !strings.Contains(err.Error(), "wait random delay") {
		t.Fatalf("expected random delay error, got: %v", err)
	}

	if err := client.DoJSONFast(context.Background(), http.MethodGet, "/me", nil, nil); err != nil {
		t.Fatalf("expected DoJSONFast to avoid random delay, got: %v", err)
	}
}
