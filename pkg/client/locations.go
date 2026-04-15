package client

import (
	"context"
	"fmt"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/picogrid/legion-simulations/pkg/models"
)

// CreateEntityLocation creates a new location for an entity
func (c *Legion) CreateEntityLocation(ctx context.Context, entityID string, req *models.CreateEntityLocationRequest) (*models.EntityLocationResponse, error) {
	body, err := toCreateEntityLocationRequest(req)
	if err != nil {
		return nil, fmt.Errorf("build entity location request: %w", err)
	}

	path := fmt.Sprintf("/v3/entities/%s/locations", entityID)
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity location: %w", err)
	}

	var raw models.PostV3EntitiesbyEntityIdLocations201Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode location response: %w", err)
	}

	return fromLocation201(raw)
}

// GetEntityLocation gets a specific location for an entity
func (c *Legion) GetEntityLocation(ctx context.Context, entityID, locationID string) (*models.EntityLocationResponse, error) {
	path := fmt.Sprintf("/v3/entities/%s/locations/%s", entityID, locationID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity location: %w", err)
	}

	var raw models.GetV3EntitiesbyEntityIdLocationsbyEntityLocationId200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode location response: %w", err)
	}

	return fromLocationByID(raw)
}

// GetEntityLocations gets all locations for an entity
func (c *Legion) GetEntityLocations(ctx context.Context, entityID string) (*models.EntityLocationPaginatedResponse, error) {
	path := fmt.Sprintf("/v3/entities/%s/locations", entityID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity locations: %w", err)
	}

	var raw models.GetV3EntitiesbyEntityIdLocations200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode locations response: %w", err)
	}

	results := make([]models.EntityLocationResponse, 0, len(raw.Results))
	for _, location := range raw.Results {
		converted, err := locationFromParts(
			location.Id,
			location.EntityId,
			location.Entity,
			location.Position.Type,
			location.Position.Coordinates,
			location.Source,
			location.RecordedAt,
			location.CreatedAt,
			location.Acceleration,
			location.AngularVelocity,
			location.Bearing,
			location.Orientation,
			location.Radius,
			location.Speed,
			location.Velocity,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, *converted)
	}

	result := models.EntityLocationPaginatedResponse{
		Results:    results,
		TotalCount: int(raw.TotalCount),
		Paging:     toPaging(raw.Paging.Next, raw.Paging.Previous),
	}

	return &result, nil
}

// SearchEntityLocations searches for entity locations based on criteria
func (c *Legion) SearchEntityLocations(ctx context.Context, req *models.SearchEntityLocationsRequest) (*models.EntityLocationPaginatedResponse, error) {
	body := &models.PostV3EntitiesLocationsSearchRequest{}
	if req != nil {
		filters := &struct {
			Categories      *[]models.PostV3EntitiesLocationsSearchRequestFiltersCategories `json:"categories,omitempty"`
			CreatedAfter    *string                                                         `json:"created_after,omitempty"`
			CreatedBefore   *string                                                         `json:"created_before,omitempty"`
			EntityIds       *[]openapi_types.UUID                                           `json:"entity_ids,omitempty"`
			ProximityFilter *struct {
				Crs      *models.PostV3EntitiesLocationsSearchRequestFiltersProximityFilterCrs `json:"crs,omitempty"`
				Position []float32                                                             `json:"position"`
				Radius   float32                                                               `json:"radius"`
			} `json:"proximity_filter,omitempty"`
			RecordedAfter  *string   `json:"recorded_after,omitempty"`
			RecordedBefore *string   `json:"recorded_before,omitempty"`
			Sources        *[]string `json:"sources,omitempty"`
		}{}
		if len(req.EntityIDs) > 0 {
			values := make([]openapi_types.UUID, len(req.EntityIDs))
			for i, id := range req.EntityIDs {
				values[i] = toOpenapiUUID(id)
			}
			filters.EntityIds = &values
		}
		filters.RecordedAfter = formatOptionalTime(req.RecordedAfter)
		filters.RecordedBefore = formatOptionalTime(req.RecordedBefore)
		if len(req.Sources) > 0 {
			sources := append([]string(nil), req.Sources...)
			filters.Sources = &sources
		}
		body.Filters = filters
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/entities/locations/search", body)
	if err != nil {
		return nil, fmt.Errorf("failed to search entity locations: %w", err)
	}

	var raw models.PostV3EntitiesLocationsSearch200Response
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	results := make([]models.EntityLocationResponse, 0, len(raw.Results))
	for _, location := range raw.Results {
		converted, err := locationFromParts(
			location.Id,
			location.EntityId,
			location.Entity,
			location.Position.Type,
			location.Position.Coordinates,
			location.Source,
			location.RecordedAt,
			location.CreatedAt,
			location.Acceleration,
			location.AngularVelocity,
			location.Bearing,
			location.Orientation,
			location.Radius,
			location.Speed,
			location.Velocity,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, *converted)
	}

	result := models.EntityLocationPaginatedResponse{
		Results:    results,
		TotalCount: int(raw.TotalCount),
		Paging:     toPaging(raw.Paging.Next, raw.Paging.Previous),
	}

	return &result, nil
}

func toCreateEntityLocationRequest(req *models.CreateEntityLocationRequest) (*models.PostV3EntitiesbyEntityIdLocationsRequest, error) {
	if req == nil || req.Position == nil || req.Position.Type == nil || len(req.Position.Coordinates) != 3 || req.Source == "" {
		return nil, fmt.Errorf("create entity location request is missing required fields")
	}

	body := &models.PostV3EntitiesbyEntityIdLocationsRequest{
		Acceleration:    float64SlicePtrTo32(req.Acceleration),
		AngularVelocity: float64SlicePtrTo32(req.AngularVelocity),
		Bearing:         float64PtrTo32(req.Bearing),
		Covariance:      nil,
		Orientation:     float64SlicePtrTo32(req.Orientation),
		Position: struct {
			Coordinates []float32 `json:"coordinates"`
			Type        string    `json:"type"`
		}{
			Coordinates: float64SliceTo32(req.Position.Coordinates),
			Type:        *req.Position.Type,
		},
		Radius:     float64PtrTo32(req.Radius),
		RecordedAt: formatOptionalTime(req.RecordedAt),
		Source:     req.Source,
		Velocity:   float64SlicePtrTo32(req.Velocity),
	}

	return body, nil
}

func fromLocation201(raw models.PostV3EntitiesbyEntityIdLocations201Response) (*models.EntityLocationResponse, error) {
	return locationFromParts(
		raw.Id,
		raw.EntityId,
		raw.Entity,
		raw.Position.Type,
		raw.Position.Coordinates,
		raw.Source,
		raw.RecordedAt,
		raw.CreatedAt,
		raw.Acceleration,
		raw.AngularVelocity,
		raw.Bearing,
		raw.Orientation,
		raw.Radius,
		raw.Speed,
		raw.Velocity,
	)
}

func fromLocationByID(raw models.GetV3EntitiesbyEntityIdLocationsbyEntityLocationId200Response) (*models.EntityLocationResponse, error) {
	return locationFromParts(
		raw.Id,
		raw.EntityId,
		raw.Entity,
		raw.Position.Type,
		raw.Position.Coordinates,
		raw.Source,
		raw.RecordedAt,
		raw.CreatedAt,
		raw.Acceleration,
		raw.AngularVelocity,
		raw.Bearing,
		raw.Orientation,
		raw.Radius,
		raw.Speed,
		raw.Velocity,
	)
}

func locationFromParts(
	id openapi_types.UUID,
	entityID openapi_types.UUID,
	entity interface{},
	positionType string,
	coordinates []float32,
	source string,
	recordedAt *string,
	createdAt string,
	acceleration *[]float32,
	angularVelocity *[]float32,
	bearing *float32,
	orientation *[]float32,
	radius *float32,
	speed *float32,
	velocity *[]float32,
) (*models.EntityLocationResponse, error) {
	createdTime, err := parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	recordedTime, err := parseOptionalTime(recordedAt)
	if err != nil {
		return nil, err
	}

	location := &models.EntityLocationResponse{
		ID:         fromOpenapiUUID(id),
		EntityID:   fromOpenapiUUID(entityID),
		Position:   models.GeomPoint{Type: &positionType, Coordinates: float32SliceTo64(coordinates)},
		Source:     source,
		RecordedAt: recordedTime,
		CreatedAt:  createdTime,
		Bearing:    float32PtrTo64(bearing),
		Radius:     float32PtrTo64(radius),
		Speed:      float32PtrTo64(speed),
	}
	if acceleration != nil {
		location.Acceleration = float32SliceTo64(*acceleration)
	}
	if angularVelocity != nil {
		location.AngularVelocity = float32SliceTo64(*angularVelocity)
	}
	if orientation != nil {
		location.Orientation = float32SliceTo64(*orientation)
	}
	if velocity != nil {
		location.Velocity = float32SliceTo64(*velocity)
	}

	if entity != nil {
		if converted, err := entityFromAnonymous(entity); err == nil {
			location.Entity = converted
		}
	}

	return location, nil
}

func entityFromAnonymous(raw interface{}) (*models.EntityResponse, error) {
	switch entity := raw.(type) {
	case *struct {
		Affiliation    models.GetV3EntitiesbyEntityIdLocations200ResponseResultsEntityAffiliation `json:"affiliation"`
		Category       models.GetV3EntitiesbyEntityIdLocations200ResponseResultsEntityCategory    `json:"category"`
		CreatedAt      string                                                                     `json:"created_at"`
		DeletedAt      *string                                                                    `json:"deleted_at,omitempty"`
		Id             openapi_types.UUID                                                         `json:"id"`
		Metadata       *map[string]interface{}                                                    `json:"metadata,omitempty"`
		Name           string                                                                     `json:"name"`
		OrganizationId openapi_types.UUID                                                         `json:"organization_id"`
		ParentId       *openapi_types.UUID                                                        `json:"parent_id"`
		Status         string                                                                     `json:"status"`
		Type           string                                                                     `json:"type"`
		UpdatedAt      string                                                                     `json:"updated_at"`
	}:
		return fromEntityFields(entity.Id, entity.OrganizationId, entity.Name, string(entity.Category), entity.Type, entity.Status, string(entity.Affiliation), entity.ParentId, entity.Metadata, nil, nil, nil, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt)
	case *struct {
		Affiliation    models.PostV3EntitiesLocationsSearch200ResponseResultsEntityAffiliation `json:"affiliation"`
		Category       models.PostV3EntitiesLocationsSearch200ResponseResultsEntityCategory    `json:"category"`
		CreatedAt      string                                                                  `json:"created_at"`
		DeletedAt      *string                                                                 `json:"deleted_at,omitempty"`
		Id             openapi_types.UUID                                                      `json:"id"`
		Metadata       *map[string]interface{}                                                 `json:"metadata,omitempty"`
		Name           string                                                                  `json:"name"`
		OrganizationId openapi_types.UUID                                                      `json:"organization_id"`
		ParentId       *openapi_types.UUID                                                     `json:"parent_id"`
		Status         string                                                                  `json:"status"`
		Type           string                                                                  `json:"type"`
		UpdatedAt      string                                                                  `json:"updated_at"`
	}:
		return fromEntityFields(entity.Id, entity.OrganizationId, entity.Name, string(entity.Category), entity.Type, entity.Status, string(entity.Affiliation), entity.ParentId, entity.Metadata, nil, nil, nil, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt)
	case *struct {
		Affiliation    models.GetV3EntitiesbyEntityIdLocationsbyEntityLocationId200ResponseEntityAffiliation `json:"affiliation"`
		Category       models.GetV3EntitiesbyEntityIdLocationsbyEntityLocationId200ResponseEntityCategory    `json:"category"`
		CreatedAt      string                                                                                `json:"created_at"`
		DeletedAt      *string                                                                               `json:"deleted_at,omitempty"`
		Id             openapi_types.UUID                                                                    `json:"id"`
		Metadata       *map[string]interface{}                                                               `json:"metadata,omitempty"`
		Name           string                                                                                `json:"name"`
		OrganizationId openapi_types.UUID                                                                    `json:"organization_id"`
		ParentId       *openapi_types.UUID                                                                   `json:"parent_id"`
		Status         string                                                                                `json:"status"`
		Type           string                                                                                `json:"type"`
		UpdatedAt      string                                                                                `json:"updated_at"`
	}:
		return fromEntityFields(entity.Id, entity.OrganizationId, entity.Name, string(entity.Category), entity.Type, entity.Status, string(entity.Affiliation), entity.ParentId, entity.Metadata, nil, nil, nil, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt)
	default:
		return nil, fmt.Errorf("unsupported entity shape")
	}
}

func float64SlicePtrTo32(values []float64) *[]float32 {
	if len(values) == 0 {
		return nil
	}

	converted := float64SliceTo32(values)
	return &converted
}
