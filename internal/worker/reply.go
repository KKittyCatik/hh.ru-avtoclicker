package worker

import (
	"context"
	"fmt"
	"log/slog"

	"hh-autoresponder/internal/bot_detector"
	"hh-autoresponder/internal/browser"
	"hh-autoresponder/internal/llm"
	"hh-autoresponder/internal/transport/ws"
)

type ReplyWorker struct {
	browserCtx *browser.AccountContext
	llmClient  llm.LLMClient
	hub        *ws.Hub
	logger     *slog.Logger
}

type DraftReply struct {
	Text               string `json:"text"`
	QuickReplyOptionID string `json:"quickReplyOptionId,omitempty"`
	NeedsHumanInput    bool   `json:"needsHumanInput"`
	IsBotFlow          bool   `json:"isBotFlow"`
}

func NewReplyWorker(browserCtx *browser.AccountContext, llmClient llm.LLMClient, hub *ws.Hub, logger *slog.Logger) *ReplyWorker {
	if logger == nil {
		logger = slog.Default()
	}
	return &ReplyWorker{browserCtx: browserCtx, llmClient: llmClient, hub: hub, logger: logger}
}

func (w *ReplyWorker) ProcessNegotiations(ctx context.Context, resume string, vacancyDescription string) error {
	if w.browserCtx == nil {
		return fmt.Errorf("browser context is not configured")
	}
	negotiations, err := w.browserCtx.GetNegotiations(ctx)
	if err != nil {
		return fmt.Errorf("list negotiations for reply worker: %w", err)
	}
	for _, n := range negotiations {
		if !n.NeedsReply {
			continue
		}
		messages, err := w.browserCtx.GetMessages(ctx, n.ID)
		if err != nil {
			return fmt.Errorf("list messages for %s: %w", n.ID, err)
		}
		if len(messages) == 0 {
			continue
		}
		last := messages[len(messages)-1]
		if bot_detector.IsBot(last) && len(last.Options) > 0 {
			err := w.browserCtx.ClickBotButton(ctx, n.ID, last.Options[0].Text)
			if err != nil {
				return fmt.Errorf("send quick reply for %s: %w", n.ID, err)
			}
			_ = w.hub.Broadcast(ctx, "new_message", map[string]any{"negotiation_id": n.ID, "reply_type": "quick_reply"})
			continue
		}
		reply, needsInput, err := llm.GenerateHRReply(ctx, w.llmClient, resume, vacancyDescription, messages)
		if err != nil {
			return fmt.Errorf("generate reply for %s: %w", n.ID, err)
		}
		if needsInput {
			w.logger.Info("negotiation requires manual reply", "negotiation_id", n.ID)
			continue
		}
		if err := w.browserCtx.SendMessage(ctx, n.ID, reply); err != nil {
			return fmt.Errorf("send text reply for %s: %w", n.ID, err)
		}
		_ = w.hub.Broadcast(ctx, "new_message", map[string]any{"negotiation_id": n.ID, "reply_type": "auto_text"})
	}
	return nil
}

func (w *ReplyWorker) GenerateDraft(ctx context.Context, negotiationID string) (DraftReply, error) {
	if w.browserCtx == nil {
		return DraftReply{}, fmt.Errorf("browser context is not configured")
	}
	messages, err := w.browserCtx.GetMessages(ctx, negotiationID)
	if err != nil {
		return DraftReply{}, fmt.Errorf("list messages for %s: %w", negotiationID, err)
	}
	if len(messages) == 0 {
		return DraftReply{}, fmt.Errorf("no messages found for %s", negotiationID)
	}

	last := messages[len(messages)-1]
	if bot_detector.IsBot(last) && len(last.Options) > 0 {
		return DraftReply{
			Text:               last.Options[0].Text,
			QuickReplyOptionID: last.Options[0].ID,
			NeedsHumanInput:    false,
			IsBotFlow:          true,
		}, nil
	}

	reply, needsInput, err := llm.GenerateHRReply(ctx, w.llmClient, "", "", messages)
	if err != nil {
		return DraftReply{}, fmt.Errorf("generate reply for %s: %w", negotiationID, err)
	}

	return DraftReply{
		Text:            reply,
		NeedsHumanInput: needsInput,
		IsBotFlow:       false,
	}, nil
}
