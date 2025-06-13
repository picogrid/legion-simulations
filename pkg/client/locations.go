package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/picogrid/legion-simulations/pkg/models"
)

// CreateEntityLocation creates a new location for an entity
func (c *Legion) CreateEntityLocation(ctx context.Context, entityID string, req *models.CreateEntityLocationRequest) (*models.EntityLocationResponse, error) {
	path := fmt.Sprintf("/v3/entities/%s/locations", entityID)
	resp, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity location: %w", err)
	}

	var location models.EntityLocationResponse
	if err := decodeResponse(resp, &location); err != nil {
		return nil, fmt.Errorf("failed to decode location response: %w", err)
	}

	return &location, nil
}

// GetEntityLocation gets a specific location for an entity
func (c *Legion) GetEntityLocation(ctx context.Context, entityID, locationID string) (*models.EntityLocationResponse, error) {
	path := fmt.Sprintf("/v3/entities/%s/locations/%s", entityID, locationID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity location: %w", err)
	}

	var location models.EntityLocationResponse
	if err := decodeResponse(resp, &location); err != nil {
		return nil, fmt.Errorf("failed to decode location response: %w", err)
	}

	return &location, nil
}

// GetEntityLocations gets all locations for an entity
func (c *Legion) GetEntityLocations(ctx context.Context, entityID string) (*models.EntityLocationPaginatedResponse, error) {
	path := fmt.Sprintf("/v3/entities/%s/locations", entityID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity locations: %w", err)
	}

	var result models.EntityLocationPaginatedResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode locations response: %w", err)
	}

	return &result, nil
}

// SearchEntityLocations searches for entity locations based on criteria
func (c *Legion) SearchEntityLocations(ctx context.Context, req *models.SearchEntityLocationsRequest) (*models.EntityLocationPaginatedResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/entities/locations/search", req)
	if err != nil {
		return nil, fmt.Errorf("failed to search entity locations: %w", err)
	}

	var result models.EntityLocationPaginatedResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &result, nil
}
