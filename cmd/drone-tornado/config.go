package dronetornado

import (
	"fmt"
	"time"
)

// Config holds the configuration for the Drone Tornado simulation
type Config struct {
	NumDrones       int
	UpdateInterval  time.Duration
	Duration        time.Duration
	RadiusMeters    float64
	RadiusOffsetM   float64
	SpeedMetersPerS float64
	CenterLat       float64
	CenterLon       float64
	CenterAltMeters float64
	OrganizationID  string
	CleanupExisting bool
	DeleteOnExit    bool
}

// ValidateAndParse validates and parses raw parameters into a Config
func ValidateAndParse(params map[string]interface{}) (*Config, error) {
	cfg := &Config{}

	// number_of_drones
	if v, ok := params["number_of_drones"]; ok {
		switch val := v.(type) {
		case int:
			cfg.NumDrones = val
		case float64:
			cfg.NumDrones = int(val)
		default:
			return nil, fmt.Errorf("number_of_drones must be an integer")
		}
	}
	if cfg.NumDrones < 1 {
		return nil, fmt.Errorf("number_of_drones must be at least 1")
	}

	// update_interval (seconds as float)
	if v, ok := params["update_interval"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.UpdateInterval = time.Duration(val * float64(time.Second))
		case int:
			cfg.UpdateInterval = time.Duration(val) * time.Second
		default:
			return nil, fmt.Errorf("update_interval must be a number (seconds)")
		}
	}
	if cfg.UpdateInterval <= 0 {
		return nil, fmt.Errorf("update_interval must be greater than 0 seconds")
	}

	// duration (Go duration string)
	if v, ok := params["duration"]; ok {
		durationStr := fmt.Sprintf("%v", v)
		d, err := time.ParseDuration(durationStr)
		if err != nil {
			return nil, fmt.Errorf("invalid duration format: %w", err)
		}
		cfg.Duration = d
	}
	if cfg.Duration <= 0 {
		return nil, fmt.Errorf("duration must be greater than 0")
	}

	// radius_m
	if v, ok := params["radius_m"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.RadiusMeters = val
		case int:
			cfg.RadiusMeters = float64(val)
		default:
			return nil, fmt.Errorf("radius_m must be a number (meters)")
		}
	}
	if cfg.RadiusMeters <= 0 {
		return nil, fmt.Errorf("radius_m must be greater than 0")
	}

	// radius_offset_m
	if v, ok := params["radius_offset_m"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.RadiusOffsetM = val
		case int:
			cfg.RadiusOffsetM = float64(val)
		default:
			return nil, fmt.Errorf("radius_offset_m must be a number (meters)")
		}
	} else {
		// default if not provided
		cfg.RadiusOffsetM = 10.0
	}
	if cfg.RadiusOffsetM < 0 {
		return nil, fmt.Errorf("radius_offset_m must be >= 0")
	}

	// speed_mps
	if v, ok := params["speed_mps"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.SpeedMetersPerS = val
		case int:
			cfg.SpeedMetersPerS = float64(val)
		default:
			return nil, fmt.Errorf("speed_mps must be a number (m/s)")
		}
	}
	if cfg.SpeedMetersPerS <= 0 {
		return nil, fmt.Errorf("speed_mps must be greater than 0")
	}

	// center_lat
	if v, ok := params["center_lat"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.CenterLat = val
		case int:
			cfg.CenterLat = float64(val)
		default:
			return nil, fmt.Errorf("center_lat must be a number")
		}
	}

	// center_lon
	if v, ok := params["center_lon"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.CenterLon = val
		case int:
			cfg.CenterLon = float64(val)
		default:
			return nil, fmt.Errorf("center_lon must be a number")
		}
	}

	// center_alt_m
	if v, ok := params["center_alt_m"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.CenterAltMeters = val
		case int:
			cfg.CenterAltMeters = float64(val)
		default:
			return nil, fmt.Errorf("center_alt_m must be a number (meters)")
		}
	}

	// organization_id
	if v, ok := params["organization_id"]; ok {
		cfg.OrganizationID = fmt.Sprintf("%v", v)
	}
	if cfg.OrganizationID == "" {
		return nil, fmt.Errorf("organization_id is required")
	}

	// cleanup_existing
	if v, ok := params["cleanup_existing"]; ok {
		switch val := v.(type) {
		case bool:
			cfg.CleanupExisting = val
		case string:
			cfg.CleanupExisting = val == "true" || val == "1" || val == "yes"
		default:
			return nil, fmt.Errorf("cleanup_existing must be a boolean")
		}
	}

	// delete_on_exit
	if v, ok := params["delete_on_exit"]; ok {
		switch val := v.(type) {
		case bool:
			cfg.DeleteOnExit = val
		case string:
			cfg.DeleteOnExit = val == "true" || val == "1" || val == "yes"
		default:
			return nil, fmt.Errorf("delete_on_exit must be a boolean")
		}
	}

	return cfg, nil
}
