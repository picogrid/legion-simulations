# Legion Simulations CLI

The main command-line interface for discovering and running Legion simulations.

## Overview

The Legion Simulations CLI (`legion-sim`) provides an interactive interface for:
- Managing Legion environment configurations
- Discovering available simulations
- Running simulations with configurable parameters
- Supporting both OAuth and API key authentication

## Installation

Build from source:
```bash
# From project root
make build
# Binary will be at ./bin/legion-sim
```

## Commands

### `run` - Run a simulation

Interactive simulation execution with parameter prompting.

```bash
# Interactive mode - prompts for environment, simulation, and parameters
./bin/legion-sim run

# With flags
./bin/legion-sim run --log-level debug --no-color
```

The run command will:
1. Prompt for environment selection (or use .env configuration)
2. List available simulations
3. Prompt for simulation-specific parameters
4. Execute the selected simulation
5. Handle graceful shutdown on Ctrl+C

### `list` - List available simulations

Display all registered simulations with descriptions.

```bash
./bin/legion-sim list
```

Output example:
```
Available simulations:
1. simple - Simple Entity Test
   Tests entity creation and location updates
```

### `env` - Manage Legion environments

#### `env add` - Add a new environment

```bash
./bin/legion-sim env add
```

Prompts for:
- Environment name (e.g., "dev", "staging", "prod")
- Legion API URL
- Authentication method (OAuth or API Key)
- API key environment variable name (if using API key auth)

#### `env list` - List configured environments

```bash
./bin/legion-sim env list
```

Shows all configured environments and their settings.

#### `env remove` - Remove an environment

```bash
./bin/legion-sim env remove
```

Select an environment to remove from the configuration.

## Configuration

### Environment Configuration

Environments are stored in `~/.legion/config.yaml`:

```yaml
environments:
  - name: "dev"
    url: "https://api.legion-dev.com"
    apiKey: "LEGION_DEV_API_KEY"  # Environment variable name
    authMethod: "oauth"  # or "apikey"
```

### .env File Support

Create a `.env` file in your working directory for easier development:

```bash
# Skip environment selection
LEGION_URL=https://legion-staging.com
LEGION_API_KEY=your-api-key-here

# OAuth credentials (if using OAuth)
LEGION_EMAIL=your-email@example.com
LEGION_PASSWORD=your-password

# Default values for prompts
DEFAULT_NUM_ENTITIES=5
DEFAULT_UPDATE_INTERVAL=10s

# Override parameters (skip prompts entirely)
LEGION_ORGANIZATION_ID=your-org-id-here
```

## Command-Line Flags

### Global Flags

- `--log-level` - Set logging verbosity: `debug`, `info`, `warn`, `error` (default: `info`)
- `--no-color` - Disable colored output
- `--help` / `-h` - Show help information

### Examples

```bash
# Run with debug logging
./bin/legion-sim run --log-level debug

# Run without colors (useful for CI/CD)
./bin/legion-sim run --no-color

# Get help
./bin/legion-sim --help
./bin/legion-sim run --help
```

## Authentication

The CLI supports two authentication methods:

### OAuth (Interactive)
- Fetches authorization URL from Legion API dynamically
- Prompts for email and password
- Handles token refresh automatically
- Tokens cached for session duration

### API Key
- Configure environment variable name in environment settings
- Set the environment variable before running
- No interactive prompts required

## Simulation Discovery

Simulations are discovered automatically by:
1. Scanning `cmd/` subdirectories
2. Looking for `simulation.yaml` configuration files
3. Registering simulations that implement the framework interface

To add a new simulation, see the main [README.md](../../README.md) for detailed instructions.

## Error Handling

The CLI provides clear error messages and suggestions:
- Network connectivity issues
- Authentication failures
- Invalid parameters
- Missing environment variables
- Simulation errors

## Logging

Use `--log-level debug` to see detailed information about:
- API requests and responses
- Authentication flow
- Simulation lifecycle events
- Parameter validation

## Keyboard Shortcuts

- `Ctrl+C` - Graceful shutdown of running simulation
- `Ctrl+D` - Exit from prompts (when applicable)

## Examples

### Basic Usage
```bash
# First time setup
./bin/legion-sim env add

# Run a simulation
./bin/legion-sim run

# List what's available
./bin/legion-sim list
```

### With Environment Variables
```bash
# Set up authentication
export LEGION_DEV_API_KEY="your-key-here"

# Run with defaults from .env
./bin/legion-sim run

# Debug mode
./bin/legion-sim run --log-level debug
```

### CI/CD Usage
```bash
# Non-interactive with all parameters
export LEGION_URL="https://api.legion-staging.com"
export LEGION_API_KEY="your-key"
export LEGION_ORGANIZATION_ID="your-org-id"
export LEGION_NUM_ENTITIES=10
export LEGION_UPDATE_INTERVAL=5s

./bin/legion-sim run --no-color
```