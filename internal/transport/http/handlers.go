package httptransport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"hh-autoresponder/internal/account"
	"hh-autoresponder/internal/browser"
	"hh-autoresponder/internal/monitor"
	"hh-autoresponder/internal/worker"
)

type Handlers struct {
	Ctx         context.Context
	BrowserCtx  *browser.AccountContext
	ApplyWorker *worker.ApplyWorker
	ReplyWorker *worker.ReplyWorker
	Stats       *monitor.Collector
	Accounts    *account.Manager
	SearchURLs  []string
	ResumeID    string
	states      *stateStore
}

func (h *Handlers) Init() {
	if h.states == nil {
		h.states = newStateStore()
	}
}

func (h *Handlers) AuthLogin(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusGone, fmt.Errorf("oauth login is disabled, use browser account credentials"))
}

func (h *Handlers) AuthCallback(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusGone, fmt.Errorf("oauth callback is disabled"))
}

func (h *Handlers) StartApply(w http.ResponseWriter, r *http.Request) {
	if err := h.ApplyWorker.Start(h.Ctx, h.SearchURLs, h.ResumeID); err != nil {
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
	if h.BrowserCtx == nil {
		h.writeError(w, http.StatusServiceUnavailable, fmt.Errorf("browser context is not configured"))
		return
	}
	items, err := h.BrowserCtx.GetNegotiations(r.Context())
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
	if h.BrowserCtx == nil {
		h.writeError(w, http.StatusServiceUnavailable, fmt.Errorf("browser context is not configured"))
		return
	}
	items, err := h.BrowserCtx.GetMessages(r.Context(), id)
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
	var req struct {
		Text               string `json:"text,omitempty"`
		QuickReplyOptionID string `json:"quick_reply_option_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Errorf("decode reply request: %w", err))
		return
	}
	if h.BrowserCtx == nil {
		h.writeError(w, http.StatusServiceUnavailable, fmt.Errorf("browser context is not configured"))
		return
	}
	if req.QuickReplyOptionID != "" {
		if err := h.BrowserCtx.ClickBotButton(r.Context(), id, req.QuickReplyOptionID); err != nil {
			h.writeError(w, http.StatusBadGateway, fmt.Errorf("send negotiation quick reply: %w", err))
			return
		}
	} else {
		if err := h.BrowserCtx.SendMessage(r.Context(), id, req.Text); err != nil {
			h.writeError(w, http.StatusBadGateway, fmt.Errorf("send negotiation reply: %w", err))
			return
		}
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
	if h.BrowserCtx == nil {
		h.writeError(w, http.StatusServiceUnavailable, fmt.Errorf("browser context is not configured"))
		return
	}
	if err := h.BrowserCtx.PublishResume(r.Context(), h.ResumeID); err != nil {
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
