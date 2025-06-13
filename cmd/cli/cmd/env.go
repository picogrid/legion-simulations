package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/picogrid/legion-simulations/pkg/config"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage Legion environments",
	Long:  `Manage Legion environment configurations`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured environments",
	RunE:  listEnvironments,
}

var envAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new environment",
	RunE:  addEnvironment,
}

var envRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an environment",
	RunE:  removeEnvironment,
}

func init() {
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envAddCmd)
	envCmd.AddCommand(envRemoveCmd)
}

func listEnvironments(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadEnvironments()
	if err != nil {
		return fmt.Errorf("failed to load environments: %w", err)
	}

	if len(cfg.Environments) == 0 {
		fmt.Println("No environments configured")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tURL\tAUTHENTICATION")
	_, _ = fmt.Fprintln(w, "----\t---\t--------------")

	for _, env := range cfg.Environments {
		authInfo := "OAuth (Interactive)"
		if env.APIKey != "" {
			authInfo = fmt.Sprintf("API Key (%s)", env.APIKey)
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", env.Name, env.URL, authInfo)
	}

	return w.Flush()
}

func addEnvironment(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadEnvironments()
	if err != nil {
		return fmt.Errorf("failed to load environments: %w", err)
	}

	var env config.Environment

	// Prompt for name
	namePrompt := &survey.Input{
		Message: "Environment name:",
	}
	if err := survey.AskOne(namePrompt, &env.Name, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Check if name already exists
	for _, existing := range cfg.Environments {
		if existing.Name == env.Name {
			return fmt.Errorf("environment %s already exists", env.Name)
		}
	}

	// Prompt for URL
	urlPrompt := &survey.Input{
		Message: "Legion API URL:",
		Default: "https://legion.example.com",
	}
	if err := survey.AskOne(urlPrompt, &env.URL, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Prompt for authentication method
	authPrompt := &survey.Select{
		Message: "Authentication method:",
		Options: []string{"OAuth (Interactive Login)", "API Key (Environment Variable)"},
		Default: "OAuth (Interactive Login)",
	}
	var authMethod string
	if err := survey.AskOne(authPrompt, &authMethod); err != nil {
		return err
	}

	if authMethod == "API Key (Environment Variable)" {
		// Prompt for API key environment variable
		apiKeyPrompt := &survey.Input{
			Message: "API key environment variable:",
			Help:    "Name of the environment variable that contains the API key",
		}
		if err := survey.AskOne(apiKeyPrompt, &env.APIKey, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	} else {
		// OAuth authentication - no API key needed
		env.APIKey = ""
	}

	// Add to config
	cfg.Environments = append(cfg.Environments, env)

	// Save config
	if err := config.SaveEnvironments(cfg); err != nil {
		return fmt.Errorf("failed to save environments: %w", err)
	}

	fmt.Printf("Environment %s added successfully\n", env.Name)
	return nil
}

func removeEnvironment(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadEnvironments()
	if err != nil {
		return fmt.Errorf("failed to load environments: %w", err)
	}

	if len(cfg.Environments) == 0 {
		fmt.Println("No environments to remove")
		return nil
	}

	// Build list of environment names
	names := make([]string, len(cfg.Environments))
	for i, env := range cfg.Environments {
		names[i] = env.Name
	}

	// Prompt for selection
	var selected string
	prompt := &survey.Select{
		Message: "Select environment to remove:",
		Options: names,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}

	// Confirm removal
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to remove %s?", selected),
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		fmt.Println("Removal cancelled")
		return nil
	}

	// Remove from config
	newEnvs := make([]config.Environment, 0, len(cfg.Environments)-1)
	for _, env := range cfg.Environments {
		if env.Name != selected {
			newEnvs = append(newEnvs, env)
		}
	}
	cfg.Environments = newEnvs

	// Save config
	if err := config.SaveEnvironments(cfg); err != nil {
		return fmt.Errorf("failed to save environments: %w", err)
	}

	fmt.Printf("Environment %s removed successfully\n", selected)
	return nil
}
