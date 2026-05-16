package httptransport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"hh-autoresponder/internal/account"
	"hh-autoresponder/internal/hh"
	"hh-autoresponder/internal/monitor"
	"hh-autoresponder/internal/worker"
)

type Handlers struct {
	Ctx         context.Context
	Auth        *hh.AuthManager
	ApplyWorker *worker.ApplyWorker
	ReplyWorker *worker.ReplyWorker
	Stats       *monitor.Collector
	Accounts    *account.Manager
	HHClient    *hh.Client
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
	if h.Auth == nil {
		h.writeError(w, http.StatusServiceUnavailable, fmt.Errorf("oauth is not configured"))
		return
	}

	state, err := generateRandomState()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Errorf("generate oauth state: %w", err))
		return
	}
	h.states.Set(state)
	http.Redirect(w, r, h.Auth.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *Handlers) AuthCallback(w http.ResponseWriter, r *http.Request) {
	if h.Auth == nil {
		h.writeError(w, http.StatusServiceUnavailable, fmt.Errorf("oauth is not configured"))
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if code == "" || state == "" {
		h.writeError(w, http.StatusBadRequest, fmt.Errorf("missing oauth callback parameters"))
		return
	}
	if !h.states.Validate(state) {
		h.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid oauth state"))
		return
	}

	ctx := r.Context()
	token, err := h.Auth.ExchangeCode(ctx, code)
	if err != nil {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("exchange oauth code: %w", err))
		return
	}

	me, err := h.HHClient.GetMe(ctx)
	if err != nil {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("fetch profile: %w", err))
		return
	}
	if strings.TrimSpace(me.ID) == "" {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("fetch profile: missing user id"))
		return
	}
	resumes, err := h.HHClient.GetMyResumes(ctx)
	if err != nil {
		h.writeError(w, http.StatusBadGateway, fmt.Errorf("fetch resumes: %w", err))
		return
	}

	resumeIDs := make([]string, 0, len(resumes))
	for _, resume := range resumes {
		if resume.ID != "" {
			resumeIDs = append(resumeIDs, resume.ID)
		}
	}

	name := strings.TrimSpace(me.FirstName + " " + me.LastName)
	if name == "" {
		name = strings.TrimSpace(me.Email)
	}
	if name == "" {
		name = me.ID
	}

	newAccount := account.Account{
		ID:        me.ID,
		Name:      name,
		Token:     token,
		ResumeIDs: resumeIDs,
	}

	if err := h.Accounts.Update(func(items *[]account.Account) error {
		for i := range *items {
			if (*items)[i].ID != me.ID {
				continue
			}
			existing := (*items)[i]
			existing.Name = newAccount.Name
			existing.Token = newAccount.Token
			existing.ResumeIDs = newAccount.ResumeIDs
			existing.NeedsReauth = false
			(*items)[i] = existing
			return nil
		}
		*items = append(*items, newAccount)
		return nil
	}); err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Errorf("update account: %w", err))
		return
	}
	if err := h.Accounts.Save(); err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Errorf("save accounts: %w", err))
		return
	}

	h.Auth.SetToken(token)
	h.ResumeID = ""
	if len(resumeIDs) > 0 {
		h.ResumeID = resumeIDs[0]
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

func generateRandomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (h *Handlers) writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handlers) writeError(w http.ResponseWriter, code int, err error) {
	h.writeJSON(w, code, map[string]string{"error": err.Error()})
}
