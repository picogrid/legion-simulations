package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/picogrid/legion-simulations/pkg/logger"

	"github.com/picogrid/legion-simulations/pkg/models"
)

// CreateFeedDefinition creates a new feed definition
func (c *Legion) CreateFeedDefinition(ctx context.Context, req *models.CreateFeedDefinitionRequest) (*models.FeedDefinitionResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/definitions", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create feed definition: %w", err)
	}

	var feedDef models.FeedDefinitionResponse
	if err := decodeResponse(resp, &feedDef); err != nil {
		return nil, fmt.Errorf("failed to decode feed definition response: %w", err)
	}

	return &feedDef, nil
}

// GetFeedDefinition gets a feed definition by ID
func (c *Legion) GetFeedDefinition(ctx context.Context, feedID string) (*models.FeedDefinitionResponse, error) {
	path := fmt.Sprintf("/v3/feeds/definitions/%s", feedID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed definition: %w", err)
	}

	var feedDef models.FeedDefinitionResponse
	if err := decodeResponse(resp, &feedDef); err != nil {
		return nil, fmt.Errorf("failed to decode feed definition response: %w", err)
	}

	return &feedDef, nil
}

// UpdateFeedDefinition updates an existing feed definition
func (c *Legion) UpdateFeedDefinition(ctx context.Context, feedID string, req *models.UpdateFeedDefinitionRequest) (*models.FeedDefinitionResponse, error) {
	path := fmt.Sprintf("/v3/feeds/definitions/%s", feedID)
	resp, err := c.doRequest(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update feed definition: %w", err)
	}

	var feedDef models.FeedDefinitionResponse
	if err := decodeResponse(resp, &feedDef); err != nil {
		return nil, fmt.Errorf("failed to decode feed definition response: %w", err)
	}

	return &feedDef, nil
}

// DeleteFeedDefinition deletes a feed definition
func (c *Legion) DeleteFeedDefinition(ctx context.Context, feedID string) error {
	path := fmt.Sprintf("/v3/feeds/definitions/%s", feedID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete feed definition: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	return nil
}

// SearchFeedDefinitions searches for feed definitions based on criteria
func (c *Legion) SearchFeedDefinitions(ctx context.Context, req *models.FeedDefinitionSearchRequest) (*models.FeedDefinitionListResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/definitions/search", req)
	if err != nil {
		return nil, fmt.Errorf("failed to search feed definitions: %w", err)
	}

	var result models.FeedDefinitionListResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &result, nil
}

// GetFeedData gets feed data by feed ID
func (c *Legion) GetFeedData(ctx context.Context, feedID string) (*models.FeedDataResponse, error) {
	path := fmt.Sprintf("/v3/feeds/%s/data", feedID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed data: %w", err)
	}

	var feedData models.FeedDataResponse
	if err := decodeResponse(resp, &feedData); err != nil {
		return nil, fmt.Errorf("failed to decode feed data response: %w", err)
	}

	return &feedData, nil
}

// SearchFeedData searches for feed data based on criteria
func (c *Legion) SearchFeedData(ctx context.Context, req *models.FeedDataSearchRequest) (*models.FeedDataListResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/data/search", req)
	if err != nil {
		return nil, fmt.Errorf("failed to search feed data: %w", err)
	}

	var result models.FeedDataListResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &result, nil
}

// IngestServiceMessage ingests a message from a service
func (c *Legion) IngestServiceMessage(ctx context.Context, req *models.ServiceIngestMessageRequest) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/ingest", req)
	if err != nil {
		return fmt.Errorf("failed to ingest service message: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	return nil
}

// IngestFeedData ingests feed data using the standard ingestion endpoint
func (c *Legion) IngestFeedData(ctx context.Context, req *models.IngestFeedDataRequest) error {
	// Log the request for debugging
	logger.Debugf("Ingesting feed data - Entity: %s, FeedDef: %s", req.EntityID, req.FeedDefinitionID)

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/data", req)
	if err != nil {
		return fmt.Errorf("failed to ingest feed data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	return nil
}
