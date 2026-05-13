package llm

import (
	"context"
	"fmt"
	"strings"

	"hh-autoresponder/internal/hh"
)

const hrReplySystemPrompt = "Ты профессиональный соискатель. Отвечай кратко, по делу, вежливо на русском языке. Не придумывай факты о себе. Если вопрос требует уточнения — верни текст с префиксом [NEEDS_INPUT]."

func GenerateHRReply(ctx context.Context, client LLMClient, resume, vacancyDesc string, history []hh.Message) (string, bool, error) {
	h := history
	if len(h) > 10 {
		h = h[len(h)-10:]
	}
	lines := make([]string, 0, len(h)+3)
	lines = append(lines, "Резюме:", resume, "", "Описание вакансии:", vacancyDesc, "", "История:")
	for _, m := range h {
		prefix := "Я"
		if strings.EqualFold(m.From, "hr") {
			prefix = "HR"
		}
		lines = append(lines, prefix+": "+m.Text)
	}

	completion, err := client.Complete(ctx, hrReplySystemPrompt, strings.Join(lines, "\n"))
	if err != nil {
		return "", false, fmt.Errorf("generate hr reply: %w", err)
	}
	needsInput := strings.HasPrefix(strings.TrimSpace(completion), "[NEEDS_INPUT]")
	return completion, needsInput, nil
}
