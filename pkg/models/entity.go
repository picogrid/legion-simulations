package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateEntityRequest represents the request structure for creating a new entity.
// @Description Request body containing the details needed to create a new entity.
// @name CreateEntityRequest
type CreateEntityRequest struct {
	// The unique identifier of the organization to which the entity belongs.
	OrganizationID uuid.UUID `json:"organization_id" binding:"required" swaggertype:"string,uuid" example:"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f"`
	// A user-defined name for the entity.
	Name string `json:"name" binding:"required" example:"Axis Camera"`
	// The category of the entity (e.g., DEVICE, EVENT, ZONE).
	Category Category `json:"category" binding:"required" swaggertype:"string" enums:"DEVICE,EVENT,ZONE"`
	// A specific type within the category.
	Type string `json:"type" binding:"required" example:"Camera"`
	// The current operational status of the entity.
	Status string `json:"status" binding:"required" example:"active"`
	// Arbitrary metadata associated with the entity (JSON object).
	Metadata *json.RawMessage `json:"metadata,omitempty" swaggertype:"primitive,string" example:"{\"ip_address\": \"192.168.1.100\", \"resolution\": \"1080p\"}"`
}

// UpdateEntityRequest represents the request structure for updating an existing entity.
// @Description Request body for updating properties of an entity. Only provided fields will be updated.
// @name UpdateEntityRequest
type UpdateEntityRequest struct {
	// The unique identifier of the entity to be updated.
	ID uuid.UUID `json:"id" swaggertype:"string,uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	// The new name for the entity.
	Name *string `json:"name,omitempty" example:"Axis Camera"`
	// The new category for the entity.
	Category *Category `json:"category,omitempty" swaggertype:"string" enums:"DEVICE,LOCATION,SENSOR,ZONE"` // Assuming enum values
	// The new type for the entity.
	Type *string `json:"type,omitempty" example:"PTZ Camera"`
	// The new status for the entity.
	Status *string `json:"status,omitempty" example:"inactive"`
	// @Description New or updated metadata for the entity (JSON object). This will replace existing metadata.
	Metadata *json.RawMessage `json:"metadata,omitempty" swaggertype:"primitive,string" example:"{\"ip_address\": \"192.168.1.101\", \"firmware\": \"v2.1\"}"`
}

// EntityResponse represents the detailed structure of an entity as returned by the API.
// @Description Contains the full details of an entity, including its properties, metadata, and timestamps.
// @name EntityResponse
type EntityResponse struct {
	// The unique identifier of the entity.
	ID uuid.UUID `json:"id" swaggertype:"string,uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	// The unique identifier of the organization to which the entity belongs.
	OrganizationID uuid.UUID `json:"organization_id" swaggertype:"string,uuid" example:"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f"`
	// The name of the entity.
	Name string `json:"name" example:"Axis Camera"`
	// The category of the entity.
	Category Category `json:"category" swaggertype:"string" enums:"DEVICE,DETECTION,ZONE,ALERT,WEATHER,GEOMETRIC"`
	// The specific type within the category.
	Type string `json:"type" example:"Camera"`
	// The current operational status of the entity.
	Status string `json:"status" example:"active"`
	// Arbitrary metadata associated with the entity (JSON object).
	Metadata *json.RawMessage `json:"metadata,omitempty" swaggertype:"primitive,string" example:"{\"ip_address\": \"192.168.1.100\", \"resolution\": \"1080p\"}"`
	// Timestamp indicating when the entity was created.
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T10:00:00Z" format:"date-time"`
	// Timestamp indicating when the entity was last updated.
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-15T15:30:45Z" format:"date-time"`
	// Timestamp indicating when the entity was deleted (soft delete). Null if not deleted.
	DeletedAt *time.Time `json:"deleted_at,omitempty" example:"null" format:"date-time"`
}

// EntityPaginatedResponse represents a paginated response of Entity
// @Description A paginated list of Entities
// @name EntityPaginatedResponse
type EntityPaginatedResponse struct {
	// Results is a slice of items of type EntityResponse.
	Results []EntityResponse `json:"results"`
	// TotalCount is the total number of items available.
	TotalCount int `json:"total_count" swaggertype:"integer" example:"100"`
	// Paging contains optional paging information.
	Paging Paging `json:"paging,omitempty"`
}

// SearchFilters represents the available filters for searching entities.
// @Description Defines the criteria that can be used to filter entity search results.
// @name SearchFilters
type SearchFilters struct {
	// Filter by a specific entity name.
	Name string `json:"name,omitempty"`
	// Filter by one or more entity categories.
	Category []Category `json:"category,omitempty"`
	// Filter by one or more entity statuses.
	Status []string `json:"status,omitempty"`
	// Filter by one or more entity types.
	Types []string `json:"types,omitempty"`
	// Filter entities created after this timestamp.
	CreatedAfter time.Time `json:"created_after,omitempty"`
	// Filter entities created before this timestamp.
	CreatedBefore time.Time `json:"created_before,omitempty"`
	// Filter entities updated after this timestamp.
	UpdatedAfter time.Time `json:"updated_after,omitempty"`
	// Filter entities updated before this timestamp.
	UpdatedBefore time.Time `json:"updated_before,omitempty"`
	// Filter by a specific list of entity IDs.
	EntityIDs []uuid.UUID `json:"entity_ids,omitempty"`
}

// SortOption represents a single field and direction for sorting search results.
// @Description Defines how to sort the entity search results based on a specific field.
// @name SortOption
type SortOption struct {
	// The field to sort by (e.g., "created_at", "name", "status").
	Field string `json:"field" binding:"required" example:"created_at"`
	// The sort order.
	Order string `json:"order" binding:"required,oneof=asc desc" example:"desc" enums:"asc,desc"`
}

// SearchEntitiesRequest represents the request structure for searching entities.
// @Description Request body containing filters and sorting options for searching entities within an organization.
// @name SearchEntitiesRequest
type SearchEntitiesRequest struct {
	// The unique identifier of the organization in which to search for entities.
	OrganizationID uuid.UUID `json:"organization_id" binding:"required" swaggertype:"string,uuid" example:"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f"`
	// Optional filters to apply to the entity search.
	Filters *SearchFilters `json:"filters,omitempty"`
	// Optional sorting criteria for the search results.
	Sort []*SortOption `json:"sort,omitempty"`
}
