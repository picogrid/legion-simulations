package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Affiliation string

const (
	AffiliationPENDING               Affiliation = "PENDING"
	AffiliationUNKNOWN               Affiliation = "UNKNOWN"
	AffiliationASSUMEDFRIEND         Affiliation = "ASSUMED_FRIEND"
	AffiliationFRIEND                Affiliation = "FRIEND"
	AffiliationNEUTRAL               Affiliation = "NEUTRAL"
	AffiliationSUSPECT               Affiliation = "SUSPECT"
	AffiliationHOSTILE               Affiliation = "HOSTILE"
	AffiliationEXERCISEPENDING       Affiliation = "EXERCISE_PENDING"
	AffiliationEXERCISEUNKNOWN       Affiliation = "EXERCISE_UNKNOWN"
	AffiliationEXERCISEFRIEND        Affiliation = "EXERCISE_FRIEND"
	AffiliationEXERCISENEUTRAL       Affiliation = "EXERCISE_NEUTRAL"
	AffiliationEXERCISEASSUMEDFRIEND Affiliation = "EXERCISE_ASSUMED_FRIEND"
	AffiliationJOKER                 Affiliation = "JOKER"
	AffiliationFAKER                 Affiliation = "FAKER"
	AffiliationNONESPECIFIED         Affiliation = "NONE_SPECIFIED"
)

type Category string

const (
	CategoryDEVICE    Category = "DEVICE"
	CategoryDETECTION Category = "DETECTION"
	CategoryALERT     Category = "ALERT"
	CategoryWEATHER   Category = "WEATHER"
	CategoryGEOMETRIC Category = "GEOMETRIC"
	CategoryZONE      Category = "ZONE"
	CategorySENSOR    Category = "SENSOR"
	CategoryVEHICLE   Category = "VEHICLE"
	CategoryUXV       Category = "UXV"
	CategoryTRACK     Category = "TRACK"
)

type MessageCategory string

const (
	MessageCategoryMESSAGE MessageCategory = "MESSAGE"
	MessageCategoryFILE    MessageCategory = "FILE"
)

type GeomPoint struct {
	Type        *string   `json:"type,omitempty"`
	Coordinates []float64 `json:"coordinates"`
}

type CreateEntityRequest struct {
	Affiliation    Affiliation      `json:"affiliation,omitempty"`
	Category       *Category        `json:"category,omitempty"`
	Classification *json.RawMessage `json:"classification,omitempty"`
	Metadata       *json.RawMessage `json:"metadata,omitempty"`
	Name           *string          `json:"name,omitempty"`
	OrganizationID *uuid.UUID       `json:"organization_id,omitempty"`
	ParentID       *uuid.UUID       `json:"parent_id,omitempty"`
	Status         *string          `json:"status,omitempty"`
	Type           *string          `json:"type,omitempty"`
}

type UpdateEntityRequest struct {
	ID             uuid.UUID        `json:"id,omitempty"`
	Affiliation    Affiliation      `json:"affiliation,omitempty"`
	Category       Category         `json:"category,omitempty"`
	Classification *json.RawMessage `json:"classification,omitempty"`
	Metadata       *json.RawMessage `json:"metadata,omitempty"`
	Name           *string          `json:"name,omitempty"`
	ParentID       *uuid.UUID       `json:"parent_id,omitempty"`
	Status         string           `json:"status,omitempty"`
	Type           *string          `json:"type,omitempty"`
}

type EntityResponse struct {
	ID                           uuid.UUID        `json:"id"`
	OrganizationID               uuid.UUID        `json:"organization_id"`
	Name                         string           `json:"name"`
	Category                     Category         `json:"category"`
	Type                         string           `json:"type"`
	Status                       string           `json:"status"`
	Affiliation                  Affiliation      `json:"affiliation"`
	ParentID                     *uuid.UUID       `json:"parent_id"`
	Metadata                     *json.RawMessage `json:"metadata,omitempty"`
	Classification               *json.RawMessage `json:"classification,omitempty"`
	TopClassification            *string          `json:"top_classification,omitempty"`
	TopClassificationProbability *float32         `json:"top_classification_probability,omitempty"`
	CreatedAt                    time.Time        `json:"created_at"`
	UpdatedAt                    time.Time        `json:"updated_at"`
	DeletedAt                    *time.Time       `json:"deleted_at,omitempty"`
}

type SearchFilters struct {
	Affiliation        []Affiliation `json:"affiliation,omitempty"`
	Category           []Category    `json:"category,omitempty"`
	CreatedAfter       *time.Time    `json:"created_after,omitempty"`
	CreatedBefore      *time.Time    `json:"created_before,omitempty"`
	EntityIDs          []uuid.UUID   `json:"entity_ids,omitempty"`
	Name               string        `json:"name,omitempty"`
	ParentIDs          []uuid.UUID   `json:"parent_ids,omitempty"`
	Status             []string      `json:"status,omitempty"`
	TopClassifications []string      `json:"top_classifications,omitempty"`
	Type               string        `json:"type,omitempty"`
	Types              []string      `json:"types,omitempty"`
	UpdatedAfter       *time.Time    `json:"updated_after,omitempty"`
	UpdatedBefore      *time.Time    `json:"updated_before,omitempty"`
}

type SearchEntitiesRequest struct {
	OrganizationID *uuid.UUID      `json:"organization_id,omitempty"`
	Filters        *SearchFilters  `json:"filters,omitempty"`
	Sort           []SortFieldSpec `json:"sort,omitempty"`
}

type SortFieldSpec struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

type EntityPaginatedResponse = PaginatedResponse[EntityResponse]

type CreateEntityLocationRequest struct {
	Position        *GeomPoint `json:"position,omitempty"`
	Source          string     `json:"source"`
	RecordedAt      *time.Time `json:"recorded_at,omitempty"`
	Acceleration    []float64  `json:"acceleration,omitempty"`
	AngularVelocity []float64  `json:"angular_velocity,omitempty"`
	Bearing         *float64   `json:"bearing,omitempty"`
	Orientation     []float64  `json:"orientation,omitempty"`
	Radius          *float64   `json:"radius,omitempty"`
	Velocity        []float64  `json:"velocity,omitempty"`
}

type EntityLocationResponse struct {
	ID              uuid.UUID       `json:"id"`
	EntityID        uuid.UUID       `json:"entity_id"`
	Entity          *EntityResponse `json:"entity,omitempty"`
	Position        GeomPoint       `json:"position"`
	Source          string          `json:"source"`
	RecordedAt      *time.Time      `json:"recorded_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	Acceleration    []float64       `json:"acceleration,omitempty"`
	AngularVelocity []float64       `json:"angular_velocity,omitempty"`
	Bearing         *float64        `json:"bearing,omitempty"`
	Orientation     []float64       `json:"orientation,omitempty"`
	Radius          *float64        `json:"radius,omitempty"`
	Speed           *float64        `json:"speed,omitempty"`
	Velocity        []float64       `json:"velocity,omitempty"`
}

type SearchEntityLocationsRequest struct {
	EntityIDs      []uuid.UUID `json:"entity_ids,omitempty"`
	RecordedAfter  *time.Time  `json:"recorded_after,omitempty"`
	RecordedBefore *time.Time  `json:"recorded_before,omitempty"`
	Sources        []string    `json:"sources,omitempty"`
}

type EntityLocationPaginatedResponse = PaginatedResponse[EntityLocationResponse]

type CreateFeedDefinitionRequest struct {
	Category         *MessageCategory `json:"category,omitempty"`
	DataType         *string          `json:"data_type,omitempty"`
	Description      string           `json:"description,omitempty"`
	EntityID         uuid.UUID        `json:"entity_id,omitempty"`
	FeedName         *string          `json:"feed_name,omitempty"`
	IsActive         *bool            `json:"is_active,omitempty"`
	IsTemplate       *bool            `json:"is_template,omitempty"`
	IntegrationID    *uuid.UUID       `json:"integration_id,omitempty"`
	Metadata         *json.RawMessage `json:"metadata,omitempty"`
	SchemaDefinition *json.RawMessage `json:"schema_definition,omitempty"`
	TemplateID       *uuid.UUID       `json:"template_id,omitempty"`
}

type UpdateFeedDefinitionRequest struct {
	Category         MessageCategory  `json:"category,omitempty"`
	DataType         *string          `json:"data_type,omitempty"`
	Description      *string          `json:"description,omitempty"`
	EntityID         *uuid.UUID       `json:"entity_id,omitempty"`
	FeedName         *string          `json:"feed_name,omitempty"`
	IsActive         *bool            `json:"is_active,omitempty"`
	IsTemplate       *bool            `json:"is_template,omitempty"`
	IntegrationID    *uuid.UUID       `json:"integration_id,omitempty"`
	Metadata         *json.RawMessage `json:"metadata,omitempty"`
	SchemaDefinition *json.RawMessage `json:"schema_definition,omitempty"`
	TemplateID       *uuid.UUID       `json:"template_id,omitempty"`
}

type FeedDefinitionSearchRequest struct {
	Category       MessageCategory `json:"category,omitempty"`
	CreatedAfter   *time.Time      `json:"created_after,omitempty"`
	CreatedBefore  *time.Time      `json:"created_before,omitempty"`
	DataType       *string         `json:"data_type,omitempty"`
	EntityID       uuid.UUID       `json:"entity_id,omitempty"`
	FeedName       *string         `json:"feed_name,omitempty"`
	IntegrationID  *uuid.UUID      `json:"integration_id,omitempty"`
	IsActive       *bool           `json:"is_active,omitempty"`
	IsTemplate     *bool           `json:"is_template,omitempty"`
	OrganizationID *uuid.UUID      `json:"organization_id,omitempty"`
	TemplateID     *uuid.UUID      `json:"template_id,omitempty"`
}

type FeedDefinitionResponse struct {
	ID               uuid.UUID        `json:"id"`
	OrganizationID   uuid.UUID        `json:"organization_id"`
	Category         MessageCategory  `json:"category"`
	DataType         string           `json:"data_type"`
	Description      *string          `json:"description,omitempty"`
	EntityID         uuid.UUID        `json:"entity_id,omitempty"`
	FeedName         string           `json:"feed_name"`
	IntegrationID    *uuid.UUID       `json:"integration_id,omitempty"`
	IsActive         bool             `json:"is_active"`
	IsTemplate       bool             `json:"is_template"`
	Metadata         *json.RawMessage `json:"metadata,omitempty"`
	SchemaDefinition *json.RawMessage `json:"schema_definition,omitempty"`
	TemplateID       *uuid.UUID       `json:"template_id,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type FeedDefinitionListResponse = PaginatedResponse[FeedDefinitionResponse]

type FeedDataSearchRequest struct {
	FeedID    uuid.UUID  `json:"feed_id,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	EntityID  *uuid.UUID `json:"entity_id,omitempty"`
}

type FeedDataResponse struct {
	ID               uuid.UUID        `json:"id"`
	OrganizationID   uuid.UUID        `json:"organization_id"`
	EntityID         uuid.UUID        `json:"entity_id"`
	FeedDefinitionID uuid.UUID        `json:"feed_definition_id"`
	Payload          *json.RawMessage `json:"payload,omitempty"`
	BlobContentType  *string          `json:"blob_content_type,omitempty"`
	BlobKey          *string          `json:"blob_key,omitempty"`
	BlobMetadata     *json.RawMessage `json:"blob_metadata,omitempty"`
	BlobSizeBytes    *float32         `json:"blob_size_bytes,omitempty"`
	BlobStorageType  *string          `json:"blob_storage_type,omitempty"`
	RecordedAt       time.Time        `json:"recorded_at"`
	ReceivedAt       time.Time        `json:"received_at"`
}

type FeedDataListResponse = PaginatedResponse[FeedDataResponse]

type ServiceIngestMessageRequest struct {
	EntityID         *uuid.UUID       `json:"entity_id,omitempty"`
	FeedDefinitionID *uuid.UUID       `json:"feed_definition_id,omitempty"`
	Metadata         *json.RawMessage `json:"metadata,omitempty"`
	Payload          *json.RawMessage `json:"payload,omitempty"`
	RecordedAt       *time.Time       `json:"recorded_at,omitempty"`
}

type IngestFeedDataRequest struct {
	EntityID         *uuid.UUID       `json:"entity_id,omitempty"`
	FeedDefinitionID *uuid.UUID       `json:"feed_definition_id,omitempty"`
	Metadata         *json.RawMessage `json:"metadata,omitempty"`
	Payload          *json.RawMessage `json:"payload,omitempty"`
	RecordedAt       *time.Time       `json:"recorded_at,omitempty"`
}
