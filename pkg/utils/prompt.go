package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/picogrid/legion-simulations/pkg/simulation"
)

// PromptForParameters prompts the user for simulation parameters
func PromptForParameters(params []simulation.Parameter) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, param := range params {
		value, err := promptForParameter(param)
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", param.Name, err)
		}
		result[param.Name] = value
	}

	return result, nil
}

// promptForParameter prompts for a single parameter
func promptForParameter(param simulation.Parameter) (interface{}, error) {
	// Check if we should skip prompts entirely (for CI/automation)
	if os.Getenv("LEGION_SKIP_PROMPTS") == "true" {
		// Check for environment variable override
		envKey := "LEGION_" + strings.ToUpper(param.Name)
		if envValue := os.Getenv(envKey); envValue != "" {
			return parseEnvValue(envValue, param)
		}
		// Use default if no env var set
		if param.Default != nil {
			return param.Default, nil
		}
		if param.Required {
			return nil, fmt.Errorf("required parameter %s not provided and no default available", param.Name)
		}
	}

	// Check for environment variable to use as default
	envKey := "LEGION_" + strings.ToUpper(param.Name)
	if envValue := os.Getenv(envKey); envValue != "" {
		// Update the default value from environment
		parsed, err := parseEnvValue(envValue, param)
		if err == nil {
			param.Default = parsed
		}
	}

	switch param.Type {
	case "integer":
		return promptInteger(param)
	case "float":
		return promptFloat(param)
	case "string":
		return promptString(param)
	case "boolean":
		return promptBoolean(param)
	case "duration":
		return promptDuration(param)
	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", param.Type)
	}
}

// parseEnvValue parses an environment variable value according to the parameter type
func parseEnvValue(value string, param simulation.Parameter) (interface{}, error) {
	switch param.Type {
	case "integer":
		return strconv.Atoi(value)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "string":
		return value, nil
	case "boolean":
		return strconv.ParseBool(value)
	case "duration":
		duration, err := time.ParseDuration(value)
		if err != nil {
			return nil, err
		}
		return duration, nil
	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", param.Type)
	}
}

func promptInteger(param simulation.Parameter) (int, error) {
	defaultStr := ""
	if param.Default != nil {
		switch v := param.Default.(type) {
		case int:
			defaultStr = strconv.Itoa(v)
		case float64:
			defaultStr = strconv.Itoa(int(v))
		}
	}

	prompt := &survey.Input{
		Message: param.Description,
		Default: defaultStr,
	}

	var result string
	if err := survey.AskOne(prompt, &result, survey.WithValidator(survey.Required)); err != nil {
		return 0, err
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %w", err)
	}

	// Validate range
	if param.Min != nil {
		minRange := toInt(param.Min)
		if value < minRange {
			return 0, fmt.Errorf("value must be at least %d", minRange)
		}
	}
	if param.Max != nil {
		maxRange := toInt(param.Max)
		if value > maxRange {
			return 0, fmt.Errorf("value must be at most %d", maxRange)
		}
	}

	return value, nil
}

func promptFloat(param simulation.Parameter) (float64, error) {
	defaultStr := ""
	if param.Default != nil {
		defaultStr = fmt.Sprintf("%v", param.Default)
	}

	prompt := &survey.Input{
		Message: param.Description,
		Default: defaultStr,
	}

	var result string
	if err := survey.AskOne(prompt, &result, survey.WithValidator(survey.Required)); err != nil {
		return 0, err
	}

	value, err := strconv.ParseFloat(result, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %w", err)
	}

	// Validate range
	if param.Min != nil {
		minRange := toFloat64(param.Min)
		if value < minRange {
			return 0, fmt.Errorf("value must be at least %g", minRange)
		}
	}
	if param.Max != nil {
		maxRange := toFloat64(param.Max)
		if value > maxRange {
			return 0, fmt.Errorf("value must be at most %g", maxRange)
		}
	}

	return value, nil
}

func promptString(param simulation.Parameter) (string, error) {
	defaultStr := ""
	if param.Default != nil {
		defaultStr = fmt.Sprintf("%v", param.Default)
	}

	// If options are provided, use a select prompt
	if len(param.Options) > 0 {
		prompt := &survey.Select{
			Message: param.Description,
			Options: param.Options,
			Default: defaultStr,
		}

		var result string
		if err := survey.AskOne(prompt, &result); err != nil {
			return "", err
		}
		return result, nil
	}

	// Otherwise use input prompt
	prompt := &survey.Input{
		Message: param.Description,
		Default: defaultStr,
	}

	var result string
	var validators []survey.Validator
	if param.Required {
		validators = append(validators, survey.Required)
	}

	if err := survey.AskOne(prompt, &result, survey.WithValidator(survey.ComposeValidators(validators...))); err != nil {
		return "", err
	}

	return result, nil
}

func promptBoolean(param simulation.Parameter) (bool, error) {
	defaultBool := false
	if param.Default != nil {
		switch v := param.Default.(type) {
		case bool:
			defaultBool = v
		case string:
			defaultBool = v == "true" || v == "yes" || v == "1"
		}
	}

	prompt := &survey.Confirm{
		Message: param.Description,
		Default: defaultBool,
	}

	var result bool
	if err := survey.AskOne(prompt, &result); err != nil {
		return false, err
	}

	return result, nil
}

func promptDuration(param simulation.Parameter) (time.Duration, error) {
	defaultStr := ""
	if param.Default != nil {
		defaultStr = fmt.Sprintf("%v", param.Default)
	}

	prompt := &survey.Input{
		Message: param.Description + " (e.g., 5m, 1h30m, 30s)",
		Default: defaultStr,
	}

	var result string
	if err := survey.AskOne(prompt, &result, survey.WithValidator(func(val interface{}) error {
		str := val.(string)
		_, err := time.ParseDuration(str)
		if err != nil {
			return fmt.Errorf("invalid duration format (use formats like 5m, 1h30m, 30s)")
		}
		return nil
	})); err != nil {
		return 0, err
	}

	duration, err := time.ParseDuration(result)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// Helper functions
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}
