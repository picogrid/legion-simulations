package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateFeedDefinitionRequest represents the request to create a new feed definition
// @Description Request body for creating a new feed definition with its metadata and schema.
// @name CreateFeedDefinitionRequest
type CreateFeedDefinitionRequest struct {
	// Category indicates the primary intended handling ('MESSAGE', 'FILE')
	Category DataCategory `json:"category" binding:"required" swaggertype:"string" enums:"MESSAGE,FILE" example:"MESSAGE"`

	// FeedName is the user-friendly name of the feed type, unique per organization
	FeedName string `json:"feed_name" binding:"required" example:"temperature_readings"`

	// DataType is a user-defined string describing the data (e.g., 'image/jpeg')
	DataType string `json:"data_type" binding:"required" example:"application/json"`

	// Description provides optional details about the feed definition
	Description *string `json:"description,omitempty" example:"Temperature readings from IoT sensors"`

	// SchemaDefinition holds an optional JSON Schema definition or reference
	SchemaDefinition *json.RawMessage `json:"schema_definition,omitempty" swaggertype:"string" format:"json" example:"{\"type\":\"object\",\"properties\":{\"temperature\":{\"type\":\"number\"},\"unit\":{\"type\":\"string\"}}}"`

	// IsActive indicates if the feed definition is currently active
	IsActive bool `json:"is_active" example:"true"`

	// EntityID links this definition to a specific entity (NULL for org-level templates)
	EntityID *uuid.UUID `json:"entity_id,omitempty" swaggertype:"string" format:"uuid" example:"c3d4e5f6-a1b2-7890-1234-567890abcdef"`

	// TemplateID references the template this feed was created from (if applicable)
	TemplateID *uuid.UUID `json:"template_id,omitempty" swaggertype:"string" format:"uuid" example:"d4e5f6a1-b2c3-7890-1234-567890abcdef"`

	// IntegrationID tracks which integration created/manages this feed definition
	IntegrationID *uuid.UUID `json:"integration_id,omitempty" swaggertype:"string" format:"uuid" example:"e5f6a1b2-c3d4-7890-1234-567890abcdef"`

	// IsTemplate indicates if this is an org-level template (true) or entity-specific feed (false)
	IsTemplate bool `json:"is_template" example:"false"`

	// Metadata holds additional information about the feed definition
	Metadata *json.RawMessage `json:"metadata,omitempty" swaggertype:"string" format:"json" example:"{\"custom_field\":\"custom_value\"}"`
}

// UpdateFeedDefinitionRequest represents the request to update an existing feed definition
// @Description Request body for updating an existing feed definition's metadata and schema.
// @name UpdateFeedDefinitionRequest
type UpdateFeedDefinitionRequest struct {
	// Category indicates the primary intended handling ('MESSAGE', 'FILE')
	Category DataCategory `json:"category" binding:"required" swaggertype:"string" enums:"MESSAGE,FILE" example:"FILE"`

	// FeedName is the user-friendly name of the feed type, unique per organization
	FeedName string `json:"feed_name" binding:"required" example:"image_uploads"`

	// DataType is a user-defined string describing the data (e.g., 'image/jpeg')
	DataType string `json:"data_type" binding:"required" example:"image/jpeg"`

	// Description provides optional details about the feed definition
	Description *string `json:"description,omitempty" example:"Image uploads from mobile devices"`

	// SchemaDefinition holds an optional JSON Schema definition or reference
	SchemaDefinition *json.RawMessage `json:"schema_definition,omitempty" swaggertype:"string" format:"json" example:"{\"type\":\"object\",\"properties\":{\"width\":{\"type\":\"integer\"},\"height\":{\"type\":\"integer\"}}}"`

	// IsActive indicates if the feed definition is currently active
	IsActive bool `json:"is_active" example:"true"`

	// EntityID links this definition to a specific entity (NULL for org-level templates)
	EntityID *uuid.UUID `json:"entity_id,omitempty" swaggertype:"string" format:"uuid" example:"c3d4e5f6-a1b2-7890-1234-567890abcdef"`

	// TemplateID references the template this feed was created from (if applicable)
	TemplateID *uuid.UUID `json:"template_id,omitempty" swaggertype:"string" format:"uuid" example:"d4e5f6a1-b2c3-7890-1234-567890abcdef"`

	// IntegrationID tracks which integration created/manages this feed definition
	IntegrationID *uuid.UUID `json:"integration_id,omitempty" swaggertype:"string" format:"uuid" example:"e5f6a1b2-c3d4-7890-1234-567890abcdef"`

	// IsTemplate indicates if this is an org-level template (true) or entity-specific feed (false)
	IsTemplate bool `json:"is_template" example:"false"`

	// Metadata holds additional information about the feed definition
	Metadata *json.RawMessage `json:"metadata,omitempty" swaggertype:"string" format:"json" example:"{\"custom_field\":\"custom_value\"}"`
}

// FeedDefinitionResponse represents the response for a feed definition
// @Description Contains the complete details of a feed definition, including its ID, metadata, schema, and timestamps.
// @name FeedDefinitionResponse
type FeedDefinitionResponse struct {
	// ID is the unique identifier for the feed definition
	ID uuid.UUID `json:"id" swaggertype:"string" format:"uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`

	// OrganizationID links this definition to a specific organization
	OrganizationID uuid.UUID `json:"organization_id" swaggertype:"string" format:"uuid" example:"b2c3d4e5-f6a1-7890-1234-567890abcdef"`

	// EntityID links this definition to a specific entity (NULL for org-level templates)
	EntityID *uuid.UUID `json:"entity_id,omitempty" swaggertype:"string" format:"uuid" example:"c3d4e5f6-a1b2-7890-1234-567890abcdef"`

	// TemplateID references the template this feed was created from (if applicable)
	TemplateID *uuid.UUID `json:"template_id,omitempty" swaggertype:"string" format:"uuid" example:"d4e5f6a1-b2c3-7890-1234-567890abcdef"`

	// IntegrationID tracks which integration created/manages this feed definition
	IntegrationID *uuid.UUID `json:"integration_id,omitempty" swaggertype:"string" format:"uuid" example:"e5f6a1b2-c3d4-7890-1234-567890abcdef"`

	// Category indicates the primary intended handling ('MESSAGE', 'FILE')
	Category DataCategory `json:"category" swaggertype:"string" enums:"MESSAGE,FILE" example:"MESSAGE"`

	// FeedName is the user-friendly name of the feed type, unique per organization
	FeedName string `json:"feed_name" example:"temperature_readings"`

	// DataType is a user-defined string describing the data (e.g., 'image/jpeg')
	DataType string `json:"data_type" example:"application/json"`

	// Description provides optional details about the feed definition
	Description *string `json:"description,omitempty" example:"Temperature readings from IoT sensors"`

	// SchemaDefinition holds an optional JSON Schema definition or reference
	SchemaDefinition *json.RawMessage `json:"schema_definition,omitempty" swaggertype:"string" format:"json" example:"{\"type\":\"object\",\"properties\":{\"temperature\":{\"type\":\"number\"},\"unit\":{\"type\":\"string\"}}}"`

	// IsActive indicates if the feed definition is currently active
	IsActive bool `json:"is_active" example:"true"`

	// IsTemplate indicates if this is an org-level template (true) or entity-specific feed (false)
	IsTemplate bool `json:"is_template" example:"false"`

	// Metadata holds additional information about the feed definition
	Metadata *json.RawMessage `json:"metadata,omitempty" swaggertype:"string" format:"json" example:"{\"custom_field\":\"custom_value\"}"`

	// CreatedAt records when the feed definition was created
	CreatedAt time.Time `json:"created_at" swaggertype:"string" format:"date-time" example:"2024-02-16T21:45:33Z"`

	// UpdatedAt records when the feed definition was last updated
	UpdatedAt time.Time `json:"updated_at" swaggertype:"string" format:"date-time" example:"2024-02-16T21:45:33Z"`
}

// FeedDefinitionListResponse represents a paginated list of feed definitions
// @Description A paginated response containing a list of feed definitions, total count, and pagination information.
// @name FeedDefinitionListResponse
type FeedDefinitionListResponse struct {
	// Results contains the list of feed definitions for the current page
	Results []FeedDefinitionResponse `json:"results"`

	// TotalCount is the total number of feed definitions matching the query
	TotalCount int `json:"total_count" example:"42"`

	// Paging contains optional paging information
	Paging Paging `json:"paging,omitempty"`
}

// FeedDefinitionSearchRequest represents the request to search feed definitions
// @Description Request body for searching feed definitions based on various criteria
// @name FeedDefinitionSearchRequest
type FeedDefinitionSearchRequest struct {
	// OrganizationID filters by organization
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" swaggertype:"string" format:"uuid" example:"b2c3d4e5-f6a1-7890-1234-567890abcdef"`

	// EntityID filters by entity that owns the feed definition
	EntityID *uuid.UUID `json:"entity_id,omitempty" swaggertype:"string" format:"uuid" example:"c3d4e5f6-a1b2-7890-1234-567890abcdef"`

	// TemplateID filters by template this definition was created from
	TemplateID *uuid.UUID `json:"template_id,omitempty" swaggertype:"string" format:"uuid" example:"d4e5f6a1-b2c3-7890-1234-567890abcdef"`

	// IntegrationID filters by integration that manages this definition
	IntegrationID *uuid.UUID `json:"integration_id,omitempty" swaggertype:"string" format:"uuid" example:"e5f6a1b2-c3d4-7890-1234-567890abcdef"`

	// IsTemplate filters by whether this is a template
	IsTemplate *bool `json:"is_template,omitempty" example:"false"`

	// Category filters by data category
	Category *DataCategory `json:"category,omitempty" swaggertype:"string" enums:"MESSAGE,FILE" example:"MESSAGE"`

	// IsActive filters by active status
	IsActive *bool `json:"is_active,omitempty" example:"true"`
}
