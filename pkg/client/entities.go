package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
)

// CreateEntity creates a new entity in Legion
func (c *Legion) CreateEntity(ctx context.Context, req *models.CreateEntityRequest) (*models.EntityResponse, error) {
	body, err := toCreateEntityRequest(req)
	if err != nil {
		return nil, fmt.Errorf("build create entity request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/entities", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var raw models.PostV3Entities200Response
		if err := decodeResponse(resp, &raw); err != nil {
			return nil, fmt.Errorf("failed to decode entity response: %w", err)
		}
		return fromEntityResponse200(raw)
	case http.StatusCreated:
		var raw models.PostV3Entities201Response
		if err := decodeResponse(resp, &raw); err != nil {
			return nil, fmt.Errorf("failed to decode entity response: %w", err)
		}
		return fromEntityResponse201(raw)
	default:
		return nil, fmt.Errorf("unexpected create entity status: %d", resp.StatusCode)
	}
}

// GetEntity retrieves an entity by ID
func (c *Legion) GetEntity(ctx context.Context, entityID string) (*models.EntityResponse, error) {
	path := fmt.Sprintf("/v3/entities/%s", entityID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	var raw models.GetV3EntitiesbyEntityId200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode entity response: %w", err)
	}

	return fromGetEntityResponse(raw)
}

// UpdateEntity updates an existing entity
func (c *Legion) UpdateEntity(ctx context.Context, entityID string, req *models.UpdateEntityRequest) (*models.EntityResponse, error) {
	body, err := toUpdateEntityRequest(req)
	if err != nil {
		return nil, fmt.Errorf("build update entity request: %w", err)
	}

	path := fmt.Sprintf("/v3/entities/%s", entityID)
	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	var raw models.PutV3EntitiesbyEntityId200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode entity response: %w", err)
	}

	return fromPutEntityResponse(raw)
}

// DeleteEntity deletes an entity by ID
func (c *Legion) DeleteEntity(ctx context.Context, entityID string) error {
	path := fmt.Sprintf("/v3/entities/%s", entityID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			logger.Errorf("failed to close response body: %v", closeErr)
		}
	}(resp.Body)

	return nil
}

// SearchEntities searches for entities based on the provided criteria
func (c *Legion) SearchEntities(ctx context.Context, req *models.SearchEntitiesRequest) (*models.EntityPaginatedResponse, error) {
	body, err := toSearchEntitiesRequest(req)
	if err != nil {
		return nil, fmt.Errorf("build search entities request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/entities/search", body)
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}

	var raw models.PostV3EntitiesSearch200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	results := make([]models.EntityResponse, 0, len(raw.Results))
	for _, entity := range raw.Results {
		metadata, err := mapToRawMessage(entity.Metadata)
		if err != nil {
			return nil, fmt.Errorf("decode entity metadata: %w", err)
		}
		classification, err := mapToRawMessage(entity.Classification)
		if err != nil {
			return nil, fmt.Errorf("decode entity classification: %w", err)
		}
		createdAt, err := parseTime(entity.CreatedAt)
		if err != nil {
			return nil, err
		}
		updatedAt, err := parseTime(entity.UpdatedAt)
		if err != nil {
			return nil, err
		}
		deletedAt, err := parseOptionalTime(entity.DeletedAt)
		if err != nil {
			return nil, err
		}

		results = append(results, models.EntityResponse{
			ID:                           fromOpenapiUUID(entity.Id),
			OrganizationID:               fromOpenapiUUID(entity.OrganizationId),
			Name:                         entity.Name,
			Category:                     models.Category(entity.Category),
			Type:                         entity.Type,
			Status:                       entity.Status,
			Affiliation:                  models.Affiliation(entity.Affiliation),
			ParentID:                     uuidPtrFromOpenapi(entity.ParentId),
			Metadata:                     metadata,
			Classification:               classification,
			TopClassification:            entity.TopClassification,
			TopClassificationProbability: entity.TopClassificationProbability,
			CreatedAt:                    createdAt,
			UpdatedAt:                    updatedAt,
			DeletedAt:                    deletedAt,
		})
	}

	result := models.EntityPaginatedResponse{
		Results:    results,
		TotalCount: int(raw.TotalCount),
		Paging:     toPaging(raw.Paging.Next, raw.Paging.Previous),
	}

	return &result, nil
}

func toCreateEntityRequest(req *models.CreateEntityRequest) (*models.PostV3EntitiesRequest, error) {
	if req == nil || req.Name == nil || req.OrganizationID == nil || req.Category == nil || req.Status == nil || req.Type == nil {
		return nil, fmt.Errorf("create entity request is missing required fields")
	}

	metadata, err := rawMessageToMap(req.Metadata)
	if err != nil {
		return nil, err
	}
	classification, err := rawMessageToMap(req.Classification)
	if err != nil {
		return nil, err
	}

	body := &models.PostV3EntitiesRequest{
		Category:       models.PostV3EntitiesRequestCategory(*req.Category),
		Classification: classification,
		Metadata:       metadata,
		Name:           *req.Name,
		OrganizationId: toOpenapiUUID(*req.OrganizationID),
		ParentId:       toOpenapiUUIDPtr(req.ParentID),
		Status:         *req.Status,
		Type:           *req.Type,
	}
	if req.Affiliation != "" {
		affiliation := models.PostV3EntitiesRequestAffiliation(req.Affiliation)
		body.Affiliation = &affiliation
	}

	return body, nil
}

func toUpdateEntityRequest(req *models.UpdateEntityRequest) (*models.PutV3EntitiesbyEntityIdRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("update entity request is nil")
	}

	metadata, err := rawMessageToMap(req.Metadata)
	if err != nil {
		return nil, err
	}
	classification, err := rawMessageToMap(req.Classification)
	if err != nil {
		return nil, err
	}

	body := &models.PutV3EntitiesbyEntityIdRequest{
		Classification: classification,
		Metadata:       metadata,
		Name:           req.Name,
		ParentId:       toOpenapiUUIDPtr(req.ParentID),
		Type:           req.Type,
	}
	if req.Affiliation != "" {
		affiliation := models.PutV3EntitiesbyEntityIdRequestAffiliation(req.Affiliation)
		body.Affiliation = &affiliation
	}
	if req.Category != "" {
		category := models.PutV3EntitiesbyEntityIdRequestCategory(req.Category)
		body.Category = &category
	}
	if req.Status != "" {
		status := req.Status
		body.Status = &status
	}

	return body, nil
}

func toSearchEntitiesRequest(req *models.SearchEntitiesRequest) (*models.PostV3EntitiesSearchRequest, error) {
	if req == nil {
		return &models.PostV3EntitiesSearchRequest{}, nil
	}

	body := &models.PostV3EntitiesSearchRequest{}
	if req.Filters != nil {
		filters := &struct {
			Affiliation        *[]models.PostV3EntitiesSearchRequestFiltersAffiliation `json:"affiliation,omitempty"`
			Category           *[]models.PostV3EntitiesSearchRequestFiltersCategory    `json:"category,omitempty"`
			CreatedAfter       *string                                                 `json:"created_after,omitempty"`
			CreatedBefore      *string                                                 `json:"created_before,omitempty"`
			EntityIds          *[]openapi_types.UUID                                   `json:"entity_ids,omitempty"`
			Name               *string                                                 `json:"name,omitempty"`
			ParentIds          *[]openapi_types.UUID                                   `json:"parent_ids,omitempty"`
			Status             *[]string                                               `json:"status,omitempty"`
			TopClassifications *[]string                                               `json:"top_classifications,omitempty"`
			Types              *[]string                                               `json:"types,omitempty"`
			UpdatedAfter       *string                                                 `json:"updated_after,omitempty"`
			UpdatedBefore      *string                                                 `json:"updated_before,omitempty"`
		}{}
		if len(req.Filters.Affiliation) > 0 {
			values := make([]models.PostV3EntitiesSearchRequestFiltersAffiliation, len(req.Filters.Affiliation))
			for i, affiliation := range req.Filters.Affiliation {
				values[i] = models.PostV3EntitiesSearchRequestFiltersAffiliation(affiliation)
			}
			filters.Affiliation = &values
		}
		if len(req.Filters.Category) > 0 {
			values := make([]models.PostV3EntitiesSearchRequestFiltersCategory, len(req.Filters.Category))
			for i, category := range req.Filters.Category {
				values[i] = models.PostV3EntitiesSearchRequestFiltersCategory(category)
			}
			filters.Category = &values
		}
		if req.Filters.CreatedAfter != nil {
			filters.CreatedAfter = formatOptionalTime(req.Filters.CreatedAfter)
		}
		if req.Filters.CreatedBefore != nil {
			filters.CreatedBefore = formatOptionalTime(req.Filters.CreatedBefore)
		}
		if len(req.Filters.EntityIDs) > 0 {
			values := make([]openapi_types.UUID, len(req.Filters.EntityIDs))
			for i, id := range req.Filters.EntityIDs {
				values[i] = toOpenapiUUID(id)
			}
			filters.EntityIds = &values
		}
		if req.Filters.Name != "" {
			name := req.Filters.Name
			filters.Name = &name
		}
		if len(req.Filters.ParentIDs) > 0 {
			values := make([]openapi_types.UUID, len(req.Filters.ParentIDs))
			for i, id := range req.Filters.ParentIDs {
				values[i] = toOpenapiUUID(id)
			}
			filters.ParentIds = &values
		}
		if len(req.Filters.Status) > 0 {
			statuses := append([]string(nil), req.Filters.Status...)
			filters.Status = &statuses
		}
		if len(req.Filters.TopClassifications) > 0 {
			classifications := append([]string(nil), req.Filters.TopClassifications...)
			filters.TopClassifications = &classifications
		}
		types := req.Filters.Types
		if req.Filters.Type != "" {
			types = append(types, req.Filters.Type)
		}
		if len(types) > 0 {
			filterTypes := append([]string(nil), types...)
			filters.Types = &filterTypes
		}
		if req.Filters.UpdatedAfter != nil {
			filters.UpdatedAfter = formatOptionalTime(req.Filters.UpdatedAfter)
		}
		if req.Filters.UpdatedBefore != nil {
			filters.UpdatedBefore = formatOptionalTime(req.Filters.UpdatedBefore)
		}
		body.Filters = filters
	}

	if len(req.Sort) > 0 {
		sortFields := make([]struct {
			Field string                                      `json:"field"`
			Order models.PostV3EntitiesSearchRequestSortOrder `json:"order"`
		}, 0, len(req.Sort))
		for _, field := range req.Sort {
			order := models.PostV3EntitiesSearchRequestSortOrder(field.Order)
			sortFields = append(sortFields, struct {
				Field string                                      `json:"field"`
				Order models.PostV3EntitiesSearchRequestSortOrder `json:"order"`
			}{
				Field: field.Field,
				Order: order,
			})
		}
		body.Sort = &sortFields
	}

	return body, nil
}

func fromEntityResponse200(raw models.PostV3Entities200Response) (*models.EntityResponse, error) {
	return fromEntityFields(
		raw.Id,
		raw.OrganizationId,
		raw.Name,
		string(raw.Category),
		raw.Type,
		raw.Status,
		string(raw.Affiliation),
		raw.ParentId,
		raw.Metadata,
		raw.Classification,
		raw.TopClassification,
		raw.TopClassificationProbability,
		raw.CreatedAt,
		raw.UpdatedAt,
		raw.DeletedAt,
	)
}

func fromEntityResponse201(raw models.PostV3Entities201Response) (*models.EntityResponse, error) {
	return fromEntityFields(
		raw.Id,
		raw.OrganizationId,
		raw.Name,
		string(raw.Category),
		raw.Type,
		raw.Status,
		string(raw.Affiliation),
		raw.ParentId,
		raw.Metadata,
		raw.Classification,
		raw.TopClassification,
		raw.TopClassificationProbability,
		raw.CreatedAt,
		raw.UpdatedAt,
		raw.DeletedAt,
	)
}

func fromGetEntityResponse(raw models.GetV3EntitiesbyEntityId200Response) (*models.EntityResponse, error) {
	return fromEntityFields(
		raw.Id,
		raw.OrganizationId,
		raw.Name,
		string(raw.Category),
		raw.Type,
		raw.Status,
		string(raw.Affiliation),
		raw.ParentId,
		raw.Metadata,
		raw.Classification,
		raw.TopClassification,
		raw.TopClassificationProbability,
		raw.CreatedAt,
		raw.UpdatedAt,
		raw.DeletedAt,
	)
}

func fromPutEntityResponse(raw models.PutV3EntitiesbyEntityId200Response) (*models.EntityResponse, error) {
	return fromEntityFields(
		raw.Id,
		raw.OrganizationId,
		raw.Name,
		string(raw.Category),
		raw.Type,
		raw.Status,
		string(raw.Affiliation),
		raw.ParentId,
		raw.Metadata,
		raw.Classification,
		raw.TopClassification,
		raw.TopClassificationProbability,
		raw.CreatedAt,
		raw.UpdatedAt,
		raw.DeletedAt,
	)
}

func fromEntityFields(
	id openapi_types.UUID,
	orgID openapi_types.UUID,
	name string,
	category string,
	entityType string,
	status string,
	affiliation string,
	parentID *openapi_types.UUID,
	metadata *map[string]interface{},
	classification *map[string]interface{},
	topClassification *string,
	topClassificationProbability *float32,
	createdAt string,
	updatedAt string,
	deletedAt *string,
) (*models.EntityResponse, error) {
	metadataRaw, err := mapToRawMessage(metadata)
	if err != nil {
		return nil, err
	}
	classificationRaw, err := mapToRawMessage(classification)
	if err != nil {
		return nil, err
	}
	createdTime, err := parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	updatedTime, err := parseTime(updatedAt)
	if err != nil {
		return nil, err
	}
	deletedTime, err := parseOptionalTime(deletedAt)
	if err != nil {
		return nil, err
	}

	return &models.EntityResponse{
		ID:                           fromOpenapiUUID(id),
		OrganizationID:               fromOpenapiUUID(orgID),
		Name:                         name,
		Category:                     models.Category(category),
		Type:                         entityType,
		Status:                       status,
		Affiliation:                  models.Affiliation(affiliation),
		ParentID:                     uuidPtrFromOpenapi(parentID),
		Metadata:                     metadataRaw,
		Classification:               classificationRaw,
		TopClassification:            topClassification,
		TopClassificationProbability: topClassificationProbability,
		CreatedAt:                    createdTime,
		UpdatedAt:                    updatedTime,
		DeletedAt:                    deletedTime,
	}, nil
}

func uuidPtrFromOpenapi(id *openapi_types.UUID) *uuid.UUID {
	if id == nil {
		return nil
	}

	converted := fromOpenapiUUID(*id)
	return &converted
}
