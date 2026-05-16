package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type ApplyResult struct {
	Success          bool
	VacancyID        string
	AlreadyApplied   bool
	HasQuestionnaire bool
	Error            string
	Timestamp        time.Time
}

func (ac *AccountContext) ApplyToVacancy(ctx context.Context, vacancyID string, resumeID string, coverLetter string) (*ApplyResult, error) {
	tctx, cancel := withTimeout(ctx, 90*time.Second)
	defer cancel()

	result := &ApplyResult{VacancyID: vacancyID, Timestamp: time.Now()}
	vacancyURL := fmt.Sprintf("https://hh.ru/vacancy/%s", vacancyID)
	if _, err := ac.Page.Goto(vacancyURL); err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("open vacancy: %w", err)
	}
	if _, err := waitForSelectorWithDebug(ac.Page, `[data-qa="vacancy-response-link-top"]`, 20*time.Second); err != nil {
		result.Error = err.Error()
		return result, err
	}

	content, _ := ac.Page.Content()
	if strings.Contains(content, "Вы откликались") {
		result.Success = true
		result.AlreadyApplied = true
		return result, nil
	}

	if err := ac.clickLikeHuman(tctx, `[data-qa="vacancy-response-link-top"]`); err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("click response button: %w", err)
	}
	if _, err := waitForSelectorWithDebug(ac.Page, `[data-qa="vacancy-response-submit-popup"]`, 20*time.Second); err != nil {
		result.Error = err.Error()
		return result, err
	}

	if resumeID != "" {
		if _, err := ac.Page.SelectOption(`[data-qa="resume-select"]`, playwright.SelectOptionValues{Values: &[]string{resumeID}}); err != nil {
			ac.logger.Info("resume select skipped", "error", err, "resume_id", resumeID)
		}
	}
	if strings.TrimSpace(coverLetter) != "" {
		if err := ac.typeLikeHuman(`[data-qa="vacancy-response-popup-cover-letter"]`, coverLetter, 70); err != nil {
			result.Error = err.Error()
			return result, fmt.Errorf("type cover letter: %w", err)
		}
	}

	questionnaire, _ := ac.Page.QuerySelector(`[data-qa="task-body"]`)
	if questionnaire != nil {
		result.HasQuestionnaire = true
	}

	if err := ac.clickLikeHuman(tctx, `[data-qa="vacancy-response-submit-popup"]`); err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("submit response: %w", err)
	}

	if _, err := waitForSelectorWithDebug(ac.Page, `text=Отклик отправлен`, 20*time.Second); err != nil {
		result.Error = err.Error()
		return result, err
	}
	if err := randomPause(tctx, 3*time.Second, 7*time.Second); err != nil {
		result.Error = err.Error()
		return result, err
	}
	result.Success = true
	return result, nil
}
