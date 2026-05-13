package hh

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type AuthManager struct {
	oauthConfig *oauth2.Config
	httpClient  *http.Client
	logger      *slog.Logger

	mu          sync.RWMutex
	token       OAuthToken
	needsReauth *bool
}

func NewAuthManager(clientID, clientSecret, redirectURL string, scopes []string, logger *slog.Logger, needsReauth *bool) *AuthManager {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthManager{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://hh.ru/oauth/authorize",
				TokenURL: "https://hh.ru/oauth/token",
			},
			RedirectURL: redirectURL,
			Scopes:      scopes,
		},
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		logger:      logger,
		needsReauth: needsReauth,
	}
}

func (a *AuthManager) AuthCodeURL(state string) string {
	return a.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (a *AuthManager) ExchangeCode(ctx context.Context, code string) (OAuthToken, error) {
	tok, err := a.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return OAuthToken{}, fmt.Errorf("exchange auth code: %w", err)
	}
	converted := fromOAuthToken(tok)
	a.mu.Lock()
	a.token = converted
	a.mu.Unlock()
	if a.needsReauth != nil {
		*a.needsReauth = false
	}
	return converted, nil
}

func (a *AuthManager) SetToken(token OAuthToken) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.token = token
}

func (a *AuthManager) AccessToken(ctx context.Context) (string, error) {
	a.mu.RLock()
	current := a.token
	a.mu.RUnlock()

	if current.AccessToken == "" {
		return "", nil
	}
	if time.Until(current.ExpiresAt) > 5*time.Minute {
		return current.AccessToken, nil
	}

	refreshed, err := a.refresh(ctx, current)
	if err != nil {
		return "", fmt.Errorf("refresh token: %w", err)
	}

	a.mu.Lock()
	a.token = refreshed
	a.mu.Unlock()
	if a.needsReauth != nil {
		*a.needsReauth = false
	}
	return refreshed.AccessToken, nil
}

func (a *AuthManager) refresh(ctx context.Context, token OAuthToken) (OAuthToken, error) {
	src := a.oauthConfig.TokenSource(ctx, token.toOAuthToken())
	newTok, err := src.Token()
	if err != nil {
		return OAuthToken{}, fmt.Errorf("request refreshed token: %w", err)
	}
	a.logger.Info("oauth token refreshed")
	return fromOAuthToken(newTok), nil
}

func (a *AuthManager) HandleAuthFailure(status int) {
	if status != http.StatusUnauthorized && status != http.StatusForbidden {
		return
	}
	if a.needsReauth != nil {
		*a.needsReauth = true
	}
	a.logger.Error("account needs reauthorization", "status", status)
}

func fromOAuthToken(tok *oauth2.Token) OAuthToken {
	if tok == nil {
		return OAuthToken{}
	}
	return OAuthToken{AccessToken: tok.AccessToken, RefreshToken: tok.RefreshToken, ExpiresAt: tok.Expiry}
}

func (o OAuthToken) toOAuthToken() *oauth2.Token {
	return &oauth2.Token{AccessToken: o.AccessToken, RefreshToken: o.RefreshToken, Expiry: o.ExpiresAt}
}
