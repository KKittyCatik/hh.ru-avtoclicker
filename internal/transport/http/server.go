package httptransport

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"hh-autoresponder/internal/transport/ws"
)

func NewServer(addr string, handlers *Handlers, hub *ws.Hub) *http.Server {
	handlers.Init()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
	    w.Header().Set("Content-Type", "application/json")
	    w.WriteHeader(http.StatusOK)
	    _, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/ws", hub.ServeHTTP)
	r.Get("/callback", handlers.AuthCallback)

	r.Route("/api", func(r chi.Router) {
		r.Get("/auth/login", handlers.AuthLogin)
		r.Post("/apply/start", handlers.StartApply)
		r.Post("/apply/stop", handlers.StopApply)
		r.Get("/stats", handlers.GetStats)
		r.Get("/negotiations", handlers.ListNegotiations)
		r.Get("/negotiations/{id}/messages", handlers.ListNegotiationMessages)
		r.Post("/negotiations/{id}/reply", handlers.ReplyNegotiation)
		r.Post("/negotiations/{id}/generate-reply", handlers.GenerateNegotiationReply)
		r.Post("/negotiations/reply/run", handlers.TriggerReplies)
		r.Get("/accounts", handlers.ListAccounts)
		r.Post("/resume/publish", handlers.PublishResume)
	})
	r.Handle("/*", spaHandler(frontendRoot()))

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

func frontendRoot() string {
	if _, err := os.Stat("web/dist/index.html"); err == nil {
		return "web/dist"
	}
	return "web"
}

func spaHandler(root string) http.Handler {
	fs := http.Dir(root)
	fileServer := http.FileServer(fs)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleaned := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		if cleaned != "/" {
			if file, err := fs.Open(strings.TrimPrefix(cleaned, "/")); err == nil {
				_ = file.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		clone := *r
		clonedURL := *r.URL
		clonedURL.Path = "/"
		clonedURL.RawPath = ""
		clonedURL.ForceQuery = false
		clonedURL.RawQuery = ""
		clone.URL = &url.URL{}
		*clone.URL = clonedURL
		fileServer.ServeHTTP(w, &clone)
	})
}
