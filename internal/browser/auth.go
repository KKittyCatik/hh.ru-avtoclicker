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

func (ac *AccountContext) Login(email, password string) error {
	ctx, cancel := withTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if _, err := ac.Page.Goto("https://hh.ru/account/login", playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateLoad,
	}); err != nil {
		return fmt.Errorf("open login page: %w", err)
	}

	emailSelectors := []string{
		`[data-qa="login-input-username"]`,
		`[data-qa="account-signup-email"]`,
		`input[name="login"]`,
		`input[autocomplete="username"]`,
		`input[type="email"]`,
	}
	var emailFound bool
	for _, sel := range emailSelectors {
		if _, err := ac.Page.WaitForSelector(sel, playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(10000),
			State:   playwright.WaitForSelectorStateVisible,
		}); err == nil {
			emailFound = true
			break
		}
	}
	if !emailFound {
		_, _ = ac.Page.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String(fmt.Sprintf("debug/login_err_%s.png", time.Now().Format("20060102_150405.000"))),
		})
		return fmt.Errorf("login page did not load: no email input found")
	}
	if err := randomPause(ctx, 500*time.Millisecond, 2*time.Second); err != nil {
		return fmt.Errorf("pause after page load: %w", err)
	}

	if err := fillFirstSelector(ac, emailSelectors, email); err != nil {
		return fmt.Errorf("fill email: %w", err)
	}

	byPasswordSelectors := []string{
		`[data-qa="expand-login-by-password"]`,
		`[data-qa="login-by-password"]`,
		`text=Войти по паролю`,
	}
	for _, sel := range byPasswordSelectors {
		if btn, _ := ac.Page.QuerySelector(sel); btn != nil {
			_ = ac.clickLikeHuman(ctx, sel)
			_ = randomPause(ctx, 500*time.Millisecond, 1*time.Second)
			break
		}
	}

	passwordSelectors := []string{
		`[data-qa="login-input-password"]`,
		`[data-qa="account-login-password"]`,
		`input[name="password"]`,
		`input[type="password"]`,
	}
	if err := typeFirstSelector(ac, passwordSelectors, password, 90); err != nil {
		return fmt.Errorf("fill password: %w", err)
	}

	submitSelectors := []string{
		`[data-qa="login-submit"]`,
		`[data-qa="account-login-submit"]`,
		`button[type="submit"]`,
	}
	if err := clickFirstSelector(ctx, ac, submitSelectors); err != nil {
		return fmt.Errorf("submit login: %w", err)
	}

	if err := ac.Page.WaitForURL("https://hh.ru/**", playwright.PageWaitForURLOptions{Timeout: playwright.Float(loginWaitURLTimeoutMs)}); err != nil {
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
		return fmt.Errorf("login failed: profile element not found")
	}
	ac.LoggedIn = true
	return ac.SaveSession()
}

func (ac *AccountContext) IsLoggedIn() (bool, error) {
	selectors := []string{`[data-qa="mainmenu_applicantProfile"]`, `[data-qa="mainmenu_vacancyResponses"]`, `a[href*="/applicant/"]`}
	for _, selector := range selectors {
		el, err := ac.Page.QuerySelector(selector)
		if err != nil {
			return false, fmt.Errorf("query login selector %s: %w", selector, err)
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
