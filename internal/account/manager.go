package account

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"hh-autoresponder/internal/hh"
)

type Preferences struct {
	MinSalary      int      `json:"min_salary"`
	Schedule       string   `json:"schedule"`
	ExcludeAgency  bool     `json:"exclude_agency"`
	BlacklistIDs   []string `json:"blacklist_ids"`
	BlacklistNames []string `json:"blacklist_names"`
}

type Account struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Token       hh.OAuthToken `json:"token"`
	ResumeIDs   []string      `json:"resume_ids"`
	SearchURLs  []string      `json:"search_urls"`
	Preferences Preferences   `json:"preferences"`
	NeedsReauth bool          `json:"needs_reauth"`
}

type Manager struct {
	mu       sync.RWMutex
	accounts []Account
	path     string
}

func NewManager(path string) *Manager {
	return &Manager{path: path}
}

func (m *Manager) Load() error {
	b, err := os.ReadFile(m.path)
	if err != nil {
		return fmt.Errorf("read accounts file: %w", err)
	}
	var items []Account
	if strings.TrimSpace(string(b)) != "" {
		if err := json.Unmarshal(b, &items); err != nil {
			return fmt.Errorf("unmarshal accounts file: %w", err)
		}
	}
	if items == nil {
		items = []Account{}
	}
	m.mu.Lock()
	m.accounts = items
	m.mu.Unlock()
	return nil
}

func (m *Manager) Save() error {
	m.mu.RLock()
	items := append([]Account(nil), m.accounts...)
	m.mu.RUnlock()

	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal accounts: %w", err)
	}
	if err := os.WriteFile(m.path, b, 0o644); err != nil {
		return fmt.Errorf("write accounts file: %w", err)
	}
	return nil
}

func (m *Manager) GetAll() []Account {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]Account(nil), m.accounts...)
}

func (m *Manager) GetActive() []Account {
	now := time.Now()
	m.mu.RLock()
	defer m.mu.RUnlock()
	active := make([]Account, 0, len(m.accounts))
	for _, a := range m.accounts {
		if a.NeedsReauth {
			continue
		}
		if a.Token.AccessToken == "" {
			continue
		}
		if !a.Token.ExpiresAt.IsZero() && now.After(a.Token.ExpiresAt) {
			continue
		}
		active = append(active, a)
	}
	return active
}

func (m *Manager) Update(mutator func(items *[]Account) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := mutator(&m.accounts); err != nil {
		return fmt.Errorf("mutate accounts: %w", err)
	}
	return nil
}
