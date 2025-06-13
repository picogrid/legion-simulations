package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/picogrid/legion-simulations/pkg/logger"

	"github.com/picogrid/legion-simulations/pkg/models"
)

// CreateEntity creates a new entity in Legion
func (c *Legion) CreateEntity(ctx context.Context, req *models.CreateEntityRequest) (*models.EntityResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/entities", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	var entity models.EntityResponse
	if err := decodeResponse(resp, &entity); err != nil {
		return nil, fmt.Errorf("failed to decode entity response: %w", err)
	}

	return &entity, nil
}

// GetEntity retrieves an entity by ID
func (c *Legion) GetEntity(ctx context.Context, entityID string) (*models.EntityResponse, error) {
	path := fmt.Sprintf("/v3/entities/%s", entityID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	var entity models.EntityResponse
	if err := decodeResponse(resp, &entity); err != nil {
		return nil, fmt.Errorf("failed to decode entity response: %w", err)
	}

	return &entity, nil
}

// UpdateEntity updates an existing entity
func (c *Legion) UpdateEntity(ctx context.Context, entityID string, req *models.UpdateEntityRequest) (*models.EntityResponse, error) {
	path := fmt.Sprintf("/v3/entities/%s", entityID)
	resp, err := c.doRequest(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	var entity models.EntityResponse
	if err := decodeResponse(resp, &entity); err != nil {
		return nil, fmt.Errorf("failed to decode entity response: %w", err)
	}

	return &entity, nil
}

// DeleteEntity deletes an entity by ID
func (c *Legion) DeleteEntity(ctx context.Context, entityID string) error {
	path := fmt.Sprintf("/v3/entities/%s", entityID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	return nil
}

// SearchEntities searches for entities based on the provided criteria
func (c *Legion) SearchEntities(ctx context.Context, req *models.SearchEntitiesRequest) (*models.EntityPaginatedResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/entities/search", req)
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}

	var result models.EntityPaginatedResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &result, nil
}
