package simple

import (
	"fmt"
	"time"
)

// Config holds the configuration for the simple simulation
type Config struct {
	NumEntities    int
	EntityType     string
	UpdateInterval time.Duration
	Duration       time.Duration
	OrganizationID string
}

// ValidateAndParse validates and parses the raw parameters into a Config
func ValidateAndParse(params map[string]interface{}) (*Config, error) {
	config := &Config{}

	// Parse num_entities
	if v, ok := params["num_entities"]; ok {
		switch val := v.(type) {
		case int:
			config.NumEntities = val
		case float64:
			config.NumEntities = int(val)
		default:
			return nil, fmt.Errorf("num_entities must be an integer")
		}
	}
	if config.NumEntities < 1 || config.NumEntities > 5 {
		return nil, fmt.Errorf("num_entities must be between 1 and 5")
	}

	// Parse entity_type
	if v, ok := params["entity_type"]; ok {
		config.EntityType = fmt.Sprintf("%v", v)
	}
	validTypes := map[string]bool{"Camera": true, "Drone": true, "Sensor": true}
	if !validTypes[config.EntityType] {
		return nil, fmt.Errorf("entity_type must be one of: Camera, Drone, Sensor")
	}

	// Parse update_interval
	if v, ok := params["update_interval"]; ok {
		switch val := v.(type) {
		case float64:
			config.UpdateInterval = time.Duration(val * float64(time.Second))
		case int:
			config.UpdateInterval = time.Duration(val) * time.Second
		default:
			return nil, fmt.Errorf("update_interval must be a number")
		}
	}
	if config.UpdateInterval < time.Second || config.UpdateInterval > 60*time.Second {
		return nil, fmt.Errorf("update_interval must be between 1 and 60 seconds")
	}

	// Parse duration
	if v, ok := params["duration"]; ok {
		durationStr := fmt.Sprintf("%v", v)
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return nil, fmt.Errorf("invalid duration format: %w", err)
		}
		config.Duration = duration
	}

	// Parse organization_id
	if v, ok := params["organization_id"]; ok {
		config.OrganizationID = fmt.Sprintf("%v", v)
	}
	if config.OrganizationID == "" {
		return nil, fmt.Errorf("organization_id is required")
	}

	return config, nil
}
