package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/picogrid/legion-simulations/pkg/auth"
	"github.com/picogrid/legion-simulations/pkg/client"
	"github.com/picogrid/legion-simulations/pkg/config"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/simulation"
	"github.com/picogrid/legion-simulations/pkg/utils"

	// Import simulations to register them
	_ "github.com/picogrid/legion-simulations/cmd/drone-swarm/simulation"
	_ "github.com/picogrid/legion-simulations/cmd/simple"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a simulation",
	Long:  `Run a simulation interactively or with specified parameters`,
	RunE:  runSimulation,
}

func init() {
	runCmd.Flags().StringP("simulation", "s", "", "simulation name to run")
	runCmd.Flags().StringP("params", "p", "", "parameters file (YAML)")
}

func runSimulation(cmd *cobra.Command, _ []string) error {
	if err := loadSimulations(); err != nil {
		return fmt.Errorf("failed to load simulations: %w", err)
	}

	envConfig, apiKey, err := selectEnvironment()
	if err != nil {
		return fmt.Errorf("failed to select environment: %w", err)
	}

	var legionClient *client.Legion

	// Check if we should use OAuth authentication
	if apiKey == "" || strings.ToLower(apiKey) == "oauth" {
		// Use the new function that fetches auth config from Legion
		tokenManager, err := auth.AuthenticateUserWithLegion(context.Background(), envConfig.URL)
		if err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}

		legionClient, err = auth.CreateAuthenticatedClient(envConfig.URL, tokenManager)
		if err != nil {
			return fmt.Errorf("failed to create authenticated client: %w", err)
		}
	} else {
		legionClient, err = client.NewLegionClient(envConfig.URL, apiKey)
		if err != nil {
			return fmt.Errorf("failed to create Legion client: %w", err)
		}
	}

	logger.Progress("Testing connection to Legion...")
	if err := legionClient.ValidateConnection(context.Background()); err != nil {
		return fmt.Errorf("failed to connect to Legion: %w", err)
	}
	logger.Success("Successfully connected to Legion")

	// Get organizations and let user select
	orgID, err := selectOrganization(legionClient)
	if err != nil {
		return fmt.Errorf("failed to select organization: %w", err)
	}

	simName, err := selectSimulation(cmd)
	if err != nil {
		return fmt.Errorf("failed to select simulation: %w", err)
	}

	sim, err := simulation.DefaultRegistry.Get(simName)
	if err != nil {
		return fmt.Errorf("failed to get simulation: %w", err)
	}

	simInfos, err := utils.DiscoverSimulations()
	if err != nil {
		return fmt.Errorf("failed to discover simulations: %w", err)
	}

	var simConfig *simulation.SimulationConfig
	for _, info := range simInfos {
		if info.Config.Name == simName {
			simConfig = &info.Config
			break
		}
	}

	if simConfig == nil {
		return fmt.Errorf("simulation configuration not found for %s", simName)
	}

	// Filter out organization_id from parameters since we already have it
	filteredParams := make([]simulation.Parameter, 0, len(simConfig.Parameters))
	for _, param := range simConfig.Parameters {
		if param.Name != "organization_id" {
			filteredParams = append(filteredParams, param)
		}
	}

	params, err := utils.PromptForParameters(filteredParams)
	if err != nil {
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Add organization ID to parameters
	params["organization_id"] = orgID

	if err := sim.Configure(params); err != nil {
		return fmt.Errorf("failed to configure simulation: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Warn("\nReceived interrupt signal, stopping simulation...")
		err := sim.Stop()
		if err != nil {
			logger.Errorf("Failed to stop simulation: %v", err)
			return
		}
		cancel()
	}()

	logger.LogSection(fmt.Sprintf("Starting %s", sim.Name()))
	if err := sim.Run(ctx, legionClient); err != nil {
		return fmt.Errorf("simulation failed: %w", err)
	}

	logger.Success("Simulation completed successfully")
	return nil
}

func loadSimulations() error {
	// For now, simulations need to be imported directly
	// This ensures their init() functions run and register themselves
	// In the future, we can use plugins or dynamic loading

	// The simple simulation is already imported via the blank import below
	return nil
}

func selectEnvironment() (*config.Environment, string, error) {
	// Check if URL is provided via flag or environment variable
	if envURL != "" {
		return &config.Environment{
			Name: "Custom",
			URL:  envURL,
		}, "", nil
	}

	// Check for environment variables
	if legionURL := os.Getenv("LEGION_URL"); legionURL != "" {
		apiKey := os.Getenv("LEGION_API_KEY")
		return &config.Environment{
			Name: "Environment",
			URL:  legionURL,
		}, apiKey, nil
	}

	// Load environment configurations
	envConfig, err := config.LoadEnvironments()
	if err != nil {
		return nil, "", err
	}

	// Check if environment is specified via flag
	if envName != "" {
		for _, env := range envConfig.Environments {
			if env.Name == envName {
				apiKey := client.GetAPIKey(env.APIKey)
				return &env, apiKey, nil
			}
		}
		return nil, "", fmt.Errorf("environment %s not found", envName)
	}

	// Interactive selection
	options := make([]string, len(envConfig.Environments)+1)
	for i, env := range envConfig.Environments {
		options[i] = env.Name
	}
	options[len(options)-1] = "Custom URL"

	var selected string
	prompt := &survey.Select{
		Message: "Select environment:",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, "", err
	}

	// Handle custom URL
	if selected == "Custom URL" {
		var customURL string
		urlPrompt := &survey.Input{
			Message: "Enter Legion API URL:",
			Default: "https://legion.example.com",
		}
		if err := survey.AskOne(urlPrompt, &customURL); err != nil {
			return nil, "", err
		}

		var apiKey string
		keyPrompt := &survey.Password{
			Message: "Enter API key (optional):",
		}
		err := survey.AskOne(keyPrompt, &apiKey)
		if err != nil {
			return nil, "", err
		}

		return &config.Environment{
			Name: "Custom",
			URL:  customURL,
		}, apiKey, nil
	}

	// Find selected environment
	for _, env := range envConfig.Environments {
		if env.Name == selected {
			apiKey := client.GetAPIKey(env.APIKey)
			if apiKey == "" && env.APIKey != "" {
				// Prompt for API key if env var is not set
				var key string
				keyPrompt := &survey.Password{
					Message: fmt.Sprintf("Enter API key for %s:", env.Name),
				}
				if err := survey.AskOne(keyPrompt, &key); err != nil {
					return nil, "", err
				}
				apiKey = key
			}
			return &env, apiKey, nil
		}
	}

	return nil, "", fmt.Errorf("environment not found")
}

func selectSimulation(cmd *cobra.Command) (string, error) {
	// Check if simulation is specified via flag
	simName, _ := cmd.Flags().GetString("simulation")
	if simName != "" {
		return simName, nil
	}

	// Discover available simulations
	simInfos, err := utils.DiscoverSimulations()
	if err != nil {
		return "", err
	}

	if len(simInfos) == 0 {
		return "", fmt.Errorf("no simulations found")
	}

	// Build options for selection
	options := make([]string, len(simInfos))
	descriptions := make(map[string]string)

	for i, info := range simInfos {
		options[i] = info.Config.Name
		descriptions[info.Config.Name] = info.Config.Description
	}

	// Interactive selection
	var selected string
	prompt := &survey.Select{
		Message: "Select simulation:",
		Options: options,
		Description: func(value string, index int) string {
			return descriptions[value]
		},
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", err
	}

	return selected, nil
}

func selectOrganization(legionClient *client.Legion) (string, error) {
	// Check if organization ID is provided via environment variable
	if orgID := os.Getenv("LEGION_ORG_ID"); orgID != "" {
		logger.Infof("Using organization ID from LEGION_ORG_ID: %s", orgID)
		return orgID, nil
	}

	if orgID := os.Getenv("LEGION_ORGANIZATION_ID"); orgID != "" {
		logger.Infof("Using organization ID from LEGION_ORGANIZATION_ID: %s", orgID)
		return orgID, nil
	}

	// For now, we'll prompt for the organization ID
	// In the future, we can enhance this to fetch the list of organizations
	// when the Legion API client supports it properly

	var orgID string
	orgPrompt := &survey.Input{
		Message: "Enter organization ID:",
		Help:    "Enter your Legion organization ID (UUID format). You can find this in the Legion UI or set LEGION_ORG_ID environment variable.",
	}

	if err := survey.AskOne(orgPrompt, &orgID, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	// Validate it's a valid UUID
	if _, err := uuid.Parse(orgID); err != nil {
		return "", fmt.Errorf("invalid organization ID format: %w", err)
	}

	logger.Infof("Using organization ID: %s", orgID)
	return orgID, nil
}
