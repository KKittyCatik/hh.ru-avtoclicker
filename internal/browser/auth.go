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

	if _, err := ac.Page.Goto("https://hh.ru/account/login"); err != nil {
		return fmt.Errorf("open login page: %w", err)
	}
	if err := randomPause(ctx, 500*time.Millisecond, 2*time.Second); err != nil {
		return fmt.Errorf("pause after goto: %w", err)
	}

	emailSelectors := []string{`[data-qa="account-signup-email"]`, `input[name="email"]`, `input[type="email"]`}
	if err := fillFirstSelector(ac, emailSelectors, email); err != nil {
		return fmt.Errorf("fill email: %w", err)
	}

	if loginByPasswordBtn, _ := ac.Page.QuerySelector(`text=Войти по паролю`); loginByPasswordBtn != nil {
		_ = ac.clickLikeHuman(ctx, `text=Войти по паролю`)
	}

	passwordSelectors := []string{`[data-qa="account-login-password"]`, `input[name="password"]`, `input[type="password"]`}
	if err := typeFirstSelector(ac, passwordSelectors, password, 90); err != nil {
		return fmt.Errorf("fill password: %w", err)
	}

	submitSelectors := []string{`[data-qa="account-login-submit"]`, `button[type="submit"]`}
	if err := clickFirstSelector(ctx, ac, submitSelectors); err != nil {
		return fmt.Errorf("submit login: %w", err)
	}

	if err := ac.Page.WaitForURL("https://hh.ru/**", playwright.PageWaitForURLOptions{Timeout: playwright.Float(15000)}); err != nil {
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
