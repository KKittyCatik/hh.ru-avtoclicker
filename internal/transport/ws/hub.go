package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type Event struct {
	Event   string `json:"event"`
	Payload any    `json:"payload"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: map[*websocket.Conn]struct{}{}}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()

	ctx := r.Context()
	for {
		if _, _, err := conn.Read(ctx); err != nil {
			break
		}
	}
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	_ = conn.Close(websocket.StatusNormalClosure, "connection closed by server")
}

func (h *Hub) Broadcast(ctx context.Context, event string, payload any) error {
	b, err := json.Marshal(Event{Event: event, Payload: payload})
	if err != nil {
		return fmt.Errorf("marshal ws event: %w", err)
	}

	h.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := c.Write(writeCtx, websocket.MessageText, b)
		cancel()
		if err != nil {
			h.mu.Lock()
			delete(h.clients, c)
			h.mu.Unlock()
			_ = c.Close(websocket.StatusInternalError, "broadcast failed")
		}
	}
	return nil
}
