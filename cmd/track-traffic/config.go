package tracktraffic

import (
	"fmt"
	"time"
)

// Config holds the configuration for the track traffic simulation.
type Config struct {
	TotalTracks     int
	MaxConcurrency  int
	UpdateInterval  time.Duration
	Duration        time.Duration
	CenterLat       float64
	CenterLon       float64
	CenterAltMeters float64
	GridSpacingM    float64
	GridJitterM     float64
	HistoryPoints   int
	HistoryStep     time.Duration
	DeleteOnExit    bool
	OrganizationID  string
}

// ValidateAndParse validates and parses raw parameters into a Config.
func ValidateAndParse(params map[string]interface{}) (*Config, error) {
	cfg := &Config{
		DeleteOnExit: true,
	}

	if v, ok := params["total_tracks"]; ok {
		switch val := v.(type) {
		case int:
			cfg.TotalTracks = val
		case float64:
			cfg.TotalTracks = int(val)
		default:
			return nil, fmt.Errorf("total_tracks must be an integer")
		}
	}
	if cfg.TotalTracks < 1 {
		return nil, fmt.Errorf("total_tracks must be at least 1")
	}

	if v, ok := params["max_concurrency"]; ok {
		switch val := v.(type) {
		case int:
			cfg.MaxConcurrency = val
		case float64:
			cfg.MaxConcurrency = int(val)
		default:
			return nil, fmt.Errorf("max_concurrency must be an integer")
		}
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = minInt(maxInt(cfg.TotalTracks/10, 8), 64)
	}
	if cfg.MaxConcurrency > cfg.TotalTracks {
		cfg.MaxConcurrency = cfg.TotalTracks
	}

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
	if cfg.UpdateInterval < time.Second {
		return nil, fmt.Errorf("update_interval must be at least 1 second")
	}

	if v, ok := params["duration"]; ok {
		durationStr := fmt.Sprintf("%v", v)
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return nil, fmt.Errorf("invalid duration format: %w", err)
		}
		cfg.Duration = duration
	}
	if cfg.Duration <= 0 {
		return nil, fmt.Errorf("duration must be greater than 0")
	}

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

	if v, ok := params["grid_spacing_m"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.GridSpacingM = val
		case int:
			cfg.GridSpacingM = float64(val)
		default:
			return nil, fmt.Errorf("grid_spacing_m must be a number (meters)")
		}
	}
	if cfg.GridSpacingM < 100 {
		return nil, fmt.Errorf("grid_spacing_m must be at least 100 meters")
	}

	if v, ok := params["grid_jitter_m"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.GridJitterM = val
		case int:
			cfg.GridJitterM = float64(val)
		default:
			return nil, fmt.Errorf("grid_jitter_m must be a number (meters)")
		}
	}
	if cfg.GridJitterM < 0 {
		return nil, fmt.Errorf("grid_jitter_m must be greater than or equal to 0")
	}
	if cfg.GridJitterM > cfg.GridSpacingM*0.35 {
		return nil, fmt.Errorf("grid_jitter_m must be no more than 35%% of grid_spacing_m to keep tracks separated")
	}

	if v, ok := params["history_points"]; ok {
		switch val := v.(type) {
		case int:
			cfg.HistoryPoints = val
		case float64:
			cfg.HistoryPoints = int(val)
		default:
			return nil, fmt.Errorf("history_points must be an integer")
		}
	}
	if cfg.HistoryPoints < 2 {
		return nil, fmt.Errorf("history_points must be at least 2")
	}

	if v, ok := params["history_step_seconds"]; ok {
		switch val := v.(type) {
		case float64:
			cfg.HistoryStep = time.Duration(val * float64(time.Second))
		case int:
			cfg.HistoryStep = time.Duration(val) * time.Second
		default:
			return nil, fmt.Errorf("history_step_seconds must be a number (seconds)")
		}
	}
	if cfg.HistoryStep < time.Second {
		return nil, fmt.Errorf("history_step_seconds must be at least 1 second")
	}

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

	if v, ok := params["organization_id"]; ok {
		cfg.OrganizationID = fmt.Sprintf("%v", v)
	}
	if cfg.OrganizationID == "" {
		return nil, fmt.Errorf("organization_id is required")
	}

	return cfg, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
