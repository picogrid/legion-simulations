package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/picogrid/legion-simulations/pkg/models"
)

// GetOrganization gets the current organization details
func (c *Legion) GetOrganization(ctx context.Context) (*models.OrganizationResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v3/organizations", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	var org models.OrganizationResponse
	if err := decodeResponse(resp, &org); err != nil {
		return nil, fmt.Errorf("failed to decode organization response: %w", err)
	}

	return &org, nil
}

// GetOrganizationUsers gets the users in the organization
func (c *Legion) GetOrganizationUsers(ctx context.Context) (*models.OrganizationUserPaginatedResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v3/organizations/users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization users: %w", err)
	}

	var users models.OrganizationUserPaginatedResponse
	if err := decodeResponse(resp, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users response: %w", err)
	}

	return &users, nil
}
