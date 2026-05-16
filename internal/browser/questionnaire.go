package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"

	"hh-autoresponder/internal/hh"
	"hh-autoresponder/internal/llm"
)

func (ac *AccountContext) FillQuestionnaire(ctx context.Context, llmClient llm.LLMClient, resume string) error {
	tctx, cancel := withTimeout(ctx, defaultOpTimeout)
	defer cancel()

	items, err := ac.Page.QuerySelectorAll(`[data-qa="task-body"]`)
	if err != nil {
		return fmt.Errorf("query questionnaire blocks: %w", err)
	}
	if len(items) == 0 {
		return nil
	}

	questions := make([]hh.Question, 0, len(items))
	for i, item := range items {
		text, _ := item.TextContent()
		qType := "text"
		if radio, _ := item.QuerySelector(`input[type="radio"]`); radio != nil {
			qType = "radio"
		}
		if checkbox, _ := item.QuerySelector(`input[type="checkbox"]`); checkbox != nil {
			qType = "checkbox"
		}
		options := []hh.QuestionOption{}
		optionNodes, _ := item.QuerySelectorAll("label")
		for oi, optionNode := range optionNodes {
			ot, _ := optionNode.TextContent()
			ot = strings.TrimSpace(ot)
			if ot == "" {
				continue
			}
			options = append(options, hh.QuestionOption{ID: fmt.Sprintf("q%d_o%d", i, oi), Text: ot})
		}
		questions = append(questions, hh.Question{ID: fmt.Sprintf("q%d", i), Type: qType, Text: strings.TrimSpace(text), Options: options})
	}

	answers, err := llm.FillQuestionnaire(tctx, llmClient, resume, questions)
	if err != nil {
		return fmt.Errorf("llm questionnaire answers: %w", err)
	}
	for i, ans := range answers {
		if i >= len(items) {
			break
		}
		item := items[i]
		switch {
		case ans.Text != "":
			if input, _ := item.QuerySelector(`textarea, input[type="text"]`); input != nil {
				_ = input.Type(ans.Text, playwright.ElementHandleTypeOptions{Delay: playwright.Float(70)})
			}
		case len(ans.OptionIDs) > 0:
			labels, _ := item.QuerySelectorAll("label")
			if len(labels) > 0 {
				_ = labels[0].Click()
			}
		}
		if err := randomPause(tctx, 500*time.Millisecond, 1500*time.Millisecond); err != nil {
			return err
		}
	}
	return nil
}
