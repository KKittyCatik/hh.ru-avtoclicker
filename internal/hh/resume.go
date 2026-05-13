package hh

import (
	"context"
	"fmt"
	"time"
)

func (c *Client) ListResumes(ctx context.Context) ([]Resume, error) {
	var response struct {
		Items []Resume `json:"items"`
	}
	if err := c.DoJSON(ctx, "GET", "/resumes/mine", nil, &response); err != nil {
		return nil, fmt.Errorf("list resumes: %w", err)
	}
	return response.Items, nil
}

func (c *Client) PublishResume(ctx context.Context, resumeID string) error {
	if err := c.DoJSON(ctx, "POST", "/resumes/"+resumeID+"/publish", map[string]any{}, nil); err != nil {
		return fmt.Errorf("publish resume %s: %w", resumeID, err)
	}
	return nil
}

func (c *Client) PublishReadyResumes(ctx context.Context, resumes []Resume) ([]string, error) {
	now := time.Now()
	published := make([]string, 0)
	for _, r := range resumes {
		if !r.NextPublishAt.IsZero() && r.NextPublishAt.After(now) {
			continue
		}
		if err := c.PublishResume(ctx, r.ID); err != nil {
			return nil, fmt.Errorf("publish ready resume %s: %w", r.ID, err)
		}
		published = append(published, r.ID)
	}
	return published, nil
}
