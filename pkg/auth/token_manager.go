package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TokenManager manages access and refresh tokens with automatic renewal
type TokenManager struct {
	keycloak      *KeycloakClient
	accessToken   string
	refreshToken  string
	expiresAt     time.Time
	refreshMargin time.Duration
	mu            sync.RWMutex
}

// NewTokenManager creates a new token manager
func NewTokenManager(keycloak *KeycloakClient, tokenResp *TokenResponse) *TokenManager {
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &TokenManager{
		keycloak:      keycloak,
		accessToken:   tokenResp.AccessToken,
		refreshToken:  tokenResp.RefreshToken,
		expiresAt:     expiresAt,
		refreshMargin: 30 * time.Second, // Refresh 30 seconds before expiry
	}
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (tm *TokenManager) GetAccessToken(ctx context.Context) (string, error) {
	tm.mu.RLock()

	if time.Now().Before(tm.expiresAt.Add(-tm.refreshMargin)) {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}

	tm.mu.RUnlock()

	return tm.refreshAccessToken(ctx)
}

// refreshAccessToken refreshes the access token
func (tm *TokenManager) refreshAccessToken(ctx context.Context) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if time.Now().Before(tm.expiresAt.Add(-tm.refreshMargin)) {
		return tm.accessToken, nil
	}

	tokenResp, err := tm.keycloak.RefreshToken(ctx, tm.refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	tm.accessToken = tokenResp.AccessToken
	tm.refreshToken = tokenResp.RefreshToken
	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return tm.accessToken, nil
}

// UpdateTokens updates the token manager with new authentication data
func (tm *TokenManager) UpdateTokens(tokenResp *TokenResponse) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.accessToken = tokenResp.AccessToken
	tm.refreshToken = tokenResp.RefreshToken
	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
}

// IsExpired checks if the current token is expired
func (tm *TokenManager) IsExpired() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return time.Now().After(tm.expiresAt)
}
