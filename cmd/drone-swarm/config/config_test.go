package config

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Test loading the existing config.yaml file
	config, err := LoadConfig("../config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate basic simulation settings
	if config.Simulation.Name != "drone-swarm" {
		t.Errorf("Expected simulation name 'drone-swarm', got '%s'", config.Simulation.Name)
	}

	if config.Simulation.Description != "Counter-UAS vs Drone Swarm Engagement Simulation" {
		t.Errorf("Unexpected simulation description: %s", config.Simulation.Description)
	}

	if config.Simulation.UpdateInterval != 3*time.Second {
		t.Errorf("Expected update interval 3s, got %v", config.Simulation.UpdateInterval)
	}

	// Validate defaults
	if config.Defaults.NumCounterUASSystems != 5 {
		t.Errorf("Expected 5 Counter-UAS systems, got %d", config.Defaults.NumCounterUASSystems)
	}

	if config.Defaults.NumUASThreats != 20 {
		t.Errorf("Expected 20 UAS threats, got %d", config.Defaults.NumUASThreats)
	}

	if config.Defaults.EngagementTypeMix != 0.7 {
		t.Errorf("Expected engagement type mix 0.7, got %f", config.Defaults.EngagementTypeMix)
	}

	// Validate swarm config
	if config.SwarmConfig.FormationType != "distributed" {
		t.Errorf("Expected formation type 'distributed', got '%s'", config.SwarmConfig.FormationType)
	}

	if config.SwarmConfig.WaveDelay != 45*time.Second {
		t.Errorf("Expected wave delay 45s, got %v", config.SwarmConfig.WaveDelay)
	}

	if config.SwarmConfig.WaveCount != 3 {
		t.Errorf("Expected wave count 3, got %d", config.SwarmConfig.WaveCount)
	}

	if config.SwarmConfig.EvasionProbability != 0.7 {
		t.Errorf("Expected evasion probability 0.7, got %f", config.SwarmConfig.EvasionProbability)
	}

	// Validate defense config
	if config.DefenseConfig.PlacementPattern != "ring" {
		t.Errorf("Expected placement pattern 'ring', got '%s'", config.DefenseConfig.PlacementPattern)
	}

	if config.DefenseConfig.EngagementRules != "closest" {
		t.Errorf("Expected engagement rules 'closest', got '%s'", config.DefenseConfig.EngagementRules)
	}

	if config.DefenseConfig.KineticRatio != 0.7 {
		t.Errorf("Expected kinetic ratio 0.7, got %f", config.DefenseConfig.KineticRatio)
	}

	if config.DefenseConfig.DetectionRadiusKm != 10 {
		t.Errorf("Expected detection radius 10km, got %f", config.DefenseConfig.DetectionRadiusKm)
	}

	// Validate engagement config
	if config.Engagement.KineticSuccessRateRange.Min != 0.7 {
		t.Errorf("Expected kinetic success rate min 0.7, got %f", config.Engagement.KineticSuccessRateRange.Min)
	}

	if config.Engagement.KineticSuccessRateRange.Max != 0.9 {
		t.Errorf("Expected kinetic success rate max 0.9, got %f", config.Engagement.KineticSuccessRateRange.Max)
	}

	if config.Engagement.KineticAmmoCapacity != 5 {
		t.Errorf("Expected kinetic ammo capacity 5, got %d", config.Engagement.KineticAmmoCapacity)
	}

	// Validate termination conditions
	expectedSuccessConditions := []string{"all_threats_neutralized"}
	if len(config.Termination.SuccessConditions) != len(expectedSuccessConditions) {
		t.Errorf("Expected %d success conditions, got %d", len(expectedSuccessConditions), len(config.Termination.SuccessConditions))
	}

	expectedFailureConditions := []string{"defensive_breach"}
	if len(config.Termination.FailureConditions) != len(expectedFailureConditions) {
		t.Errorf("Expected %d failure conditions, got %d", len(expectedFailureConditions), len(config.Termination.FailureConditions))
	}
}

func TestDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	// Test validation
	if err := config.Validate(); err != nil {
		t.Fatalf("Default config validation failed: %v", err)
	}

	// Ensure default config matches expected values
	if config.Simulation.Name != "drone-swarm" {
		t.Errorf("Expected default simulation name 'drone-swarm', got '%s'", config.Simulation.Name)
	}

	if config.Defaults.NumCounterUASSystems <= 0 {
		t.Errorf("Default config must have positive number of Counter-UAS systems")
	}

	if config.Defaults.NumUASThreats <= 0 {
		t.Errorf("Default config must have positive number of UAS threats")
	}
}

func TestConfigValidation(t *testing.T) {
	// Test invalid configurations
	tests := []struct {
		name   string
		config *SimulationConfig
		hasErr bool
	}{
		{
			name:   "empty name",
			config: &SimulationConfig{},
			hasErr: true,
		},
		{
			name: "negative update interval",
			config: &SimulationConfig{
				Simulation: SimulationSettings{
					Name:           "test",
					UpdateInterval: -1 * time.Second,
				},
			},
			hasErr: true,
		},
		{
			name: "zero entities",
			config: &SimulationConfig{
				Simulation: SimulationSettings{
					Name:           "test",
					UpdateInterval: 1 * time.Second,
				},
				Defaults: DefaultsConfig{
					NumCounterUASSystems: 0,
					NumUASThreats:        10,
				},
			},
			hasErr: true,
		},
		{
			name: "invalid evasion probability",
			config: &SimulationConfig{
				Simulation: SimulationSettings{
					Name:           "test",
					UpdateInterval: 1 * time.Second,
				},
				Defaults: DefaultsConfig{
					NumCounterUASSystems: 5,
					NumUASThreats:        10,
				},
				SwarmConfig: SwarmConfig{
					EvasionProbability: 1.5, // Invalid: > 1.0
				},
			},
			hasErr: true,
		},
		{
			name:   "valid config",
			config: GetDefaultConfig(),
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.hasErr && err == nil {
				t.Errorf("Expected validation error for %s", tt.name)
			}
			if !tt.hasErr && err != nil {
				t.Errorf("Unexpected validation error for %s: %v", tt.name, err)
			}
		})
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	// Test environment variable overrides
	config := GetDefaultConfig()
	originalSystems := config.Defaults.NumCounterUASSystems
	originalThreats := config.Defaults.NumUASThreats

	// Set environment variables
	t.Setenv("NUM_COUNTER_UAS_SYSTEMS", "10")
	t.Setenv("NUM_UAS_THREATS", "30")
	t.Setenv("ENGAGEMENT_TYPE_MIX", "0.8")
	t.Setenv("LOG_LEVEL", "debug")

	// Apply environment overrides
	MergeWithEnvironment(config)

	// Check that values were overridden
	if config.Defaults.NumCounterUASSystems == originalSystems {
		t.Errorf("Environment override for NUM_COUNTER_UAS_SYSTEMS failed")
	}

	if config.Defaults.NumUASThreats == originalThreats {
		t.Errorf("Environment override for NUM_UAS_THREATS failed")
	}

	if config.Defaults.EngagementTypeMix != 0.8 {
		t.Errorf("Expected engagement type mix 0.8, got %f", config.Defaults.EngagementTypeMix)
	}

	if config.Logging.ConsoleLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.Logging.ConsoleLevel)
	}
}

func TestCLIOverrides(t *testing.T) {
	config := GetDefaultConfig()

	overrides := map[string]interface{}{
		"num_counter_uas_systems": 15,
		"num_uas_threats":         25,
		"formation_type":          "waves",
		"placement_pattern":       "cluster",
		"verbose_logging":         true,
	}

	MergeWithCLIOverrides(config, overrides)

	if config.Defaults.NumCounterUASSystems != 15 {
		t.Errorf("Expected 15 Counter-UAS systems, got %d", config.Defaults.NumCounterUASSystems)
	}

	if config.Defaults.NumUASThreats != 25 {
		t.Errorf("Expected 25 UAS threats, got %d", config.Defaults.NumUASThreats)
	}

	if config.SwarmConfig.FormationType != "waves" {
		t.Errorf("Expected formation type 'waves', got '%s'", config.SwarmConfig.FormationType)
	}

	if config.DefenseConfig.PlacementPattern != "cluster" {
		t.Errorf("Expected placement pattern 'cluster', got '%s'", config.DefenseConfig.PlacementPattern)
	}

	if !config.Advanced.VerboseLogging {
		t.Errorf("Expected verbose logging to be enabled")
	}
}
