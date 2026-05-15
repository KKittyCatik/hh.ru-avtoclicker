package hh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	payload, err := json.Marshal(map[string]any{})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}
	requestURL := c.baseURL + "/resumes/" + resumeID + "/publish"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.tokenSource != nil {
		tok, err := c.tokenSource.AccessToken(ctx)
		if err != nil {
			return fmt.Errorf("resolve access token: %w", err)
		}
		if tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("publish resume %s: %w", resumeID, err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode == http.StatusForbidden {
		var errResp struct {
			Errors []struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"errors"`
		}
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil {
			for _, e := range errResp.Errors {
				if e.Type == "resumes" {
					return fmt.Errorf("resume publish cooldown active: %s", e.Value)
				}
			}
		}
		return fmt.Errorf("publish resume %s: status=%d body=%s", resumeID, resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	if resp.StatusCode == http.StatusUnauthorized {
		if c.tokenSource != nil {
			c.tokenSource.HandleAuthFailure(resp.StatusCode)
		}
		return fmt.Errorf("publish resume %s: status=%d body=%s", resumeID, resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("publish resume %s: status=%d body=%s", resumeID, resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
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
