package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/picogrid/legion-simulations/pkg/logger"

	"github.com/picogrid/legion-simulations/pkg/models"
)

// GetMe retrieves the current user information
func (c *Legion) GetMe(ctx context.Context) (*models.UserResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v3/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	var user models.UserResponse
	if err := decodeResponse(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

// GetMyOrganizations gets the organizations the current user belongs to
func (c *Legion) GetMyOrganizations(ctx context.Context) (*models.OrganizationResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v3/me/orgs", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	var orgs models.OrganizationResponse
	if err := decodeResponse(resp, &orgs); err != nil {
		return nil, fmt.Errorf("failed to decode organizations response: %w", err)
	}

	return &orgs, nil
}

// ValidateConnection tests the connection to Legion by calling the /v3/me endpoint
func (c *Legion) ValidateConnection(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v3/me", nil)
	if err != nil {
		return fmt.Errorf("connection validation failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	return nil
}
