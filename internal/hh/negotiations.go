package hh

import (
	"context"
	"fmt"
)

type ApplyRequest struct {
	ResumeID    string `json:"resume_id"`
	VacancyID   string `json:"vacancy_id"`
	CoverLetter string `json:"cover_letter"`
}

type SendReplyRequest struct {
	Text               string `json:"text,omitempty"`
	QuickReplyOptionID string `json:"quick_reply_option_id,omitempty"`
}

func (c *Client) ApplyToVacancy(ctx context.Context, req ApplyRequest) error {
	if err := c.DoJSON(ctx, "POST", "/negotiations", req, nil); err != nil {
		return fmt.Errorf("apply to vacancy %s: %w", req.VacancyID, err)
	}
	return nil
}

func (c *Client) ListNegotiations(ctx context.Context) ([]Negotiation, error) {
	var response struct {
		Items []Negotiation `json:"items"`
	}
	if err := c.DoJSON(ctx, "GET", "/negotiations", nil, &response); err != nil {
		return nil, fmt.Errorf("list negotiations: %w", err)
	}
	return response.Items, nil
}

func (c *Client) ListNegotiationMessages(ctx context.Context, negotiationID string) ([]Message, error) {
	var response struct {
		Items []Message `json:"items"`
	}
	if err := c.DoJSON(ctx, "GET", "/negotiations/"+negotiationID+"/messages", nil, &response); err != nil {
		return nil, fmt.Errorf("list negotiation messages %s: %w", negotiationID, err)
	}
	return response.Items, nil
}

func (c *Client) SendNegotiationReply(ctx context.Context, negotiationID string, req SendReplyRequest) error {
	if err := c.DoJSON(ctx, "POST", "/negotiations/"+negotiationID+"/messages", req, nil); err != nil {
		return fmt.Errorf("send negotiation reply %s: %w", negotiationID, err)
	}
	return nil
}
