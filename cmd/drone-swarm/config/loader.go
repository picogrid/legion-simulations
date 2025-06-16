package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*SimulationConfig, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML
	var config SimulationConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadConfigOrDefault loads config from file or returns default, with environment overrides
func LoadConfigOrDefault(path string) (*SimulationConfig, error) {
	var config *SimulationConfig
	var err error

	if path != "" {
		config, err = LoadConfig(path)
		if err != nil {
			// Log error but continue with default
			fmt.Printf("Warning: Could not load config from %s: %v\n", path, err)
			config = nil
		}
	}

	// Try default locations if no config loaded yet
	if config == nil {
		defaultPaths := []string{
			"config.yaml",
			"drone-swarm.yaml",
			filepath.Join("cmd", "drone-swarm", "config.yaml"),
			filepath.Join(".", "config.yaml"),
		}

		for _, p := range defaultPaths {
			if _, err := os.Stat(p); err == nil {
				config, err = LoadConfig(p)
				if err == nil {
					fmt.Printf("Loaded config from: %s\n", p)
					break
				}
			}
		}
	}

	// Use default config if still no config loaded
	if config == nil {
		fmt.Println("Using default configuration")
		config = GetDefaultConfig()
	}

	// Always apply environment variable overrides
	MergeWithEnvironment(config)

	return config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *SimulationConfig, path string) error {
	// Validate before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// MergeWithCLIOverrides applies CLI parameter overrides to the configuration
func MergeWithCLIOverrides(config *SimulationConfig, overrides map[string]interface{}) {
	for key, value := range overrides {
		switch key {
		case "num_counter_uas_systems":
			if count, ok := value.(int); ok && count > 0 {
				config.Defaults.NumCounterUASSystems = count
			}
		case "num_uas_threats":
			if count, ok := value.(int); ok && count > 0 {
				config.Defaults.NumUASThreats = count
			}
		case "engagement_type_mix":
			if ratio, ok := value.(float64); ok && ratio >= 0 && ratio <= 1 {
				config.Defaults.EngagementTypeMix = ratio
				config.DefenseConfig.KineticRatio = ratio
			}
		case "center_latitude":
			if lat, ok := value.(float64); ok {
				config.Defaults.CenterLocation.Latitude = lat
			}
		case "center_longitude":
			if lon, ok := value.(float64); ok {
				config.Defaults.CenterLocation.Longitude = lon
			}
		case "center_altitude":
			if alt, ok := value.(float64); ok {
				config.Defaults.CenterLocation.Altitude = alt
			}
		case "formation_type":
			if formation, ok := value.(string); ok {
				validFormations := []string{"distributed", "concentrated", "waves"}
				for _, valid := range validFormations {
					if formation == valid {
						config.SwarmConfig.FormationType = formation
						break
					}
				}
			}
		case "placement_pattern":
			if placement, ok := value.(string); ok {
				validPlacements := []string{"ring", "cluster", "line"}
				for _, valid := range validPlacements {
					if placement == valid {
						config.DefenseConfig.PlacementPattern = placement
						break
					}
				}
			}
		case "wave_count":
			if count, ok := value.(int); ok && count > 0 {
				config.SwarmConfig.WaveCount = count
			}
		case "wave_delay":
			if duration, ok := value.(time.Duration); ok && duration > 0 {
				config.SwarmConfig.WaveDelay = duration
			}
		case "autonomy_distribution":
			if autonomy, ok := value.(string); ok {
				validAutonomy := []string{"low", "mixed", "high"}
				for _, valid := range validAutonomy {
					if autonomy == valid {
						config.SwarmConfig.AutonomyDistribution = autonomy
						break
					}
				}
			}
		case "evasion_probability":
			if prob, ok := value.(float64); ok && prob >= 0 && prob <= 1 {
				config.SwarmConfig.EvasionProbability = prob
			}
		case "success_rate_modifier":
			if modifier, ok := value.(float64); ok && modifier > 0 {
				config.DefenseConfig.SuccessRateModifier = modifier
			}
		case "verbose_logging":
			if verbose, ok := value.(bool); ok {
				config.Advanced.VerboseLogging = verbose
			}
		case "enable_aar":
			if enable, ok := value.(bool); ok {
				config.Logging.EnableAAR = enable
			}
		case "log_level":
			if level, ok := value.(string); ok {
				validLevels := []string{"debug", "info", "warn", "error"}
				for _, valid := range validLevels {
					if level == valid {
						config.Logging.ConsoleLevel = level
						break
					}
				}
			}
		}
	}
}

// LoadConfigWithOverrides loads config and applies both environment and CLI overrides
func LoadConfigWithOverrides(path string, cliOverrides map[string]interface{}) (*SimulationConfig, error) {
	config, err := LoadConfigOrDefault(path)
	if err != nil {
		return nil, err
	}

	// Apply CLI overrides after environment variables
	if cliOverrides != nil {
		MergeWithCLIOverrides(config, cliOverrides)
	}

	// Final validation
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed after overrides: %w", err)
	}

	return config, nil
}

// MergeWithEnvironment merges config with environment variables
func MergeWithEnvironment(config *SimulationConfig) {
	// Override organization ID if set
	if orgID := os.Getenv("LEGION_ORGANIZATION_ID"); orgID != "" {
		config.OrganizationID = orgID
	}

	// Override simulation parameters
	if updateInterval := os.Getenv("SIMULATION_UPDATE_INTERVAL"); updateInterval != "" {
		if duration, err := time.ParseDuration(updateInterval); err == nil {
			config.Simulation.UpdateInterval = duration
		}
	}

	// Override entity counts
	if numDefense := os.Getenv("NUM_COUNTER_UAS_SYSTEMS"); numDefense != "" {
		if count, err := strconv.Atoi(numDefense); err == nil && count > 0 {
			config.Defaults.NumCounterUASSystems = count
		}
	}

	if numThreats := os.Getenv("NUM_UAS_THREATS"); numThreats != "" {
		if count, err := strconv.Atoi(numThreats); err == nil && count > 0 {
			config.Defaults.NumUASThreats = count
		}
	}

	// Override location if set
	if lat := os.Getenv("CENTER_LATITUDE"); lat != "" {
		if latitude, err := strconv.ParseFloat(lat, 64); err == nil {
			config.Defaults.CenterLocation.Latitude = latitude
		}
	}

	if lon := os.Getenv("CENTER_LONGITUDE"); lon != "" {
		if longitude, err := strconv.ParseFloat(lon, 64); err == nil {
			config.Defaults.CenterLocation.Longitude = longitude
		}
	}

	if alt := os.Getenv("CENTER_ALTITUDE"); alt != "" {
		if altitude, err := strconv.ParseFloat(alt, 64); err == nil {
			config.Defaults.CenterLocation.Altitude = altitude
		}
	}

	// Override engagement parameters
	if kineticRatio := os.Getenv("ENGAGEMENT_TYPE_MIX"); kineticRatio != "" {
		if ratio, err := strconv.ParseFloat(kineticRatio, 64); err == nil && ratio >= 0 && ratio <= 1 {
			config.Defaults.EngagementTypeMix = ratio
			config.DefenseConfig.KineticRatio = ratio
		}
	}

	// Override formation type
	if formation := os.Getenv("SWARM_FORMATION_TYPE"); formation != "" {
		validFormations := []string{"distributed", "concentrated", "waves"}
		for _, valid := range validFormations {
			if strings.ToLower(formation) == valid {
				config.SwarmConfig.FormationType = valid
				break
			}
		}
	}

	// Override placement pattern
	if placement := os.Getenv("DEFENSE_PLACEMENT_PATTERN"); placement != "" {
		validPlacements := []string{"ring", "cluster", "line"}
		for _, valid := range validPlacements {
			if strings.ToLower(placement) == valid {
				config.DefenseConfig.PlacementPattern = valid
				break
			}
		}
	}

	// Override logging level
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		validLevels := []string{"debug", "info", "warn", "error"}
		for _, valid := range validLevels {
			if strings.ToLower(logLevel) == valid {
				config.Logging.ConsoleLevel = valid
				break
			}
		}
	}

	// Override performance settings
	if workerPool := os.Getenv("WORKER_POOL_SIZE"); workerPool != "" {
		if size, err := strconv.Atoi(workerPool); err == nil && size > 0 {
			config.Performance.WorkerPoolSize = size
		}
	}

	if batchSize := os.Getenv("BATCH_SIZE"); batchSize != "" {
		if size, err := strconv.Atoi(batchSize); err == nil && size > 0 {
			config.Performance.BatchSize = size
		}
	}

	// Override AAR settings
	if enableAAR := os.Getenv("ENABLE_AAR"); enableAAR != "" {
		if enable, err := strconv.ParseBool(enableAAR); err == nil {
			config.Logging.EnableAAR = enable
		}
	}

	if aarPath := os.Getenv("AAR_OUTPUT_PATH"); aarPath != "" {
		config.Logging.AAROutputPath = aarPath
	}

	// Override advanced settings
	if enableMetrics := os.Getenv("ENABLE_METRICS"); enableMetrics != "" {
		if enable, err := strconv.ParseBool(enableMetrics); err == nil {
			config.Advanced.EnableMetrics = enable
		}
	}

	if verboseLogging := os.Getenv("VERBOSE_LOGGING"); verboseLogging != "" {
		if enable, err := strconv.ParseBool(verboseLogging); err == nil {
			config.Advanced.VerboseLogging = enable
		}
	}
}
