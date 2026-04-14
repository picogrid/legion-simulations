package tracktraffic

import "testing"

func TestValidateAndParse(t *testing.T) {
	params := map[string]interface{}{
		"total_tracks":         42,
		"max_concurrency":      16,
		"update_interval":      2.0,
		"duration":             "5m",
		"center_lat":           37.7749,
		"center_lon":           -122.4194,
		"center_alt_m":         0.0,
		"grid_spacing_m":       700.0,
		"grid_jitter_m":        80.0,
		"history_points":       36,
		"history_step_seconds": 5.0,
		"delete_on_exit":       true,
		"organization_id":      "ecc2dce2-b664-4077-b34c-ea89e1fb045e",
	}

	cfg, err := ValidateAndParse(params)
	if err != nil {
		t.Fatalf("ValidateAndParse returned error: %v", err)
	}

	if cfg.TotalTracks != 42 {
		t.Fatalf("expected total_tracks=42, got %d", cfg.TotalTracks)
	}
	if cfg.MaxConcurrency != 16 {
		t.Fatalf("expected max_concurrency=16, got %d", cfg.MaxConcurrency)
	}
	if cfg.HistoryPoints != 36 {
		t.Fatalf("expected history_points=36, got %d", cfg.HistoryPoints)
	}
	if cfg.OrganizationID == "" {
		t.Fatal("expected organization ID to be set")
	}
	if !cfg.DeleteOnExit {
		t.Fatal("expected delete_on_exit=true")
	}
}
