package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Environment represents a Legion environment configuration
type Environment struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key,omitempty"`
}

// Config holds the environment configurations
type Config struct {
	Environments []Environment `yaml:"environments"`
	Selected     string        `yaml:"selected,omitempty"`
}

// LoadEnvironments loads environment configurations from the default location
func LoadEnvironments() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".legion-sim", "environments.yaml")
	return LoadEnvironmentsFromFile(configPath)
}

// LoadEnvironmentsFromFile loads environment configurations from a specific file
func LoadEnvironmentsFromFile(path string) (*Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveEnvironments saves the environment configuration
func SaveEnvironments(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".legion-sim")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "environments.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getDefaultConfig returns a default configuration
func getDefaultConfig() *Config {
	return &Config{
		Environments: []Environment{
			{
				Name: "Demo",
				URL:  "https://legion-demo.com",
			},
			{
				Name: "Staging",
				URL:  "https://legion-staging.com",
			},
		},
	}
}
