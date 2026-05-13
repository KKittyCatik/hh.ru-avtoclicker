package httptransport

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"hh-autoresponder/internal/transport/ws"
)

func NewServer(addr string, handlers *Handlers, hub *ws.Hub) *http.Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/ws", hub.ServeHTTP)
	r.Handle("/", http.FileServer(http.Dir("web")))

	r.Route("/api", func(r chi.Router) {
		r.Post("/apply/start", handlers.StartApply)
		r.Post("/apply/stop", handlers.StopApply)
		r.Get("/stats", handlers.GetStats)
		r.Get("/negotiations", handlers.ListNegotiations)
		r.Post("/negotiations/{id}/reply", handlers.ReplyNegotiation)
		r.Post("/negotiations/reply/run", handlers.TriggerReplies)
		r.Get("/accounts", handlers.ListAccounts)
		r.Post("/resume/publish", handlers.PublishResume)
	})

	return &http.Server{
		Addr:              normalizeAddr(addr),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func normalizeAddr(addr string) string {
	trimmed := strings.TrimSpace(addr)
	if trimmed == "" {
		return ":8080"
	}
	if strings.Contains(trimmed, ":") {
		return trimmed
	}
	return fmt.Sprintf(":%s", trimmed)
}
