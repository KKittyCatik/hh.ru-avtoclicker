package bot_detector

import (
	"testing"
	"time"

	"hh-autoresponder/internal/hh"
)

func TestIsBotByOptions(t *testing.T) {
	msg := hh.Message{Type: "question", Options: []hh.MessageOption{{ID: "1", Text: "Да"}}}
	if !IsBot(msg) {
		t.Fatal("expected bot message")
	}
}

func TestIsBotByFastResponse(t *testing.T) {
	n := time.Now()
	msg := hh.Message{CreatedAt: n, NegotiationCreatedAt: n.Add(-2 * time.Second), Text: "Здравствуйте"}
	if !IsBot(msg) {
		t.Fatal("expected bot by timing")
	}
}
