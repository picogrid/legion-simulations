package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/picogrid/legion-simulations/pkg/simulation"
	"gopkg.in/yaml.v3"
)

// SimulationInfo contains information about a discovered simulation
type SimulationInfo struct {
	Path   string
	Config simulation.SimulationConfig
}

// DiscoverSimulations finds all simulations in the cmd directory
func DiscoverSimulations() ([]SimulationInfo, error) {
	var simulations []SimulationInfo

	// Get the project root
	rootDir, err := findProjectRoot()
	if err != nil {
		return nil, err
	}

	cmdDir := filepath.Join(rootDir, "cmd")

	// Walk through cmd directory
	err = filepath.Walk(cmdDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for simulation.yaml files
		if info.Name() == "simulation.yaml" {
			simInfo, err := loadSimulationConfig(path)
			if err != nil {
				// Log error but continue scanning
				fmt.Printf("Warning: failed to load %s: %v\n", path, err)
				return nil
			}
			simulations = append(simulations, *simInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan for simulations: %w", err)
	}

	return simulations, nil
}

// loadSimulationConfig loads a simulation configuration from a file
func loadSimulationConfig(path string) (*SimulationInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read simulation config: %w", err)
	}

	var config simulation.SimulationConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse simulation config: %w", err)
	}

	return &SimulationInfo{
		Path:   filepath.Dir(path),
		Config: config,
	}, nil
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot() (string, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			return "", fmt.Errorf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
