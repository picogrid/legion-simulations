package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
)

// CreateFeedDefinition creates a new feed definition
func (c *Legion) CreateFeedDefinition(ctx context.Context, req *models.CreateFeedDefinitionRequest) (*models.FeedDefinitionResponse, error) {
	body, err := toCreateFeedDefinitionRequest(req)
	if err != nil {
		return nil, fmt.Errorf("build create feed definition request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/definitions", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create feed definition: %w", err)
	}

	var raw models.PostV3FeedsDefinitions201Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode feed definition response: %w", err)
	}

	return fromFeedDefinitionCreated(raw)
}

// GetFeedDefinition gets a feed definition by ID
func (c *Legion) GetFeedDefinition(ctx context.Context, feedID string) (*models.FeedDefinitionResponse, error) {
	path := fmt.Sprintf("/v3/feeds/definitions/%s", feedID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed definition: %w", err)
	}

	var raw models.GetV3FeedsDefinitionsbyFeedDefinitionId200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode feed definition response: %w", err)
	}

	return fromFeedDefinitionFetched(raw)
}

// UpdateFeedDefinition updates an existing feed definition
func (c *Legion) UpdateFeedDefinition(ctx context.Context, feedID string, req *models.UpdateFeedDefinitionRequest) (*models.FeedDefinitionResponse, error) {
	body, err := toUpdateFeedDefinitionRequest(req)
	if err != nil {
		return nil, fmt.Errorf("build update feed definition request: %w", err)
	}

	path := fmt.Sprintf("/v3/feeds/definitions/%s", feedID)
	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update feed definition: %w", err)
	}

	var raw models.PutV3FeedsDefinitionsbyFeedDefinitionId200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode feed definition response: %w", err)
	}

	return fromFeedDefinitionUpdated(raw)
}

// DeleteFeedDefinition deletes a feed definition
func (c *Legion) DeleteFeedDefinition(ctx context.Context, feedID string) error {
	path := fmt.Sprintf("/v3/feeds/definitions/%s", feedID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete feed definition: %w", err)
	}
	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			logger.Errorf("failed to close response body: %v", closeErr)
		}
	}(resp.Body)

	return nil
}

// SearchFeedDefinitions searches for feed definitions based on criteria
func (c *Legion) SearchFeedDefinitions(ctx context.Context, req *models.FeedDefinitionSearchRequest) (*models.FeedDefinitionListResponse, error) {
	body := &models.PostV3FeedsDefinitionsSearchRequest{}
	if req != nil {
		if req.Category != "" {
			category := models.PostV3FeedsDefinitionsSearchRequestCategory(req.Category)
			body.Category = &category
		}
		body.CreatedAfter = formatOptionalTime(req.CreatedAfter)
		body.CreatedBefore = formatOptionalTime(req.CreatedBefore)
		body.DataType = req.DataType
		if req.EntityID != uuid.Nil {
			entityID := toOpenapiUUID(req.EntityID)
			body.EntityId = &entityID
		}
		body.FeedName = req.FeedName
		body.IntegrationId = toOpenapiUUIDPtr(req.IntegrationID)
		body.IsActive = req.IsActive
		body.IsTemplate = req.IsTemplate
		body.OrganizationId = toOpenapiUUIDPtr(req.OrganizationID)
		body.TemplateId = toOpenapiUUIDPtr(req.TemplateID)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/definitions/search", body)
	if err != nil {
		return nil, fmt.Errorf("failed to search feed definitions: %w", err)
	}

	var raw models.PostV3FeedsDefinitionsSearch200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	results := make([]models.FeedDefinitionResponse, 0, len(raw.Results))
	for _, feed := range raw.Results {
		converted, err := feedDefinitionFromFields(
			feed.Id,
			feed.OrganizationId,
			string(feed.Category),
			feed.DataType,
			feed.Description,
			feed.EntityId,
			feed.FeedName,
			feed.IntegrationId,
			feed.IsActive,
			feed.IsTemplate,
			feed.Metadata,
			feed.SchemaDefinition,
			feed.TemplateId,
			feed.CreatedAt,
			feed.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, *converted)
	}

	result := models.FeedDefinitionListResponse{
		Results:    results,
		TotalCount: int(raw.TotalCount),
		Paging:     toPaging(raw.Paging.Next, raw.Paging.Previous),
	}

	return &result, nil
}

// GetFeedData gets feed data by feed ID
func (c *Legion) GetFeedData(ctx context.Context, feedID string) (*models.FeedDataResponse, error) {
	path := fmt.Sprintf("/v3/feeds/data/%s", feedID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed data: %w", err)
	}

	var raw models.GetV3FeedsDatabyFeedId200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode feed data response: %w", err)
	}

	return fromFeedDataResponse(raw)
}

// SearchFeedData searches for feed data based on criteria
func (c *Legion) SearchFeedData(ctx context.Context, req *models.FeedDataSearchRequest) (*models.FeedDataListResponse, error) {
	body := &models.PostV3FeedsSearchRequest{}
	if req != nil {
		if req.StartTime != nil {
			body.RecordedAfter = formatOptionalTime(req.StartTime)
		}
		if req.EndTime != nil {
			body.RecordedBefore = formatOptionalTime(req.EndTime)
		}
		if req.EntityID != nil {
			entityIDs := []openapi_types.UUID{toOpenapiUUID(*req.EntityID)}
			body.EntityIds = &entityIDs
		}
		if req.FeedID != uuid.Nil {
			feedIDs := []openapi_types.UUID{toOpenapiUUID(req.FeedID)}
			body.FeedDefinitionIds = &feedIDs
		}
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/search", body)
	if err != nil {
		return nil, fmt.Errorf("failed to search feed data: %w", err)
	}

	var raw models.PostV3FeedsSearch200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	results := make([]models.FeedDataResponse, 0, len(raw.Results))
	for _, feedData := range raw.Results {
		converted, err := feedDataFromFields(
			feedData.Id,
			feedData.OrganizationId,
			feedData.EntityId,
			feedData.FeedDefinitionId,
			feedData.Payload,
			feedData.BlobContentType,
			feedData.BlobKey,
			feedData.BlobMetadata,
			feedData.BlobSizeBytes,
			feedData.BlobStorageType,
			feedData.RecordedAt,
			feedData.ReceivedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, *converted)
	}

	result := models.FeedDataListResponse{
		Results:    results,
		TotalCount: int(raw.TotalCount),
		Paging:     toPaging(raw.Paging.Next, raw.Paging.Previous),
	}

	return &result, nil
}

// IngestServiceMessage ingests a message from a service
func (c *Legion) IngestServiceMessage(ctx context.Context, req *models.ServiceIngestMessageRequest) error {
	body, err := toFeedMessageRequest(&models.IngestFeedDataRequest{
		EntityID:         req.EntityID,
		FeedDefinitionID: req.FeedDefinitionID,
		Metadata:         req.Metadata,
		Payload:          req.Payload,
		RecordedAt:       req.RecordedAt,
	})
	if err != nil {
		return fmt.Errorf("build service message request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/messages", body)
	if err != nil {
		return fmt.Errorf("failed to ingest service message: %w", err)
	}
	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			logger.Errorf("failed to close response body: %v", closeErr)
		}
	}(resp.Body)

	return nil
}

// IngestFeedData ingests feed data using the standard ingestion endpoint
func (c *Legion) IngestFeedData(ctx context.Context, req *models.IngestFeedDataRequest) error {
	body, err := toFeedMessageRequest(req)
	if err != nil {
		return fmt.Errorf("build ingest feed data request: %w", err)
	}

	logger.Debugf("Ingesting feed data - Entity: %s, FeedDef: %s", body.EntityId, body.FeedDefinitionId)

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/feeds/messages", body)
	if err != nil {
		return fmt.Errorf("failed to ingest feed data: %w", err)
	}
	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			logger.Errorf("failed to close response body: %v", closeErr)
		}
	}(resp.Body)

	return nil
}

func toCreateFeedDefinitionRequest(req *models.CreateFeedDefinitionRequest) (*models.PostV3FeedsDefinitionsRequest, error) {
	if req == nil || req.Category == nil || req.DataType == nil || req.FeedName == nil || req.IsActive == nil {
		return nil, fmt.Errorf("create feed definition request is missing required fields")
	}

	metadata, err := rawMessageToMap(req.Metadata)
	if err != nil {
		return nil, err
	}
	schemaDefinition, err := rawMessageToMap(req.SchemaDefinition)
	if err != nil {
		return nil, err
	}

	body := &models.PostV3FeedsDefinitionsRequest{
		Category:         models.PostV3FeedsDefinitionsRequestCategory(*req.Category),
		DataType:         *req.DataType,
		Description:      optionalString(req.Description),
		EntityId:         toOpenapiUUIDPtr(uuidPtrIfNotNil(req.EntityID)),
		FeedName:         *req.FeedName,
		IntegrationId:    toOpenapiUUIDPtr(req.IntegrationID),
		IsActive:         *req.IsActive,
		IsTemplate:       boolValue(req.IsTemplate),
		Metadata:         metadata,
		SchemaDefinition: schemaDefinition,
		TemplateId:       toOpenapiUUIDPtr(req.TemplateID),
	}

	return body, nil
}

func toUpdateFeedDefinitionRequest(req *models.UpdateFeedDefinitionRequest) (*models.PutV3FeedsDefinitionsbyFeedDefinitionIdRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("update feed definition request is nil")
	}
	if req.DataType == nil || req.FeedName == nil || req.IsActive == nil {
		return nil, fmt.Errorf("update feed definition request is missing required fields")
	}

	metadata, err := rawMessageToMap(req.Metadata)
	if err != nil {
		return nil, err
	}
	schemaDefinition, err := rawMessageToMap(req.SchemaDefinition)
	if err != nil {
		return nil, err
	}

	body := &models.PutV3FeedsDefinitionsbyFeedDefinitionIdRequest{
		DataType:         *req.DataType,
		Description:      req.Description,
		EntityId:         toOpenapiUUIDPtr(req.EntityID),
		FeedName:         *req.FeedName,
		IntegrationId:    toOpenapiUUIDPtr(req.IntegrationID),
		IsActive:         *req.IsActive,
		IsTemplate:       boolValue(req.IsTemplate),
		Metadata:         metadata,
		SchemaDefinition: schemaDefinition,
		TemplateId:       toOpenapiUUIDPtr(req.TemplateID),
	}
	body.Category = models.PutV3FeedsDefinitionsbyFeedDefinitionIdRequestCategory(req.Category)

	return body, nil
}

func toFeedMessageRequest(req *models.IngestFeedDataRequest) (*models.PostV3FeedsMessagesRequest, error) {
	if req == nil || req.EntityID == nil || req.FeedDefinitionID == nil || req.Payload == nil || req.RecordedAt == nil {
		return nil, fmt.Errorf("feed message request is missing required fields")
	}

	payload, err := rawMessageToMapValue(req.Payload)
	if err != nil {
		return nil, err
	}
	metadata, err := rawMessageToMap(req.Metadata)
	if err != nil {
		return nil, err
	}

	return &models.PostV3FeedsMessagesRequest{
		EntityId:         toOpenapiUUID(*req.EntityID),
		FeedDefinitionId: toOpenapiUUID(*req.FeedDefinitionID),
		Metadata:         metadata,
		Payload:          payload,
		RecordedAt:       req.RecordedAt.UTC().Format(time.RFC3339),
	}, nil
}

func fromFeedDefinitionCreated(raw models.PostV3FeedsDefinitions201Response) (*models.FeedDefinitionResponse, error) {
	return feedDefinitionFromFields(raw.Id, raw.OrganizationId, string(raw.Category), raw.DataType, raw.Description, raw.EntityId, raw.FeedName, raw.IntegrationId, raw.IsActive, raw.IsTemplate, raw.Metadata, raw.SchemaDefinition, raw.TemplateId, raw.CreatedAt, raw.UpdatedAt)
}

func fromFeedDefinitionFetched(raw models.GetV3FeedsDefinitionsbyFeedDefinitionId200Response) (*models.FeedDefinitionResponse, error) {
	return feedDefinitionFromFields(raw.Id, raw.OrganizationId, string(raw.Category), raw.DataType, raw.Description, raw.EntityId, raw.FeedName, raw.IntegrationId, raw.IsActive, raw.IsTemplate, raw.Metadata, raw.SchemaDefinition, raw.TemplateId, raw.CreatedAt, raw.UpdatedAt)
}

func fromFeedDefinitionUpdated(raw models.PutV3FeedsDefinitionsbyFeedDefinitionId200Response) (*models.FeedDefinitionResponse, error) {
	return feedDefinitionFromFields(raw.Id, raw.OrganizationId, string(raw.Category), raw.DataType, raw.Description, raw.EntityId, raw.FeedName, raw.IntegrationId, raw.IsActive, raw.IsTemplate, raw.Metadata, raw.SchemaDefinition, raw.TemplateId, raw.CreatedAt, raw.UpdatedAt)
}

func feedDefinitionFromFields(
	id openapi_types.UUID,
	organizationID openapi_types.UUID,
	category string,
	dataType string,
	description *string,
	entityID *openapi_types.UUID,
	feedName string,
	integrationID *openapi_types.UUID,
	isActive bool,
	isTemplate bool,
	metadata *map[string]interface{},
	schemaDefinition *map[string]interface{},
	templateID *openapi_types.UUID,
	createdAt string,
	updatedAt string,
) (*models.FeedDefinitionResponse, error) {
	createdTime, err := parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	updatedTime, err := parseTime(updatedAt)
	if err != nil {
		return nil, err
	}
	metadataRaw, err := mapToRawMessage(metadata)
	if err != nil {
		return nil, err
	}
	schemaRaw, err := mapToRawMessage(schemaDefinition)
	if err != nil {
		return nil, err
	}

	response := &models.FeedDefinitionResponse{
		ID:               fromOpenapiUUID(id),
		OrganizationID:   fromOpenapiUUID(organizationID),
		Category:         models.MessageCategory(category),
		DataType:         dataType,
		Description:      description,
		FeedName:         feedName,
		IntegrationID:    uuidPtrFromOpenapi(integrationID),
		IsActive:         isActive,
		IsTemplate:       isTemplate,
		Metadata:         metadataRaw,
		SchemaDefinition: schemaRaw,
		TemplateID:       uuidPtrFromOpenapi(templateID),
		CreatedAt:        createdTime,
		UpdatedAt:        updatedTime,
	}
	if entityID != nil {
		response.EntityID = fromOpenapiUUID(*entityID)
	}

	return response, nil
}

func fromFeedDataResponse(raw models.GetV3FeedsDatabyFeedId200Response) (*models.FeedDataResponse, error) {
	return feedDataFromFields(raw.Id, raw.OrganizationId, raw.EntityId, raw.FeedDefinitionId, raw.Payload, raw.BlobContentType, raw.BlobKey, raw.BlobMetadata, raw.BlobSizeBytes, raw.BlobStorageType, raw.RecordedAt, raw.ReceivedAt)
}

func feedDataFromFields(
	id openapi_types.UUID,
	organizationID openapi_types.UUID,
	entityID openapi_types.UUID,
	feedDefinitionID openapi_types.UUID,
	payload *map[string]interface{},
	blobContentType *string,
	blobKey *string,
	blobMetadata *map[string]interface{},
	blobSizeBytes *float32,
	blobStorageType *string,
	recordedAt string,
	receivedAt string,
) (*models.FeedDataResponse, error) {
	recordedTime, err := parseTime(recordedAt)
	if err != nil {
		return nil, err
	}
	receivedTime, err := parseTime(receivedAt)
	if err != nil {
		return nil, err
	}
	payloadRaw, err := mapToRawMessage(payload)
	if err != nil {
		return nil, err
	}
	metadataRaw, err := mapToRawMessage(blobMetadata)
	if err != nil {
		return nil, err
	}

	return &models.FeedDataResponse{
		ID:               fromOpenapiUUID(id),
		OrganizationID:   fromOpenapiUUID(organizationID),
		EntityID:         fromOpenapiUUID(entityID),
		FeedDefinitionID: fromOpenapiUUID(feedDefinitionID),
		Payload:          payloadRaw,
		BlobContentType:  blobContentType,
		BlobKey:          blobKey,
		BlobMetadata:     metadataRaw,
		BlobSizeBytes:    blobSizeBytes,
		BlobStorageType:  blobStorageType,
		RecordedAt:       recordedTime,
		ReceivedAt:       receivedTime,
	}, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func uuidPtrIfNotNil(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}

	return &id
}
