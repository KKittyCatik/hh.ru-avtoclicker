package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"

	"hh-autoresponder/internal/hh"
)

func (ac *AccountContext) GetNegotiations(ctx context.Context) ([]hh.Negotiation, error) {
	tctx, cancel := withTimeout(ctx, defaultOpTimeout)
	defer cancel()

	if _, err := ac.Page.Goto("https://hh.ru/applicant/negotiations"); err != nil {
		return nil, fmt.Errorf("open negotiations page: %w", err)
	}
	if _, err := waitForSelectorWithDebug(ac.Page, `[data-qa="negotiations-item"]`, 20*time.Second); err != nil {
		return nil, err
	}

	cards, err := ac.Page.QuerySelectorAll(`[data-qa="negotiations-item"]`)
	if err != nil {
		return nil, fmt.Errorf("query negotiations cards: %w", err)
	}
	result := make([]hh.Negotiation, 0, len(cards))
	for _, card := range cards {
		url := ""
		if link, _ := card.QuerySelector("a[href]"); link != nil {
			url, _ = link.GetAttribute("href")
		}
		if url != "" && strings.HasPrefix(url, "/") {
			url = "https://hh.ru" + url
		}
		vacancy := textFrom(card, `[data-qa="negotiations-item-vacancy-name"]`)
		company := textFrom(card, `[data-qa="negotiations-item-company-name"]`)
		last := textFrom(card, `[data-qa="negotiations-item-last-message"]`)
		status := textFrom(card, `[data-qa="negotiations-item-state"]`)
		result = append(result, hh.Negotiation{
			ID:     url,
			Status: status,
			LastMessage: hh.Message{
				Text: strings.TrimSpace(last),
				From: strings.TrimSpace(company),
			},
			VacancyID: vacancy,
		})
	}
	_ = randomPause(tctx, 500*time.Millisecond, 2*time.Second)
	return result, nil
}

func (ac *AccountContext) GetMessages(ctx context.Context, negotiationURL string) ([]hh.Message, error) {
	tctx, cancel := withTimeout(ctx, defaultOpTimeout)
	defer cancel()

	if _, err := ac.Page.Goto(negotiationURL); err != nil {
		return nil, fmt.Errorf("open negotiation page: %w", err)
	}
	if _, err := waitForSelectorWithDebug(ac.Page, `[data-qa="chatik-message"]`, 20*time.Second); err != nil {
		return nil, err
	}

	nodes, err := ac.Page.QuerySelectorAll(`[data-qa="chatik-message"]`)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	messages := make([]hh.Message, 0, len(nodes))
	for idx, node := range nodes {
		text, _ := node.TextContent()
		from := textFrom(node, `[data-qa="chatik-message-author"]`)
		msg := hh.Message{ID: fmt.Sprintf("msg-%d", idx), Text: strings.TrimSpace(text), From: strings.TrimSpace(from)}

		buttons, _ := node.QuerySelectorAll(`button`)
		for bi, btn := range buttons {
			bt, _ := btn.TextContent()
			bt = strings.TrimSpace(bt)
			if bt == "" {
				continue
			}
			msg.Options = append(msg.Options, hh.MessageOption{ID: fmt.Sprintf("btn-%d", bi), Text: bt})
		}
		messages = append(messages, msg)
	}
	_ = randomPause(tctx, 500*time.Millisecond, 2*time.Second)
	return messages, nil
}

func (ac *AccountContext) SendMessage(ctx context.Context, negotiationURL string, text string) error {
	tctx, cancel := withTimeout(ctx, defaultOpTimeout)
	defer cancel()

	if _, err := ac.Page.Goto(negotiationURL); err != nil {
		return fmt.Errorf("open negotiation page: %w", err)
	}
	if _, err := waitForSelectorWithDebug(ac.Page, `[data-qa="chat-input"]`, 20*time.Second); err != nil {
		return err
	}
	if err := ac.typeLikeHuman(`[data-qa="chat-input"]`, text, 70); err != nil {
		return fmt.Errorf("type message: %w", err)
	}
	if err := ac.clickLikeHuman(tctx, `[data-qa="chat-send-button"]`); err != nil {
		return fmt.Errorf("click send: %w", err)
	}
	return randomPause(tctx, 500*time.Millisecond, 2*time.Second)
}

func (ac *AccountContext) ClickBotButton(ctx context.Context, negotiationURL string, optionText string) error {
	tctx, cancel := withTimeout(ctx, defaultOpTimeout)
	defer cancel()

	if _, err := ac.Page.Goto(negotiationURL); err != nil {
		return fmt.Errorf("open negotiation page: %w", err)
	}
	if _, err := waitForSelectorWithDebug(ac.Page, `button`, 20*time.Second); err != nil {
		return err
	}
	buttons, err := ac.Page.QuerySelectorAll("button")
	if err != nil {
		return fmt.Errorf("query bot buttons: %w", err)
	}
	for _, btn := range buttons {
		text, _ := btn.TextContent()
		if strings.EqualFold(strings.TrimSpace(text), strings.TrimSpace(optionText)) {
			_, _ = ac.Page.Evaluate("(el) => el.scrollIntoView({block: 'center'})", btn)
			if err := randomPause(tctx, 500*time.Millisecond, 2*time.Second); err != nil {
				return err
			}
			if err := btn.Click(); err != nil {
				return fmt.Errorf("click quick reply button: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("quick reply button %q not found", optionText)
}

func textFrom(parent playwright.ElementHandle, selector string) string {
	node, _ := parent.QuerySelector(selector)
	if node == nil {
		return ""
	}
	text, _ := node.TextContent()
	return strings.TrimSpace(text)
}
