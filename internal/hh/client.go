package hh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
	HandleAuthFailure(status int)
}

type Client struct {
	baseURL     string
	httpClient  *http.Client
	tokenSource TokenSource
	logger      *slog.Logger
	rand        *rand.Rand
	randMu      sync.Mutex
}

func NewClient(httpClient *http.Client, tokenSource TokenSource, logger *slog.Logger) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Client{
		baseURL:     "https://api.hh.ru",
		httpClient:  httpClient,
		tokenSource: tokenSource,
		logger:      logger,
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *Client) DoJSON(ctx context.Context, method, path string, body any, out any) error {
	if err := c.randomDelay(ctx); err != nil {
		return fmt.Errorf("wait random delay: %w", err)
	}

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
	}

	parsedBase, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("parse base URL: %w", err)
	}
	resolvedURL, err := parsedBase.Parse(path)
	if err != nil {
		return fmt.Errorf("resolve request URL: %w", err)
	}
	requestURL := resolvedURL.String()
	backoff := time.Second
	for attempt := 0; attempt < 8; attempt++ {
		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, requestURL, reader)
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

		started := time.Now()
		resp, err := c.httpClient.Do(req)
		latency := time.Since(started)
		if err != nil {
			return fmt.Errorf("do request %s %s: %w", method, requestURL, err)
		}
		c.logger.Info("hh request", "method", method, "url", requestURL, "status", resp.StatusCode, "latency", latency.String())

		bodyBytes, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read response body: %w", readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close response body: %w", closeErr)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			wait := backoff
			if wait > 60*time.Second {
				wait = 60 * time.Second
			}
			if err := sleepContext(ctx, wait); err != nil {
				return fmt.Errorf("wait retry backoff: %w", err)
			}
			backoff *= 2
			continue
		}

		if c.tokenSource != nil && resp.StatusCode == http.StatusUnauthorized {
			c.tokenSource.HandleAuthFailure(resp.StatusCode)
		}

		if resp.StatusCode >= 300 {
			return fmt.Errorf("request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
		}

		if out != nil && len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, out); err != nil {
				return fmt.Errorf("unmarshal response body: %w", err)
			}
		}
		return nil
	}

	return fmt.Errorf("request failed after retries")
}

func (c *Client) randomDelay(ctx context.Context) error {
	c.randMu.Lock()
	seconds := c.rand.Intn(5) + 1
	c.randMu.Unlock()
	return sleepContext(ctx, time.Duration(seconds)*time.Second)
}

func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("context done: %w", ctx.Err())
	case <-t.C:
		return nil
	}
}
