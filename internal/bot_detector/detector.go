package bot_detector

import (
	"strings"
	"time"

	"hh-autoresponder/internal/hh"
)

var ctaPhrases = []string{"нажмите", "выберите", "перейдите"}
var firstPersonQuestionWords = []string{"вы", "расскажите", "когда вы"}

func IsBot(msg hh.Message) bool {
	if msg.Type == "question" && len(msg.Options) > 0 {
		return true
	}

	text := strings.ToLower(msg.Text)
	hasCTA := false
	for _, p := range ctaPhrases {
		if strings.Contains(text, p) {
			hasCTA = true
			break
		}
	}
	hasPersonal := false
	for _, p := range firstPersonQuestionWords {
		if strings.Contains(text, p) {
			hasPersonal = true
			break
		}
	}
	if hasCTA && !hasPersonal {
		return true
	}

	if !msg.NegotiationCreatedAt.IsZero() && !msg.CreatedAt.IsZero() {
		if msg.CreatedAt.Sub(msg.NegotiationCreatedAt) < 3*time.Second {
			return true
		}
	}

	return false
}
