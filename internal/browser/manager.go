package browser

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

const (
	defaultOpTimeout = 30 * time.Second
	// 15000ms (15s) timeout for login redirect wait after submit.
	loginWaitURLTimeoutMs = 15000
)

// BrowserManager управляет пулом браузерных контекстов (по одному на аккаунт)
type BrowserManager struct {
	pw       *playwright.Playwright
	browser  playwright.Browser
	contexts map[string]*AccountContext
	mu       sync.RWMutex
	logger   *slog.Logger
	notify   func(context.Context, string, any) error
}

type AccountContext struct {
	AccountID string
	Context   playwright.BrowserContext
	Page      playwright.Page
	LoggedIn  bool
	logger    *slog.Logger
	notify    func(context.Context, string, any) error
}

func NewBrowserManager(ctx context.Context) (*BrowserManager, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("init browser manager canceled: %w", ctx.Err())
		default:
		}
	}
	if err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
		Verbose:  false,
	}); err != nil {
		return nil, fmt.Errorf("install playwright: %w", err)
	}
	if err := os.MkdirAll("debug", 0o755); err != nil {
		return nil, fmt.Errorf("create debug dir: %w", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("run playwright: %w", err)
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			_ = pw.Stop()
			return nil, fmt.Errorf("init browser manager canceled: %w", ctx.Err())
		default:
		}
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		_ = pw.Stop()
		return nil, fmt.Errorf("launch chromium: %w", err)
	}

	return &BrowserManager{
		pw:       pw,
		browser:  browser,
		contexts: make(map[string]*AccountContext),
		logger:   slog.Default(),
	}, nil
}

func (m *BrowserManager) AddAccount(accountID string) (*AccountContext, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.contexts[accountID]; ok {
		return existing, nil
	}

	sessionDir := filepath.Join("data", "sessions", accountID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}
	statePath := filepath.Join(sessionDir, "storage_state.json")
	var statePathOption *string
	if _, statErr := os.Stat(statePath); statErr == nil {
		statePathOption = playwright.String(statePath)
	}

	ctx, err := m.browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent:        playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		Viewport:         &playwright.Size{Width: 1366, Height: 768},
		Locale:           playwright.String("ru-RU"),
		TimezoneId:       playwright.String("Europe/Moscow"),
		StorageStatePath: statePathOption,
	})
	if err != nil {
		return nil, fmt.Errorf("create browser context: %w", err)
	}

	page, err := ctx.NewPage()
	if err != nil {
		_ = ctx.Close()
		return nil, fmt.Errorf("create page: %w", err)
	}

	ac := &AccountContext{AccountID: accountID, Context: ctx, Page: page, logger: m.logger, notify: m.notify}
	m.contexts[accountID] = ac
	return ac, nil
}

func (m *BrowserManager) GetContext(accountID string) (*AccountContext, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ac, ok := m.contexts[accountID]
	if !ok {
		return nil, fmt.Errorf("account context %s not found", accountID)
	}
	return ac, nil
}

func (m *BrowserManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, ctx := range m.contexts {
		_ = ctx.SaveSession()
		if ctx.Context != nil {
			_ = ctx.Context.Close()
		}
	}
	m.contexts = map[string]*AccountContext{}

	if m.browser != nil {
		if err := m.browser.Close(); err != nil {
			return fmt.Errorf("close browser: %w", err)
		}
	}
	if m.pw != nil {
		if err := m.pw.Stop(); err != nil {
			return fmt.Errorf("stop playwright: %w", err)
		}
	}
	return nil
}

func (m *BrowserManager) SetNotifier(fn func(context.Context, string, any) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notify = fn
	for _, ac := range m.contexts {
		ac.notify = fn
	}
}

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), timeout)
	}
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func randomDuration(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}
	delta := max - min
	return min + time.Duration(rand.Int63n(int64(delta)))
}

func randomPause(ctx context.Context, min, max time.Duration) error {
	d := randomDuration(min, max)
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func waitForSelectorWithDebug(page playwright.Page, selector string, timeout time.Duration) (playwright.ElementHandle, error) {
	handle, err := page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(float64(timeout.Milliseconds()))})
	if err == nil {
		return handle, nil
	}
	_ = os.MkdirAll("debug", 0o755)
	path := fmt.Sprintf("debug/err_%s.png", time.Now().Format("20060102_150405.000"))
	_, _ = page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(path)})
	return nil, fmt.Errorf("wait for selector %s: %w", selector, err)
}

func (ac *AccountContext) typeLikeHuman(selector, value string, delayMs float64) error {
	if err := ac.Page.Type(selector, value, playwright.PageTypeOptions{Delay: playwright.Float(delayMs)}); err == nil {
		return nil
	}
	return ac.Page.Fill(selector, value)
}

func (ac *AccountContext) clickLikeHuman(ctx context.Context, selector string) error {
	el, err := ac.Page.QuerySelector(selector)
	if err != nil {
		return fmt.Errorf("query %s: %w", selector, err)
	}
	if el == nil {
		return fmt.Errorf("element %s not found", selector)
	}
	_, _ = ac.Page.Evaluate("(el) => el.scrollIntoView({block: 'center'})", el)
	if err := randomPause(ctx, 500*time.Millisecond, 2*time.Second); err != nil {
		return fmt.Errorf("pause before click: %w", err)
	}
	if err := ac.Page.Click(selector); err != nil {
		return fmt.Errorf("click %s: %w", selector, err)
	}
	return nil
}

func (ac *AccountContext) sessionPath() string {
	return filepath.Join("data", "sessions", ac.AccountID, "storage_state.json")
}
