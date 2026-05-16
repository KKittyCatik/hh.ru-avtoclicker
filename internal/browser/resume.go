package browser

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ErrPublishNotAvailable = errors.New("resume publish not available yet")

var nextPublishPattern = regexp.MustCompile(`Можно поднять через\s*(?:(\d+)\s*час)?\s*(?:(\d+)\s*мин)?`)

func (ac *AccountContext) PublishResume(ctx context.Context, resumeID string) error {
	tctx, cancel := withTimeout(ctx, defaultOpTimeout)
	defer cancel()

	if _, err := ac.Page.Goto(fmt.Sprintf("https://hh.ru/resume/%s", resumeID)); err != nil {
		return fmt.Errorf("open resume page: %w", err)
	}
	btn, err := waitForSelectorWithDebug(ac.Page, `[data-qa="resume-update-button"]`, 20*time.Second)
	if err != nil {
		return err
	}
	disabled, err := btn.GetAttribute("disabled")
	if err != nil {
		return fmt.Errorf("read publish button state: %w", err)
	}
	if disabled != "" {
		return ErrPublishNotAvailable
	}
	_, _ = ac.Page.Evaluate("(el) => el.scrollIntoView({block: 'center'})", btn)
	if err := randomPause(tctx, 500*time.Millisecond, 2*time.Second); err != nil {
		return err
	}
	if err := btn.Click(); err != nil {
		return fmt.Errorf("click publish button: %w", err)
	}
	if _, err := waitForSelectorWithDebug(ac.Page, `text=Обновлено только что`, 20*time.Second); err != nil {
		return err
	}
	return nil
}

func (ac *AccountContext) GetNextPublishTime(ctx context.Context, resumeID string) (time.Time, error) {
	_, cancel := withTimeout(ctx, defaultOpTimeout)
	defer cancel()

	if _, err := ac.Page.Goto(fmt.Sprintf("https://hh.ru/resume/%s", resumeID)); err != nil {
		return time.Time{}, fmt.Errorf("open resume page: %w", err)
	}
	pageText, err := ac.Page.TextContent("body")
	if err != nil {
		return time.Time{}, fmt.Errorf("read page text: %w", err)
	}
	m := nextPublishPattern.FindStringSubmatch(pageText)
	if len(m) == 0 {
		return time.Now(), nil
	}
	var (
		hours   int
		minutes int
	)
	if len(m) > 1 && strings.TrimSpace(m[1]) != "" {
		hours, _ = strconv.Atoi(m[1])
	}
	if len(m) > 2 && strings.TrimSpace(m[2]) != "" {
		minutes, _ = strconv.Atoi(m[2])
	}
	return time.Now().Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute), nil
}
