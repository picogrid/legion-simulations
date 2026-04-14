package client

import (
	"encoding/json"
	"fmt"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/pkg/models"
)

func toOpenapiUUID(id uuid.UUID) openapi_types.UUID {
	return openapi_types.UUID(id)
}

func toOpenapiUUIDPtr(id *uuid.UUID) *openapi_types.UUID {
	if id == nil {
		return nil
	}

	converted := toOpenapiUUID(*id)
	return &converted
}

func fromOpenapiUUID(id openapi_types.UUID) uuid.UUID {
	return uuid.UUID(id)
}

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q: %w", value, err)
	}

	return parsed, nil
}

func parseOptionalTime(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	parsed, err := parseTime(*value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func formatOptionalTime(value *time.Time) *string {
	if value == nil || value.IsZero() {
		return nil
	}

	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

func intPtrFromFloat32(value *float32) *int {
	if value == nil {
		return nil
	}

	v := int(*value)
	return &v
}

func rawMessageToMap(payload *json.RawMessage) (*map[string]interface{}, error) {
	if payload == nil || len(*payload) == 0 {
		return nil, nil
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(*payload, &decoded); err != nil {
		return nil, err
	}

	return &decoded, nil
}

func rawMessageToMapValue(payload *json.RawMessage) (map[string]interface{}, error) {
	if payload == nil || len(*payload) == 0 {
		return map[string]interface{}{}, nil
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(*payload, &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}

func mapToRawMessage(payload *map[string]interface{}) (*json.RawMessage, error) {
	if payload == nil {
		return nil, nil
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	raw := json.RawMessage(encoded)
	return &raw, nil
}

func float32SliceTo64(values []float32) []float64 {
	if len(values) == 0 {
		return nil
	}

	converted := make([]float64, len(values))
	for i, value := range values {
		converted[i] = float64(value)
	}
	return converted
}

func float64SliceTo32(values []float64) []float32 {
	if len(values) == 0 {
		return nil
	}

	converted := make([]float32, len(values))
	for i, value := range values {
		converted[i] = float32(value)
	}
	return converted
}

func float32MatrixTo64(values [][]float32) [][]float64 {
	if len(values) == 0 {
		return nil
	}

	converted := make([][]float64, len(values))
	for i, row := range values {
		converted[i] = float32SliceTo64(row)
	}
	return converted
}

func float64MatrixTo32(values [][]float64) [][]float32 {
	if len(values) == 0 {
		return nil
	}

	converted := make([][]float32, len(values))
	for i, row := range values {
		converted[i] = float64SliceTo32(row)
	}
	return converted
}

func float32PtrTo64(value *float32) *float64 {
	if value == nil {
		return nil
	}

	converted := float64(*value)
	return &converted
}

func float64PtrTo32(value *float64) *float32 {
	if value == nil {
		return nil
	}

	converted := float32(*value)
	return &converted
}

func toPaging(next *float32, previous *float32) models.Paging {
	return models.Paging{
		Next:     intPtrFromFloat32(next),
		Previous: intPtrFromFloat32(previous),
	}
}
