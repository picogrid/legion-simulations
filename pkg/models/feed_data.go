package models

import (
	"encoding/json"
	"io"
	"time"

	"github.com/google/uuid"
)

// IngestFeedDataRequest represents the request to ingest new feed data
// @Description Request body for ingesting new feed data into the system.
// @name IngestFeedDataRequest
type IngestFeedDataRequest struct {
	// EntityID identifies the logical entity (e.g., device, sensor, user) that produced the data
	EntityID uuid.UUID `json:"entity_id" binding:"required" swaggertype:"string" format:"uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`

	// FeedDefinitionID links this data point to its corresponding definition
	FeedDefinitionID uuid.UUID `json:"feed_definition_id" binding:"required" swaggertype:"string" format:"uuid" example:"b2c3d4e5-f6a1-7890-1234-567890abcdef"`

	// RecordedAt is the time the event/measurement actually happened
	RecordedAt time.Time `json:"recorded_at" binding:"required" swaggertype:"string" format:"date-time" example:"2024-02-16T21:45:33Z"`

	// Payload stores the actual data, which could be telemetry, logs, etc
	Payload *json.RawMessage `json:"payload,omitempty" swaggertype:"object"`

	// FileContentBase64 should be used for binary or file data, typically when the FeedDefinition.Category is "FILE"
	FileContentBase64 *string `json:"file_content_base64,omitempty" swaggertype:"string" format:"byte" example:"aGVsbG8gd29ybGQ="`

	// FileContentType is the MIME type of the file content (e.g., "image/jpeg", "application/pdf")
	FileContentType *string `json:"file_content_type,omitempty" swaggertype:"string" example:"image/jpeg"`

	// BlobMetadata stores additional provider or user metadata for the blob (only for FILE type)
	BlobMetadata *json.RawMessage `json:"blob_metadata,omitempty" swaggertype:"object"`
}

// IngestFeedFileDataRequest represents the metadata part of a file ingestion request
// @Description Request metadata for ingesting new file-based feed data into the system via multipart/form-data.
// @name IngestFeedFileDataRequest
type IngestFeedFileDataRequest struct {
	// EntityID identifies the logical entity (e.g., device, sensor, user) that produced the data
	EntityID uuid.UUID `json:"entity_id" binding:"required" swaggertype:"string" format:"uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`

	// FeedDefinitionID links this data point to its corresponding definition (which should have Category="FILE")
	FeedDefinitionID uuid.UUID `json:"feed_definition_id" binding:"required" swaggertype:"string" format:"uuid" example:"b2c3d4e5-f6a1-7890-1234-567890abcdef"`

	// RecordedAt is the time the event/measurement actually happened
	RecordedAt time.Time `json:"recorded_at" binding:"required" swaggertype:"string" format:"date-time" example:"2024-02-16T21:45:33Z"`

	// FileContentType is the MIME type of the file content (e.g., "image/jpeg", "application/pdf")
	// This can be provided in the metadata or inferred from the file part.
	FileContentType *string `json:"file_content_type,omitempty" swaggertype:"string" example:"image/jpeg"`

	// BlobMetadata stores additional provider or user metadata for the blob
	BlobMetadata *json.RawMessage `json:"blob_metadata,omitempty" swaggertype:"object"`
}

// ServiceIngestFileRequest is the DTO used by the FeedDataIngestionService
// to accept file data, abstracting away transport-specific types.
// @Description Data required by the ingestion service for a single file upload, including content stream and metadata.
// @name ServiceIngestFileRequest
type ServiceIngestFileRequest struct {
	// EntityID identifies the logical entity
	EntityID uuid.UUID
	// FeedDefinitionID links this data point to its definition
	FeedDefinitionID uuid.UUID
	// RecordedAt is when the event actually happened
	RecordedAt time.Time `json:"recorded_at" binding:"required"`
	// FileContentType is the MIME type of the file
	FileContentType *string
	// BlobMetadata stores additional user metadata for the blob
	BlobMetadata *json.RawMessage

	// Filename is the original name of the file
	Filename string
	// FileSize is the size of the file in bytes
	FileSize int64
	// FileContent is the stream of the file's content. This should not be marshalled to JSON.
	FileContent io.Reader `json:"-"`
}

// ServiceIngestMessageRequest is the DTO used by the FeedDataIngestionService
// to accept message payload data, abstracting away transport-specific types.
// @Description Data required by the ingestion service for a single message payload.
// @name ServiceIngestMessageRequest
type ServiceIngestMessageRequest struct {
	// EntityID identifies the logical entity
	EntityID uuid.UUID `json:"entity_id" binding:"required"`
	// FeedDefinitionID links this data point to its definition
	FeedDefinitionID uuid.UUID `json:"feed_definition_id" binding:"required"`
	// RecordedAt is when the event actually happened
	RecordedAt time.Time `json:"recorded_at" binding:"required"`
	// Payload stores the actual data
	Payload *json.RawMessage `json:"payload" binding:"required" swaggertype:"object"`
	// Metadata stores additional optional metadata
	Metadata *json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
}

// FeedDataResponse represents the response for a feed data entry
// @Description Contains the complete details of a feed data entry, including timestamps and blob information.
// @name FeedDataResponse
type FeedDataResponse struct {
	// RecordedAt is the time the event/measurement actually happened
	RecordedAt time.Time `json:"recorded_at" swaggertype:"string" format:"date-time" example:"2024-02-16T21:45:33Z"`

	// ReceivedAt is the time the data was ingested by the platform
	ReceivedAt time.Time `json:"received_at" swaggertype:"string" format:"date-time" example:"2024-02-16T21:45:40Z"`

	// EntityID identifies the logical entity (e.g., device, sensor, user) that produced the data
	EntityID uuid.UUID `json:"entity_id" swaggertype:"string" format:"uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`

	// FeedDefinitionID links this data point to its corresponding definition
	FeedDefinitionID uuid.UUID `json:"feed_definition_id" swaggertype:"string" format:"uuid" example:"b2c3d4e5-f6a1-7890-1234-567890abcdef"`

	// Payload stores the actual data, which could be telemetry, logs, etc
	Payload *json.RawMessage `json:"payload,omitempty" swaggertype:"object"`

	// BlobStorageType indicates the type of blob storage used (e.g., 'S3', 'GCS')
	BlobStorageType *string `json:"blob_storage_type,omitempty" example:"S3"`

	// BlobURI is the path or full URI to the blob in external storage
	BlobURI *string `json:"blob_uri,omitempty" example:"s3://mybucket/path/to/file.jpg"`

	// BlobSizeBytes stores the size of the blob in bytes
	BlobSizeBytes *int64 `json:"blob_size_bytes,omitempty" example:"1048576"`

	// BlobContentType is the MIME type of the blob
	BlobContentType *string `json:"blob_content_type,omitempty" example:"image/jpeg"`

	// BlobMetadata stores additional provider or user metadata for the blob
	BlobMetadata *json.RawMessage `json:"blob_metadata,omitempty" swaggertype:"string" format:"json"`
}

// FeedDataListResponse represents a paginated list of feed data entries
// @Description A paginated response containing a list of feed data entries, total count, and pagination information.
// @name FeedDataListResponse
type FeedDataListResponse struct {
	// Results contains the list of feed data entries for the current page
	Results []FeedDataResponse `json:"results"`

	// TotalCount is the total number of feed data entries matching the query
	TotalCount int `json:"total_count" example:"42"`

	// Paging contains optional paging information
	Paging Paging `json:"paging,omitempty"`
}

// FeedDataPublishMessage represents the structure for publishing feed data.
// DataPayload and FileContentBase64 are mutually exclusive; this should be enforced by validation.
type FeedDataPublishMessage struct {
	// RecordedAt is the mandatory timestamp indicating when the event or measurement actually occurred
	RecordedAt time.Time `json:"recorded_at" binding:"required"`

	// EntityID is the mandatory UUID of the logical entity that produced the data
	EntityID uuid.UUID `json:"entity_id" binding:"required"`

	// FeedDefinitionID is the mandatory UUID linking this data point to its corresponding FeedDefinition
	FeedDefinitionID uuid.UUID `json:"feed_definition_id" binding:"required"`

	// DataPayload should be used for structured data, typically when the FeedDefinition.Category is "MESSAGE"
	DataPayload *json.RawMessage `json:"data_payload,omitempty" swaggertype:"object"`

	// FileContentBase64 should be used for binary or file data, typically when the FeedDefinition.Category is "FILE"
	FileContentBase64 *string `json:"file_content_base64,omitempty"`

	// FileContentType is the MIME type of the file content (e.g., "image/jpeg")
	FileContentType *string `json:"file_content_type,omitempty"`

	// Metadata can store additional provider or user-defined metadata for the feed item
	Metadata *json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
}

// FeedDataSearchRequest defines the parameters for searching feed data
// @Description Request parameters for searching feed data with various filters and sorting options.
// @name FeedDataSearchRequest
type FeedDataSearchRequest struct {
	// EntityIDs filters feed data by specific entity IDs
	EntityIDs []uuid.UUID `json:"entity_ids" swaggertype:"array,string" format:"uuid" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`

	// FeedDefinitionIDs filters feed data by specific feed definition IDs
	FeedDefinitionIDs []uuid.UUID `json:"feed_definition_ids" swaggertype:"array,string" format:"uuid" example:"b2c3d4e5-f6a1-7890-1234-567890abcdef"`

	// OccurredAfter filters feed data recorded after this timestamp
	OccurredAfter time.Time `json:"occurred_after" swaggertype:"string" format:"date-time" example:"2024-02-16T00:00:00Z"`

	// OccurredBefore filters feed data recorded before this timestamp
	OccurredBefore time.Time `json:"occurred_before" swaggertype:"string" format:"date-time" example:"2024-02-17T00:00:00Z"`

	// ReceivedAfter filters feed data received after this timestamp
	ReceivedAfter time.Time `json:"received_after" swaggertype:"string" format:"date-time" example:"2024-02-16T00:00:00Z"`

	// ReceivedBefore filters feed data received before this timestamp
	ReceivedBefore time.Time `json:"received_before" swaggertype:"string" format:"date-time" example:"2024-02-17T00:00:00Z"`

	// BlobStorageTypes filters feed data by blob storage types
	BlobStorageTypes []string `json:"blob_storage_types" swaggertype:"array,string" example:"S3,GCS"`

	// BlobContentTypes filters feed data by blob content types
	BlobContentTypes []string `json:"blob_content_types" swaggertype:"array,string" example:"image/jpeg,image/png"`

	// SortFields specifies the fields to sort by and the sort direction
	SortFields []SortField `json:"sort_fields" swaggertype:"array,object"`
}
