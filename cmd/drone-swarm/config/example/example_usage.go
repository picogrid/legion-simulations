package main

import (
	"fmt"
	"log"
	"os"

	"github.com/picogrid/legion-simulations/cmd/drone-swarm/config"
)

func main() {
	fmt.Println("=== Counter-UAS Configuration System Demo ===")

	// 1. Load default configuration
	fmt.Println("1. Loading default configuration...")
	_ = config.GetDefaultConfig()
	fmt.Printf("Default config loaded successfully.\n\n")

	// 2. Load from YAML file
	fmt.Println("2. Loading configuration from YAML file...")
	_, err := config.LoadConfigOrDefault("config.yaml")
	if err != nil {
		log.Printf("Warning: Could not load YAML config: %v", err)
	} else {
		fmt.Printf("YAML config loaded successfully.\n\n")
	}

	// 3. Demonstrate environment variable overrides
	fmt.Println("3. Demonstrating environment variable overrides...")
	os.Setenv("NUM_COUNTER_UAS_SYSTEMS", "8")
	os.Setenv("NUM_UAS_THREATS", "25")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ENGAGEMENT_TYPE_MIX", "0.8")

	envConfig := config.GetDefaultConfig()
	config.MergeWithEnvironment(envConfig)
	fmt.Printf("Environment overrides applied.\n\n")

	// 4. Demonstrate CLI overrides
	fmt.Println("4. Demonstrating CLI overrides...")
	cliOverrides := map[string]interface{}{
		"formation_type":    "waves",
		"placement_pattern": "cluster",
		"wave_count":        5,
		"verbose_logging":   true,
	}

	cliConfig := config.GetDefaultConfig()
	config.MergeWithCLIOverrides(cliConfig, cliOverrides)
	fmt.Printf("CLI overrides applied.\n\n")

	// 5. Show the final configuration
	fmt.Println("5. Final configuration with all overrides:")
	finalConfig, err := config.LoadConfigWithOverrides("config.yaml", cliOverrides)
	if err != nil {
		log.Printf("Error loading config with overrides: %v", err)
		finalConfig = cliConfig
	}

	fmt.Println("\n" + finalConfig.String())

	// 6. Validate configuration
	fmt.Println("\n\n6. Configuration validation:")
	if err := finalConfig.Validate(); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Configuration validation passed.\n")
	}

	// 7. Show key configuration highlights
	fmt.Printf("\n7. Key simulation parameters:\n")
	fmt.Printf("   • Simulation: %s\n", finalConfig.Simulation.Name)
	fmt.Printf("   • Entities: %d Counter-UAS vs %d UAS threats\n",
		finalConfig.Defaults.NumCounterUASSystems,
		finalConfig.Defaults.NumUASThreats)
	fmt.Printf("   • Formation: %s with %d waves\n",
		finalConfig.SwarmConfig.FormationType,
		finalConfig.SwarmConfig.WaveCount)
	fmt.Printf("   • Defense: %s placement with %s engagement\n",
		finalConfig.DefenseConfig.PlacementPattern,
		finalConfig.DefenseConfig.EngagementRules)
	fmt.Printf("   • Engagement Mix: %.0f%% kinetic, %.0f%% EW\n",
		finalConfig.Defaults.EngagementTypeMix*100,
		(1-finalConfig.Defaults.EngagementTypeMix)*100)
	fmt.Printf("   • Success Rates: Kinetic %.0f%%-%.0f%%, EW %.0f%%-%.0f%%\n",
		finalConfig.Engagement.KineticSuccessRateRange.Min*100,
		finalConfig.Engagement.KineticSuccessRateRange.Max*100,
		finalConfig.Engagement.EWSuccessRateRange.Min*100,
		finalConfig.Engagement.EWSuccessRateRange.Max*100)

	fmt.Println("\n=== Configuration Demo Complete ===")
}
