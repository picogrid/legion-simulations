package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GetAuthorizationURLFromLegion fetches the authorization URL from the Legion API
func GetAuthorizationURLFromLegion(ctx context.Context, legionURL string) (string, error) {
	_, err := url.Parse(legionURL)
	if err != nil {
		return "", fmt.Errorf("invalid Legion URL: %w", err)
	}

	transport := &http.Transport{}
	httpClient := &http.Client{Transport: transport}

	authURLEndpoint := fmt.Sprintf("%s/v3/integrations/oauth/authorization-url", legionURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURLEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get authorization URL: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Log close error but don't override the main error
			fmt.Printf("Warning: failed to close response body: %v\n", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get authorization URL: status %d", resp.StatusCode)
	}

	var authResp struct {
		AuthorizationURL string `json:"authorization_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if authResp.AuthorizationURL == "" {
		return "", fmt.Errorf("empty authorization URL in response")
	}

	return authResp.AuthorizationURL, nil
}

// OAuthServerMetadata represents the OAuth 2.0 Authorization Server Metadata
// as defined in RFC 8414
type OAuthServerMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint,omitempty"`
	JwksURI                           string   `json:"jwks_uri"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                   []string `json:"scopes_supported,omitempty"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported,omitempty"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`
}

// GetOAuthMetadataFromLegion fetches the OAuth 2.0 server metadata from Legion's well-known endpoint
func GetOAuthMetadataFromLegion(ctx context.Context, legionURL string) (*OAuthServerMetadata, error) {
	_, err := url.Parse(legionURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Legion URL: %w", err)
	}

	transport := &http.Transport{}
	httpClient := &http.Client{Transport: transport}

	wellKnownEndpoint := fmt.Sprintf("%s/v3/.well-known/oauth-authorization-server", legionURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnownEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth metadata: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Log close error but don't override the main error
			fmt.Printf("Warning: failed to close response body: %v\n", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get OAuth metadata: status %d", resp.StatusCode)
	}

	var metadata OAuthServerMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode OAuth metadata: %w", err)
	}

	return &metadata, nil
}

// GetAuthConfigFromLegion creates an AuthConfig by fetching the auth URL from Legion
// It first tries to get OAuth metadata from the well-known endpoint, then falls back
// to the authorization URL endpoint if that fails
func GetAuthConfigFromLegion(ctx context.Context, legionURL string) (AuthConfig, error) {
	// First, try to get OAuth metadata from the well-known endpoint
	metadata, metadataErr := GetOAuthMetadataFromLegion(ctx, legionURL)
	if metadataErr == nil && metadata.AuthorizationEndpoint != "" {
		// Successfully got metadata, extract config from authorization endpoint
		u, err := url.Parse(metadata.AuthorizationEndpoint)
		if err != nil {
			return AuthConfig{}, fmt.Errorf("invalid authorization endpoint in metadata: %w", err)
		}

		pathParts := strings.Split(u.Path, "/")
		var keycloakURL string
		var realm string

		for i, part := range pathParts {
			if part == "realms" && i+1 < len(pathParts) {
				realm = pathParts[i+1]
				keycloakURL = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
				if strings.Contains(u.Path, "/auth/realms") {
					keycloakURL += "/auth"
				}
				break
			}
		}

		if realm != "" && keycloakURL != "" {
			return AuthConfig{
				KeycloakURL: keycloakURL,
				Realm:       realm,
				ClientID:    "frontend...orion", // TODO: Need a better client
			}, nil
		}
	}

	// Fall back to the authorization URL endpoint
	authURL, err := GetAuthorizationURLFromLegion(ctx, legionURL)
	if err != nil {
		return AuthConfig{}, fmt.Errorf("failed to get authorization URL from Legion: %w", err)
	}

	u, err := url.Parse(authURL)
	if err != nil {
		return AuthConfig{}, fmt.Errorf("invalid authorization URL: %w", err)
	}

	// The auth URL format is typically:
	// https://auth.legion.com/auth/realms/legion/protocol/openid-connect/auth
	// We need to extract the base URL and realm

	pathParts := strings.Split(u.Path, "/")
	var keycloakURL string
	var realm string

	for i, part := range pathParts {
		if part == "realms" && i+1 < len(pathParts) {
			realm = pathParts[i+1]
			keycloakURL = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			if strings.Contains(u.Path, "/auth/realms") {
				keycloakURL += "/auth"
			}
			break
		}
	}

	if realm == "" {
		return AuthConfig{}, fmt.Errorf("could not extract realm from authorization URL")
	}

	return AuthConfig{
		KeycloakURL: keycloakURL,
		Realm:       realm,
		ClientID:    "frontend...orion", // This is still hardcoded as it's client-specific
	}, nil
}

// AuthenticateUserWithLegion authenticates a user using the auth URL from Legion
func AuthenticateUserWithLegion(ctx context.Context, legionURL string) (*TokenManager, error) {
	config, err := GetAuthConfigFromLegion(ctx, legionURL)
	if err != nil {
		fmt.Println("⚠️  Could not fetch auth config from Legion, using defaults")
		config = DefaultAuthConfig()
	}

	// Fix localhost port issue: if Legion URL is localhost and Keycloak URL uses port 8080, change to 8443
	if strings.Contains(legionURL, "localhost") && strings.Contains(config.KeycloakURL, "localhost:8080") {
		config.KeycloakURL = strings.Replace(config.KeycloakURL, "localhost:8080", "localhost:8443", 1)
		fmt.Println("⚠️  Adjusted Keycloak URL for localhost: using port 8443 instead of 8080")
	}

	return AuthenticateUser(ctx, config)
}
