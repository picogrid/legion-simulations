package config

import (
	"fmt"
	"time"
)

// SimulationConfig holds the complete simulation configuration
type SimulationConfig struct {
	// Basic simulation settings
	Simulation SimulationSettings `yaml:"simulation"`

	// Legion settings
	OrganizationID string `yaml:"organization_id,omitempty"`

	// Performance settings
	Performance PerformanceConfig `yaml:"performance"`

	// Swarm configuration
	SwarmConfig SwarmConfig `yaml:"swarm_config"`

	// Defense configuration
	DefenseConfig DefenseConfig `yaml:"defense_config"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging"`

	// Default parameters
	Defaults DefaultsConfig `yaml:"defaults"`

	// Advanced options
	Advanced AdvancedConfig `yaml:"advanced"`

	// Engagement parameters
	Engagement EngagementConfig `yaml:"engagement"`

	// Target prioritization
	TargetPriority TargetPriorityConfig `yaml:"target_priority"`

	// Termination conditions
	Termination TerminationConfig `yaml:"termination"`
}

// SimulationSettings holds basic simulation settings
type SimulationSettings struct {
	Name           string        `yaml:"name"`
	Description    string        `yaml:"description"`
	UpdateInterval time.Duration `yaml:"update_interval"`
}

// Location represents a geographic location
type Location struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
	Altitude  float64 `yaml:"altitude"`
}

// SpeedRange defines a range of speeds
type SpeedRange struct {
	Min int `yaml:"min"` // kph
	Max int `yaml:"max"` // kph
}

// CooldownRange defines a range of cooldown times
type CooldownRange struct {
	Min int `yaml:"min"` // seconds
	Max int `yaml:"max"` // seconds
}

// SuccessRateRange defines a range of success rates
type SuccessRateRange struct {
	Min float64 `yaml:"min"` // 0.0 to 1.0
	Max float64 `yaml:"max"` // 0.0 to 1.0
}

// SwarmConfig defines UAS swarm configuration
type SwarmConfig struct {
	FormationType        string        `yaml:"formation_type"` // "distributed", "concentrated", "waves"
	WaveDelay            time.Duration `yaml:"wave_delay"`
	WaveCount            int           `yaml:"wave_count"`
	AutonomyDistribution string        `yaml:"autonomy_distribution"` // "low", "mixed", "high"
	EvasionProbability   float64       `yaml:"evasion_probability"`   // 0.0 to 1.0
	SpeedRange           SpeedRange    `yaml:"speed_range"`
}

// DefenseConfig defines Counter-UAS system configuration
type DefenseConfig struct {
	PlacementPattern     string        `yaml:"placement_pattern"`     // "ring", "cluster", "line"
	EngagementRules      string        `yaml:"engagement_rules"`      // "closest", "highest_threat", "distributed"
	KineticRatio         float64       `yaml:"kinetic_ratio"`         // 0.0 to 1.0
	SuccessRateModifier  float64       `yaml:"success_rate_modifier"` // difficulty adjustment
	DetectionRadiusKm    float64       `yaml:"detection_radius_km"`
	EngagementRadiusKm   float64       `yaml:"engagement_radius_km"`
	KineticCooldownRange CooldownRange `yaml:"kinetic_cooldown_range"`
	EWCooldownRange      CooldownRange `yaml:"ew_cooldown_range"`
}

// LoggingConfig defines logging and reporting settings
type LoggingConfig struct {
	ConsoleLevel    string `yaml:"console_level"` // "debug", "info", "warn", "error"
	EnableAAR       bool   `yaml:"enable_aar"`
	AARFormat       string `yaml:"aar_format"` // "summary", "detailed", "full"
	AAROutputPath   string `yaml:"aar_output_path"`
	EventBufferSize int    `yaml:"event_buffer_size"`
}

// DefaultsConfig defines default simulation parameters
type DefaultsConfig struct {
	NumCounterUASSystems int      `yaml:"num_counter_uas_systems"`
	NumUASThreats        int      `yaml:"num_uas_threats"`
	EngagementTypeMix    float64  `yaml:"engagement_type_mix"` // 0.0 to 1.0 (kinetic ratio)
	CenterLocation       Location `yaml:"center_location"`
}

// AdvancedConfig defines advanced simulation options
type AdvancedConfig struct {
	EnableMetrics           bool          `yaml:"enable_metrics"`
	MetricsExportInterval   time.Duration `yaml:"metrics_export_interval"`
	RecordReplay            bool          `yaml:"record_replay"`
	ReplayFilePath          string        `yaml:"replay_file_path"`
	VerboseLogging          bool          `yaml:"verbose_logging"`
	DebugEngagementCalcs    bool          `yaml:"debug_engagement_calculations"`
	RandomizeSpawnLocations bool          `yaml:"randomize_spawn_locations"`
	SpawnRadiusKm           float64       `yaml:"spawn_radius_km"`
}

// EngagementConfig defines engagement parameters
type EngagementConfig struct {
	KineticSuccessRateRange  SuccessRateRange `yaml:"kinetic_success_rate_range"`
	EWSuccessRateRange       SuccessRateRange `yaml:"ew_success_rate_range"`
	KineticAmmoCapacity      int              `yaml:"kinetic_ammo_capacity"`
	JammingAutonomyThreshold float64          `yaml:"jamming_autonomy_threshold"` // 0.0 to 1.0
}

// RoleMultipliers defines priority multipliers for different UAS roles
type RoleMultipliers struct {
	Leader   float64 `yaml:"leader"`
	Follower float64 `yaml:"follower"`
	Scout    float64 `yaml:"scout"`
}

// TargetPriorityConfig defines target prioritization weights
type TargetPriorityConfig struct {
	DistanceWeight  float64         `yaml:"distance_weight"`
	SpeedWeight     float64         `yaml:"speed_weight"`
	RoleWeight      float64         `yaml:"role_weight"`
	RoleMultipliers RoleMultipliers `yaml:"role_multipliers"`
}

// TerminationConfig defines victory and termination conditions
type TerminationConfig struct {
	SuccessConditions   []string `yaml:"success_conditions"`
	FailureConditions   []string `yaml:"failure_conditions"`
	StalemateConditions []string `yaml:"stalemate_conditions"`
}

// PerformanceConfig defines performance settings
type PerformanceConfig struct {
	WorkerPoolSize          int           `yaml:"worker_pool_size"`
	BatchSize               int           `yaml:"batch_size"`
	APIRateLimit            int           `yaml:"api_rate_limit"`
	UpdateFlushInterval     time.Duration `yaml:"update_flush_interval"`
	MaxConcurrentGoroutines int           `yaml:"max_concurrent_goroutines"`
}

// Validate checks if the configuration is valid
func (c *SimulationConfig) Validate() error {
	if c.Simulation.Name == "" {
		return fmt.Errorf("simulation name is required")
	}

	if c.Simulation.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	if c.Defaults.NumCounterUASSystems <= 0 {
		return fmt.Errorf("number of Counter-UAS systems must be positive")
	}

	if c.Defaults.NumUASThreats <= 0 {
		return fmt.Errorf("number of UAS threats must be positive")
	}

	// Validate probability ranges
	if c.SwarmConfig.EvasionProbability < 0 || c.SwarmConfig.EvasionProbability > 1 {
		return fmt.Errorf("evasion probability must be between 0.0 and 1.0")
	}

	if c.DefenseConfig.KineticRatio < 0 || c.DefenseConfig.KineticRatio > 1 {
		return fmt.Errorf("kinetic ratio must be between 0.0 and 1.0")
	}

	if c.Defaults.EngagementTypeMix < 0 || c.Defaults.EngagementTypeMix > 1 {
		return fmt.Errorf("engagement type mix must be between 0.0 and 1.0")
	}

	// Validate speed ranges
	if c.SwarmConfig.SpeedRange.Min >= c.SwarmConfig.SpeedRange.Max {
		return fmt.Errorf("speed range min must be less than max")
	}

	// Validate success rate ranges
	if c.Engagement.KineticSuccessRateRange.Min >= c.Engagement.KineticSuccessRateRange.Max {
		return fmt.Errorf("kinetic success rate range min must be less than max")
	}

	if c.Engagement.EWSuccessRateRange.Min >= c.Engagement.EWSuccessRateRange.Max {
		return fmt.Errorf("EW success rate range min must be less than max")
	}

	// Validate priority weights sum to reasonable values
	weightSum := c.TargetPriority.DistanceWeight + c.TargetPriority.SpeedWeight + c.TargetPriority.RoleWeight
	if weightSum <= 0 {
		return fmt.Errorf("target priority weights must sum to a positive value")
	}

	return nil
}

// String returns a human-readable representation of the configuration
func (c *SimulationConfig) String() string {
	return fmt.Sprintf(`Simulation Configuration:
  Name: %s
  Description: %s
  Update Interval: %v
  
Entities:
  Counter-UAS Systems: %d
  UAS Threats: %d
  
Swarm Configuration:
  Formation: %s
  Wave Count: %d
  Wave Delay: %v
  Autonomy Distribution: %s
  Evasion Probability: %.2f
  Speed Range: %d-%d kph
  
Defense Configuration:
  Placement Pattern: %s
  Engagement Rules: %s
  Kinetic Ratio: %.2f
  Success Rate Modifier: %.2f
  Detection Radius: %.1f km
  Engagement Radius: %.1f km
  
Engagement Parameters:
  Kinetic Success Rate: %.2f-%.2f
  EW Success Rate: %.2f-%.2f
  Kinetic Ammo Capacity: %d
  Jamming Autonomy Threshold: %.2f
  
Performance:
  Worker Pool Size: %d
  Batch Size: %d
  API Rate Limit: %d
  
Logging:
  Console Level: %s
  AAR Enabled: %t
  AAR Format: %s`,
		c.Simulation.Name,
		c.Simulation.Description,
		c.Simulation.UpdateInterval,
		c.Defaults.NumCounterUASSystems,
		c.Defaults.NumUASThreats,
		c.SwarmConfig.FormationType,
		c.SwarmConfig.WaveCount,
		c.SwarmConfig.WaveDelay,
		c.SwarmConfig.AutonomyDistribution,
		c.SwarmConfig.EvasionProbability,
		c.SwarmConfig.SpeedRange.Min,
		c.SwarmConfig.SpeedRange.Max,
		c.DefenseConfig.PlacementPattern,
		c.DefenseConfig.EngagementRules,
		c.DefenseConfig.KineticRatio,
		c.DefenseConfig.SuccessRateModifier,
		c.DefenseConfig.DetectionRadiusKm,
		c.DefenseConfig.EngagementRadiusKm,
		c.Engagement.KineticSuccessRateRange.Min,
		c.Engagement.KineticSuccessRateRange.Max,
		c.Engagement.EWSuccessRateRange.Min,
		c.Engagement.EWSuccessRateRange.Max,
		c.Engagement.KineticAmmoCapacity,
		c.Engagement.JammingAutonomyThreshold,
		c.Performance.WorkerPoolSize,
		c.Performance.BatchSize,
		c.Performance.APIRateLimit,
		c.Logging.ConsoleLevel,
		c.Logging.EnableAAR,
		c.Logging.AARFormat,
	)
}

// GetDefaultConfig returns a default configuration matching the Counter-UAS simulation plan
func GetDefaultConfig() *SimulationConfig {
	return &SimulationConfig{
		Simulation: SimulationSettings{
			Name:           "drone-swarm",
			Description:    "Counter-UAS vs Drone Swarm Engagement Simulation",
			UpdateInterval: 3 * time.Second,
		},

		Performance: PerformanceConfig{
			WorkerPoolSize:          10,
			BatchSize:               50,
			APIRateLimit:            100,
			UpdateFlushInterval:     1 * time.Second,
			MaxConcurrentGoroutines: 20,
		},

		SwarmConfig: SwarmConfig{
			FormationType:        "distributed",
			WaveDelay:            45 * time.Second,
			WaveCount:            3,
			AutonomyDistribution: "mixed",
			EvasionProbability:   0.7,
			SpeedRange: SpeedRange{
				Min: 50,
				Max: 200,
			},
		},

		DefenseConfig: DefenseConfig{
			PlacementPattern:    "ring",
			EngagementRules:     "closest",
			KineticRatio:        0.7,
			SuccessRateModifier: 1.0,
			DetectionRadiusKm:   10,
			EngagementRadiusKm:  5,
			KineticCooldownRange: CooldownRange{
				Min: 5,
				Max: 8,
			},
			EWCooldownRange: CooldownRange{
				Min: 8,
				Max: 10,
			},
		},

		Logging: LoggingConfig{
			ConsoleLevel:    "info",
			EnableAAR:       true,
			AARFormat:       "detailed",
			AAROutputPath:   "./reports/",
			EventBufferSize: 1000,
		},

		Defaults: DefaultsConfig{
			NumCounterUASSystems: 5,
			NumUASThreats:        20,
			EngagementTypeMix:    0.7,
			CenterLocation: Location{
				Latitude:  37.7749,
				Longitude: -122.4194,
				Altitude:  100,
			},
		},

		Advanced: AdvancedConfig{
			EnableMetrics:           true,
			MetricsExportInterval:   10 * time.Second,
			RecordReplay:            false,
			ReplayFilePath:          "./replays/",
			VerboseLogging:          false,
			DebugEngagementCalcs:    false,
			RandomizeSpawnLocations: true,
			SpawnRadiusKm:           12,
		},

		Engagement: EngagementConfig{
			KineticSuccessRateRange: SuccessRateRange{
				Min: 0.7,
				Max: 0.9,
			},
			EWSuccessRateRange: SuccessRateRange{
				Min: 0.5,
				Max: 0.7,
			},
			KineticAmmoCapacity:      5,
			JammingAutonomyThreshold: 0.5,
		},

		TargetPriority: TargetPriorityConfig{
			DistanceWeight: 0.5,
			SpeedWeight:    0.3,
			RoleWeight:     0.2,
			RoleMultipliers: RoleMultipliers{
				Leader:   1.5,
				Follower: 1.0,
				Scout:    1.2,
			},
		},

		Termination: TerminationConfig{
			SuccessConditions:   []string{"all_threats_neutralized"},
			FailureConditions:   []string{"defensive_breach"},
			StalemateConditions: []string{"all_systems_depleted"},
		},
	}
}
