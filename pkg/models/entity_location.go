package models

import (
	"time"

	"github.com/google/uuid"
)

type GeomPoint struct {
	Type        string    `json:"type" example:"Point"`
	Coordinates []float64 `json:"coordinates" swaggertype:"primitive,string" example:"[3980581.21, -482586.522, 4966824.593]"`
}

// ProximityFilter defines search criteria based on distance from a point.
// @Description Specifies a center point (X, Y, Z) and a radius for proximity searches.
// @name ProximityFilter
type ProximityFilter struct {
	X      float64 `json:"x"`      // ECEF X coordinate in meters
	Y      float64 `json:"y"`      // ECEF Y coordinate in meters
	Z      float64 `json:"z"`      // ECEF Z coordinate in meters
	Radius float64 `json:"radius"` // Search radius in meters
}

// CreateEntityLocationRequest represents the request structure for creating a new entity location.
// @Description Request body containing the details needed to create a new entity location.
// @name CreateEntityLocationRequest
type CreateEntityLocationRequest struct {
	// A 3D point (X, Y, Z) representing the entity's position in Earth-Centered, Earth-Fixed (ECEF) coordinates (SRID/EPSG 4978).
	Position *GeomPoint `json:"position" binding:"required" swaggertype:"primitive,string" example:"{\"type\":\"Point\",\"coordinates\":[3980581.21,-482586.522,4966824.593]}"`
	// The orientation of the entity as a quaternion in [w, x, y, z] format. Nullable.
	Orientation *[]float32 `json:"orientation,omitempty" swaggertype:"primitive,string" example:"[1.0, 0.0, 0.0, 0.0]"`
	// The timestamp when the entity location was recorded by the sender (e.g., the device or upstream system).
	// This is provided by the client and may differ from the time this record was created in the database. Nullable.
	RecordedAt *time.Time `json:"recorded_at" example:"2023-10-01T12:00:00Z" format:"date-time"`
}

// EntityLocationResponse represents the detailed structure of an entity location as returned by the API.
// @Description Contains the full details of an entity location, including its position, orientation, and timestamps.
// @name EntityLocationResponse
type EntityLocationResponse struct {
	// The unique identifier of the entity location.
	ID uuid.UUID `json:"id" swaggertype:"string,uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	// The unique identifier of the entity to which the entity location belongs.
	EntityID uuid.UUID `json:"entity_id" swaggertype:"string,uuid" example:"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f"`
	// A 3D point (X, Y, Z) representing the entity's position in Earth-Centered, Earth-Fixed (ECEF) coordinates (SRID/EPSG 4978).
	Position *GeomPoint `json:"position" swaggertype:"primitive,string" example:"{\"type\":\"Point\",\"coordinates\":[3980581.21,-482586.522,4966824.593]}"`
	// The orientation of the entity as a quaternion in [w, x, y, z] format. Nullable.
	Orientation *[]float32 `json:"orientation,omitempty" swaggertype:"primitive,string" example:"[1.0, 0.0, 0.0, 0.0]"`
	// The timestamp when the entity location was recorded by the sender (e.g., the device or upstream system).
	// This is provided by the client and may differ from the time this record was created in the database. Nullable.
	// If not provided, the database will default to NOW().
	RecordedAt *time.Time `json:"recorded_at,omitempty" example:"2023-10-01T12:00:00Z" format:"date-time"`
	// The timestamp when the entity location record was created in the system.
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T10:00:00Z" format:"date-time"`
	// The full entity data (only populated when hydrate_entities is true).
	Entity *EntityResponse `json:"entity,omitempty"`
}

// EntityLocationPaginatedResponse represents a paginated response of Entity Locations
// @Description A paginated list of Entity Locations
// @name EntityLocationPaginatedResponse
type EntityLocationPaginatedResponse struct {
	// Results is a slice of items of type EntityResponse.
	Results []EntityLocationResponse `json:"results"`
	// TotalCount is the total number of items available.
	TotalCount int `json:"total_count" swaggertype:"integer" example:"100"`
	// Paging contains optional paging information.
	Paging Paging `json:"paging,omitempty"`
}

// SearchLocationFilters represents the available filters for searching entity locations.
// @Description Defines the criteria that can be used to filter entity location search results.
// @name SearchLocationFilters
type SearchLocationFilters struct {
	// Filter by proximity to a specific position (X, Y, Z) in Earth-Centered, Earth-Fixed (ECEF) coordinates (SRID/EPSG 4978).
	ProximityFilter *ProximityFilter `json:"proximity_filter,omitempty"`
	// Filter entity locations created after this timestamp.
	CreatedAfter time.Time `json:"created_after,omitempty" format:"date-time" example:"2023-10-01T00:00:00Z"`
	// Filter entity locations created before this timestamp.
	CreatedBefore time.Time `json:"created_before,omitempty" format:"date-time" example:"2023-10-01T23:59:59Z"`
	// Filter entity locations recorded after this timestamp.
	RecordedAfter time.Time `json:"recorded_after,omitempty" format:"date-time" example:"2023-10-01T00:00:00Z"`
	// Filter entity locations recorded before this timestamp.
	RecordedBefore time.Time `json:"recorded_before,omitempty" format:"date-time" example:"2023-10-01T23:59:59Z"`
	// Filter by specific entity IDs for batch lookups.
	EntityIDs []uuid.UUID `json:"entity_ids,omitempty" swaggertype:"array,string" example:"[\"a1b2c3d4-e5f6-7890-1234-567890abcdef\",\"b7c5e4d3-a2b1-4f0e-8d9c-1a2b3c4d5e6f\"]"`
}

// SearchEntityLocationsRequest represents the request structure for searching entity locations.
// @Description Request body containing filters and sorting options for searching entity locations.
// @name SearchEntityLocationsRequest
type SearchEntityLocationsRequest struct {
	// Optional filters to apply to the entity location search.
	Filters *SearchLocationFilters `json:"filters,omitempty"`
	// Optional sorting criteria for the search results.
	Sort []*SortOption `json:"sort,omitempty"`
	// Whether to include full entity data with each location (requires entity join).
	HydrateEntities bool `json:"hydrate_entities,omitempty"`
	// Whether to return only the most recent location for each unique entity.
	LatestOnly bool `json:"latest_only,omitempty"`
}
