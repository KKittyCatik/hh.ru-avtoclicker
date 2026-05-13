package httptransport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"hh-autoresponder/internal/account"
	"hh-autoresponder/internal/hh"
	"hh-autoresponder/internal/monitor"
	"hh-autoresponder/internal/worker"
)

type Handlers struct {
	ApplyWorker *worker.ApplyWorker
	ReplyWorker *worker.ReplyWorker
	Stats       *monitor.Collector
	Accounts    *account.Manager
	HHClient    *hh.Client
	SearchURLs  []string
	ResumeID    string
}

func (h *Handlers) StartApply(w http.ResponseWriter, r *http.Request) {
	if err := h.ApplyWorker.Start(r.Context(), h.SearchURLs, h.ResumeID); err != nil {
		h.writeError(w, http.StatusConflict, fmt.Errorf("start apply worker: %w", err))
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (h *Handlers) StopApply(w http.ResponseWriter, r *http.Request) {
	h.ApplyWorker.Stop()
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.Stats.Snapshot())
}

func (h *Handlers) ListNegotiations(w http.ResponseWriter, r *http.Request) {
	items, err := h.HHClient.ListNegotiations(r.Context())
	if err != nil {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("list negotiations: %w", err))
		return
	}
	h.writeJSON(w, http.StatusOK, items)
}

func (h *Handlers) ListNegotiationMessages(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, fmt.Errorf("missing negotiation id"))
		return
	}
	items, err := h.HHClient.ListNegotiationMessages(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("list negotiation messages: %w", err))
		return
	}
	h.writeJSON(w, http.StatusOK, items)
}

func (h *Handlers) ReplyNegotiation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, fmt.Errorf("missing negotiation id"))
		return
	}
	var req hh.SendReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Errorf("decode reply request: %w", err))
		return
	}
	if err := h.HHClient.SendNegotiationReply(r.Context(), id, req); err != nil {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("send negotiation reply: %w", err))
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (h *Handlers) GenerateNegotiationReply(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, fmt.Errorf("missing negotiation id"))
		return
	}
	reply, err := h.ReplyWorker.GenerateDraft(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("generate negotiation reply: %w", err))
		return
	}
	h.writeJSON(w, http.StatusOK, reply)
}

func (h *Handlers) ListAccounts(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.Accounts.GetAll())
}

func (h *Handlers) PublishResume(w http.ResponseWriter, r *http.Request) {
	if h.ResumeID == "" {
		h.writeError(w, http.StatusBadRequest, fmt.Errorf("resume id is not configured"))
		return
	}
	if err := h.HHClient.PublishResume(r.Context(), h.ResumeID); err != nil {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("publish resume: %w", err))
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
}

func (h *Handlers) TriggerReplies(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Resume             string `json:"resume"`
		VacancyDescription string `json:"vacancy_description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Errorf("decode trigger replies request: %w", err))
		return
	}
	if err := h.ReplyWorker.ProcessNegotiations(r.Context(), req.Resume, req.VacancyDescription); err != nil {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("process replies: %w", err))
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

func (h *Handlers) writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handlers) writeError(w http.ResponseWriter, code int, err error) {
	h.writeJSON(w, code, map[string]string{"error": err.Error()})
}
