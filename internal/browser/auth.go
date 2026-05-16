package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// All known selectors for hh.ru login page elements (ordered by preference).
var (
	loginEmailSelectors = []string{
		`[data-qa="login-input-username"]`,
		`[data-qa="account-signup-email"]`,
		`input[name="login"]`,
		`input[autocomplete="username"]`,
		`input[type="email"]`,
		`input[name="email"]`,
	}
	loginByPasswordSelectors = []string{
		`[data-qa="expand-login-by-password"]`,
		`[data-qa="login-by-password"]`,
		`text=Войти по паролю`,
	}
	loginPasswordSelectors = []string{
		`[data-qa="login-input-password"]`,
		`[data-qa="account-login-password"]`,
		`input[name="password"]`,
		`input[type="password"]`,
	}
	loginSubmitSelectors = []string{
		`[data-qa="login-submit"]`,
		`[data-qa="account-login-submit"]`,
		`button[type="submit"]`,
	}
	loggedInSelectors = []string{
		`[data-qa="mainmenu_applicantProfile"]`,
		`[data-qa="mainmenu_vacancyResponses"]`,
		`[data-qa="user-logged-in"]`,
		`a[href*="/applicant/"]`,
		`a[href*="/my/"]`,
	}
)

func (ac *AccountContext) Login(email, password string) error {
	ctx, cancel := withTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if _, err := ac.Page.Goto("https://hh.ru/account/login",
		playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded},
	); err != nil {
		return fmt.Errorf("open login page: %w", err)
	}

	// Wait for page to fully settle
	if err := randomPause(ctx, 1500*time.Millisecond, 3*time.Second); err != nil {
		return fmt.Errorf("pause after goto: %w", err)
	}

	// Wait until at least one email input selector is visible (up to 15s)
	if err := ac.waitForAnySelector(loginEmailSelectors, 15*time.Second); err != nil {
		ac.screenshotDebug("login_no_email_input")
		ac.htmlDebug("login_no_email_input")
		return fmt.Errorf("login page did not load: no email input found")
	}

	if err := fillFirstSelector(ac, loginEmailSelectors, email); err != nil {
		ac.screenshotDebug("login_fill_email_failed")
		ac.htmlDebug("login_fill_email_failed")
		return fmt.Errorf("fill email: %w", err)
	}

	if err := randomPause(ctx, 300*time.Millisecond, 800*time.Millisecond); err != nil {
		return err
	}

	// Click "Login by password" button if present
	for _, sel := range loginByPasswordSelectors {
		if btn, _ := ac.Page.QuerySelector(sel); btn != nil {
			_ = ac.clickLikeHuman(ctx, sel)
			_ = randomPause(ctx, 500*time.Millisecond, 1*time.Second)
			break
		}
	}

	// Wait for password field to appear
	if err := ac.waitForAnySelector(loginPasswordSelectors, 10*time.Second); err != nil {
		ac.screenshotDebug("login_no_password_input")
		ac.htmlDebug("login_no_password_input")
		return fmt.Errorf("password field did not appear: %w", err)
	}

	if err := typeFirstSelector(ac, loginPasswordSelectors, password, 90); err != nil {
		ac.screenshotDebug("login_fill_password_failed")
		return fmt.Errorf("fill password: %w", err)
	}

	if err := randomPause(ctx, 300*time.Millisecond, 800*time.Millisecond); err != nil {
		return err
	}

	if err := clickFirstSelector(ctx, ac, loginSubmitSelectors); err != nil {
		ac.screenshotDebug("login_submit_failed")
		return fmt.Errorf("submit login: %w", err)
	}

	if err := ac.Page.WaitForURL("https://hh.ru/**", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(loginWaitURLTimeoutMs),
	}); err != nil {
		ac.screenshotDebug("login_redirect_timeout")
		return fmt.Errorf("wait for post-login redirect: %w", err)
	}

	if err := ac.waitCaptchaIfNeeded(ctx); err != nil {
		return err
	}

	logged, err := ac.IsLoggedIn()
	if err != nil {
		return fmt.Errorf("check login status: %w", err)
	}
	if !logged {
		ac.screenshotDebug("login_not_logged_in")
		ac.htmlDebug("login_not_logged_in")
		return fmt.Errorf("login failed: profile element not found after redirect")
	}
	ac.LoggedIn = true
	return ac.SaveSession()
}

func (ac *AccountContext) IsLoggedIn() (bool, error) {
	for _, selector := range loggedInSelectors {
		el, err := ac.Page.QuerySelector(selector)
		if err != nil {
			continue
		}
		if el != nil {
			ac.LoggedIn = true
			return true, nil
		}
	}
	ac.LoggedIn = false
	return false, nil
}

func (ac *AccountContext) SaveSession() error {
	statePath := ac.sessionPath()
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return fmt.Errorf("create session folder: %w", err)
	}
	if _, err := ac.Context.StorageState(statePath); err != nil {
		return fmt.Errorf("save storage state: %w", err)
	}
	return nil
}

// waitForAnySelector waits until at least one selector from the list becomes visible.
func (ac *AccountContext) waitForAnySelector(selectors []string, timeout time.Duration) error {
	perSelector := timeout / time.Duration(len(selectors))
	if perSelector < 3*time.Second {
		perSelector = 3 * time.Second
	}
	for _, sel := range selectors {
		_, err := ac.Page.WaitForSelector(sel, playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(float64(perSelector.Milliseconds())),
			State:   playwright.WaitForSelectorStateVisible,
		})
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("none of selectors appeared: %s", strings.Join(selectors, ", "))
}

// screenshotDebug saves a PNG screenshot to the debug directory.
func (ac *AccountContext) screenshotDebug(name string) {
	_ = os.MkdirAll("debug", 0o755)
	path := fmt.Sprintf("debug/%s_%s.png", name, time.Now().Format("20060102_150405"))
	_, _ = ac.Page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(path),
		FullPage: playwright.Bool(true),
	})
	if ac.logger != nil {
		ac.logger.Info("debug screenshot saved", "path", path)
	}
}

// htmlDebug saves the current page URL and HTML to a text file for terminal inspection.
func (ac *AccountContext) htmlDebug(name string) {
	_ = os.MkdirAll("debug", 0o755)
	ts := time.Now().Format("20060102_150405")

	// Save current URL
	currentURL := ac.Page.URL()

	// Get page HTML
	html, err := ac.Page.Content()
	if err != nil {
		html = fmt.Sprintf("error getting content: %v", err)
	}

	// Get page title
	title, _ := ac.Page.Title()

	// Write summary file (small, readable in terminal)
	summaryPath := fmt.Sprintf("debug/%s_%s.txt", name, ts)
	summary := fmt.Sprintf("URL: %s\nTitle: %s\n\n--- HTML (first 4000 chars) ---\n%s\n",
		currentURL, title, truncate(html, 4000))
	_ = os.WriteFile(summaryPath, []byte(summary), 0o644)

	if ac.logger != nil {
		ac.logger.Info("debug html saved", "path", summaryPath, "url", currentURL, "title", title)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n... (truncated)"
}

func (ac *AccountContext) waitCaptchaIfNeeded(ctx context.Context) error {
	captcha, err := ac.Page.QuerySelector(`[data-qa="account-captcha"]`)
	if err != nil {
		return fmt.Errorf("check captcha: %w", err)
	}
	if captcha == nil {
		return nil
	}
	if ac.notify != nil {
		_ = ac.notify(ctx, "captcha_required", map[string]any{"accountID": ac.AccountID})
	}

	waitCtx, cancel := withTimeout(ctx, 120*time.Second)
	defer cancel()

	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("captcha not solved in time")
		case <-t.C:
			again, err := ac.Page.QuerySelector(`[data-qa="account-captcha"]`)
			if err != nil {
				return fmt.Errorf("poll captcha: %w", err)
			}
			if again == nil {
				return nil
			}
		}
	}
}

func fillFirstSelector(ac *AccountContext, selectors []string, value string) error {
	for _, s := range selectors {
		el, err := ac.Page.QuerySelector(s)
		if err != nil || el == nil {
			continue
		}
		if err := ac.Page.Fill(s, value); err == nil {
			return nil
		}
	}
	return fmt.Errorf("none of selectors found: %s", strings.Join(selectors, ", "))
}

func typeFirstSelector(ac *AccountContext, selectors []string, value string, delay float64) error {
	for _, s := range selectors {
		el, err := ac.Page.QuerySelector(s)
		if err != nil || el == nil {
			continue
		}
		if err := ac.typeLikeHuman(s, value, delay); err == nil {
			return nil
		}
	}
	return fmt.Errorf("none of selectors found: %s", strings.Join(selectors, ", "))
}

func clickFirstSelector(ctx context.Context, ac *AccountContext, selectors []string) error {
	for _, s := range selectors {
		if err := ac.clickLikeHuman(ctx, s); err == nil {
			return nil
		}
	}
	return fmt.Errorf("none of selectors clickable: %s", strings.Join(selectors, ", "))
}
