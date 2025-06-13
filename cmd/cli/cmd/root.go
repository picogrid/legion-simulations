package cmd

import (
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	envName  string
	envURL   string
	logLevel string
	noColor  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "legion-sim",
	Short: "Legion simulation CLI",
	Long: `Legion Simulation CLI is a tool for running various simulations
that demonstrate Legion's C2 capabilities for unmanned systems,
data aggregation, and common operating picture generation.`,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.legion-sim/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&envName, "env", "", "environment name to use")
	rootCmd.PersistentFlags().StringVar(&envURL, "url", "", "Legion API URL (overrides environment)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	// Add commands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(envCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	// Configure logger based on flags
	logger.SetLevel(logger.ParseLevel(logLevel))
	logger.SetNoColor(noColor)

	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in home directory
		viper.AddConfigPath("$HOME/.legion-sim")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	_ = viper.ReadInConfig()
}
