package worker

import (
	"context"
	"fmt"
	"log/slog"

	"hh-autoresponder/internal/bot_detector"
	"hh-autoresponder/internal/hh"
	"hh-autoresponder/internal/llm"
	"hh-autoresponder/internal/transport/ws"
)

type ReplyWorker struct {
	hhClient  *hh.Client
	llmClient llm.LLMClient
	hub       *ws.Hub
	logger    *slog.Logger
}

func NewReplyWorker(hhClient *hh.Client, llmClient llm.LLMClient, hub *ws.Hub, logger *slog.Logger) *ReplyWorker {
	if logger == nil {
		logger = slog.Default()
	}
	return &ReplyWorker{hhClient: hhClient, llmClient: llmClient, hub: hub, logger: logger}
}

func (w *ReplyWorker) ProcessNegotiations(ctx context.Context, resume string, vacancyDescription string) error {
	negotiations, err := w.hhClient.ListNegotiations(ctx)
	if err != nil {
		return fmt.Errorf("list negotiations for reply worker: %w", err)
	}
	for _, n := range negotiations {
		if !n.NeedsReply {
			continue
		}
		messages, err := w.hhClient.ListNegotiationMessages(ctx, n.ID)
		if err != nil {
			return fmt.Errorf("list messages for %s: %w", n.ID, err)
		}
		if len(messages) == 0 {
			continue
		}
		last := messages[len(messages)-1]
		if bot_detector.IsBot(last) && len(last.Options) > 0 {
			err := w.hhClient.SendNegotiationReply(ctx, n.ID, hh.SendReplyRequest{QuickReplyOptionID: last.Options[0].ID})
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
		if err := w.hhClient.SendNegotiationReply(ctx, n.ID, hh.SendReplyRequest{Text: reply}); err != nil {
			return fmt.Errorf("send text reply for %s: %w", n.ID, err)
		}
		_ = w.hub.Broadcast(ctx, "new_message", map[string]any{"negotiation_id": n.ID, "reply_type": "auto_text"})
	}
	return nil
}
