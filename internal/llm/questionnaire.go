package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"hh-autoresponder/internal/hh"
)

var numberPattern = regexp.MustCompile(`\d+`)

func FillQuestionnaire(ctx context.Context, client LLMClient, resume string, questions []hh.Question) ([]hh.Answer, error) {
	answers := make([]hh.Answer, 0, len(questions))

	for _, q := range questions {
		switch q.Type {
		case "radio", "checkbox":
			ans, err := chooseOption(ctx, client, resume, q)
			if err != nil {
				return nil, fmt.Errorf("answer choice question %s: %w", q.ID, err)
			}
			answers = append(answers, ans)
		case "text":
			ans, err := answerText(ctx, client, resume, q)
			if err != nil {
				return nil, fmt.Errorf("answer text question %s: %w", q.ID, err)
			}
			answers = append(answers, ans)
		case "number":
			ans, err := answerNumber(resume, q)
			if err != nil {
				return nil, fmt.Errorf("answer number question %s: %w", q.ID, err)
			}
			answers = append(answers, ans)
		default:
			return nil, fmt.Errorf("unsupported question type: %s", q.Type)
		}
	}

	return answers, nil
}

func chooseOption(ctx context.Context, client LLMClient, resume string, q hh.Question) (hh.Answer, error) {
	options := make([]string, 0, len(q.Options))
	for _, o := range q.Options {
		options = append(options, fmt.Sprintf("%s: %s", o.ID, o.Text))
	}
	user := fmt.Sprintf("Резюме:\n%s\n\nВопрос: %s\nВарианты:\n%s\n\nВерни JSON: {\"option_ids\":[\"id\"]}", resume, q.Text, strings.Join(options, "\n"))
	resp, err := client.Complete(ctx, "Выбирай только из предложенных вариантов. Возвращай только JSON.", user)
	if err != nil {
		return hh.Answer{}, fmt.Errorf("complete questionnaire choice: %w", err)
	}
	var parsed struct {
		OptionIDs []string `json:"option_ids"`
	}
	if err := json.Unmarshal([]byte(resp), &parsed); err != nil {
		return hh.Answer{}, fmt.Errorf("unmarshal option response: %w", err)
	}
	ans := hh.Answer{QuestionID: q.ID, OptionIDs: parsed.OptionIDs, NeedsReview: lowConfidence(resp)}
	if q.Type == "radio" && len(ans.OptionIDs) > 1 {
		ans.OptionIDs = ans.OptionIDs[:1]
	}
	return ans, nil
}

func answerText(ctx context.Context, client LLMClient, resume string, q hh.Question) (hh.Answer, error) {
	user := fmt.Sprintf("Резюме:\n%s\n\nВопрос: %s\n\nОтветь максимум тремя предложениями.", resume, q.Text)
	resp, err := client.Complete(ctx, "Кратко ответь на вопрос работодателя на русском.", user)
	if err != nil {
		return hh.Answer{}, fmt.Errorf("complete questionnaire text: %w", err)
	}
	text := enforceThreeSentences(resp)
	return hh.Answer{QuestionID: q.ID, Text: text, NeedsReview: lowConfidence(text)}, nil
}

func answerNumber(resume string, q hh.Question) (hh.Answer, error) {
	nums := numberPattern.FindAllString(resume, -1)
	if len(nums) == 0 {
		return hh.Answer{QuestionID: q.ID, Number: 0, NeedsReview: true}, nil
	}
	n, err := strconv.Atoi(nums[0])
	if err != nil {
		return hh.Answer{}, fmt.Errorf("parse number from resume: %w", err)
	}
	return hh.Answer{QuestionID: q.ID, Number: n}, nil
}

func lowConfidence(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "не знаю") || strings.Contains(lower, "не уверен")
}

func enforceThreeSentences(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, ".")
	clean := make([]string, 0, 3)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		clean = append(clean, p)
		if len(clean) == 3 {
			break
		}
	}
	if len(clean) == 0 {
		if strings.HasSuffix(trimmed, ".") || strings.HasSuffix(trimmed, "!") || strings.HasSuffix(trimmed, "?") {
			return trimmed
		}
		return trimmed + "."
	}
	return strings.Join(clean, ". ") + "."
}
