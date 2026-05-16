package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type Preferences struct {
	MinSalary      int      `json:"min_salary"`
	Schedule       string   `json:"schedule"`
	ExcludeAgency  bool     `json:"exclude_agency"`
	BlacklistIDs   []string `json:"blacklist_ids"`
	BlacklistNames []string `json:"blacklist_names"`
}

type Account struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	ResumeIDs   []string    `json:"resume_ids"`
	SearchURLs  []string    `json:"search_urls"`
	Preferences Preferences `json:"preferences"`
	NeedsReauth bool        `json:"needs_reauth"`
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
	m.mu.RLock()
	defer m.mu.RUnlock()
	active := make([]Account, 0, len(m.accounts))
	for _, a := range m.accounts {
		if a.NeedsReauth {
			continue
		}
		if strings.TrimSpace(a.Email) == "" {
			continue
		}
		if strings.TrimSpace(a.Password) == "" {
			continue
		}
		active = append(active, a)
	}
	return active
}

func EncryptPassword(plain string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("invalid encryption key size: got %d, expected 32", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func DecryptPassword(encrypted string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("invalid encryption key size: got %d, expected 32", len(key))
	}
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("decode encrypted password: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("encrypted payload too short")
	}
	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt password: %w", err)
	}
	return string(plain), nil
}

func LoadEncryptionKeyFromEnv() ([]byte, error) {
	raw := strings.TrimSpace(os.Getenv("ENCRYPTION_KEY"))
	if raw == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode ENCRYPTION_KEY: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must decode to 32 bytes, got %d", len(key))
	}
	return key, nil
}

func (m *Manager) Update(mutator func(items *[]Account) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := mutator(&m.accounts); err != nil {
		return fmt.Errorf("mutate accounts: %w", err)
	}
	return nil
}
