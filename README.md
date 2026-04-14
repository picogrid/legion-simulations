# Legion Simulations

A Go-based simulation framework for demonstrating Legion C2 system capabilities through various scenarios including drone swarms, weather events, satellite operations, and more.

## Quick and Easy Start

```bash
# Clone and build
git clone https://github.com/picogrid/legion-simulations.git
cd legion-simulations
make build

# Run the most exciting simulation - Drone Swarm Combat!
./bin/legion-sim run -s "Drone Swarm Combat"
```

That's it! The CLI will guide you through environment setup and authentication.

## Overview

Legion Simulations provides a flexible, extensible framework for creating simulations that showcase Legion's command and control capabilities for unmanned systems, data aggregation, and common operating picture generation. The framework is designed to be configuration-driven, making it easy to create new simulations and scenarios.

## Features

- **Extensible Framework**: Easy-to-implement simulation interface for creating new scenarios
- **Interactive CLI**: User-friendly command-line interface for discovering and running simulations
- **Configuration-Driven**: YAML-based configuration for simulation parameters
- **Environment Management**: Support for multiple Legion environments (dev, staging, production)
- **Real-time Updates**: Simulations can update entity locations and states in real-time
- **Clean API Client**: Hand-written client for maintainability and simplicity

## Prerequisites

- Go 1.24 or higher
- macOS (for development environment setup)
- Access to a Legion instance

## Quick Start

### 1. Set Up Development Environment

```bash
# Complete setup (macOS)
make dev-env

# Or individual components:
make dev-brew        # Install golangci-lint, pre-commit
make dev-precommit   # Configure git hooks
make dev-gotooling   # Install Go development tools
```

### 2. Build the CLI

```bash
make build
# Or directly:
# go build -o bin/legion-sim ./cmd/cli
```

### 3. Configure Environments

```bash
# Add a Legion environment
./bin/legion-sim env add

# You'll be prompted for:
# - Environment name (e.g., "dev", "staging", "prod")
# - Legion API URL (e.g., https://legion-staging.com)
# - Authentication method (OAuth or API Key)

# List configured environments
./bin/legion-sim env list

# Remove an environment
./bin/legion-sim env remove
```

#### Managing Multiple Environments

You can easily switch between different Legion environments:

1. **Interactive Selection**: When running simulations, you'll be prompted to choose from your configured environments
2. **Environment Variables**: Skip the prompt by setting:
   ```bash
   export LEGION_URL=https://legion-staging.com
   export LEGION_API_KEY=your-staging-key
   ```
3. **Direct Edit**: Modify `~/.legion/config.yaml` directly

#### Authentication Options

Legion Simulations supports multiple authentication methods:

| Authentication Type | Status | Description |
|-------------------|--------|-------------|
| User Auth (OAuth) | ✅ | Interactive login with email/password |
| API Tokens | 🚧 TBD | Direct API token authentication |
| Integration Auth | 🚧 TBD | OAuth client credentials flow for integrations |

**Currently Supported:**

1. **OAuth (Interactive Login)** - Recommended for user accounts
   - Prompts for email and password when running simulations
   - Dynamically fetches authorization URL from Legion API
   - Automatically handles token refresh
   - Secure password input (hidden)

2. **API Key (Environment Variable)** - For automation
   - Uses an environment variable containing the API key
   - Set the variable before running: `export LEGION_API_KEY=your-key-here`

Environments are stored in `~/.legion/config.yaml`

### 4. Run a Simulation

```bash
# Interactive mode - prompts for all options
./bin/legion-sim run

# The CLI will:
# 1. Prompt for environment selection (or use LEGION_URL/LEGION_API_KEY env vars)
# 2. Authenticate with Legion (OAuth or API key)
# 3. Prompt for organization selection (or use LEGION_ORG_ID env var)
# 4. Show available simulations
# 5. Prompt for simulation parameters
# 6. Run the simulation

# List available simulations
./bin/legion-sim list
```

## Project Structure

```
legion-simulations/
   cmd/                      # Executable applications
      cli/                  # Main CLI application
      simple/               # Simple entity test simulation
      drone-swarm/          # Drone swarm simulation (planned)
      weather-events/       # Weather events simulation (planned)
      satellite-ops/        # Satellite operations simulation (planned)
   pkg/                      # Shared packages
      client/               # Hand-written Legion API client (organized by domain)
         client.go          # Core client and HTTP request handling
         entities.go        # Entity management operations
         locations.go       # Entity location operations
         users.go           # User and authentication operations
         organizations.go   # Organization management
         feeds.go           # Feed definition and data operations
         helpers.go         # Helper functions for API operations
      models/               # Generated API models
      simulation/           # Core simulation framework
      config/               # Environment configuration
      utils/                # Utility functions
      auth/                 # Authentication (Keycloak client, token management)
   openapi.yaml              # Legion API specification
   Makefile                  # Build and development tasks
   go.mod                    # Go module definition
```

## Creating a New Simulation

### 1. Create Directory Structure

```bash
mkdir -p cmd/my-simulation
cd cmd/my-simulation
```

### 2. Define Simulation Configuration

Create `simulation.yaml`:

```yaml
name: "My Simulation"
description: "Description of what this simulation does"
version: "1.0.0"
category: "demo"

parameters:
  - name: "num_entities"
    type: "integer"
    description: "Number of entities to create"
    default: 10
    min: 1
    max: 100
    required: true
  
  - name: "update_interval"
    type: "float"
    description: "Update frequency in seconds"
    default: 5.0
    min: 1.0
    max: 60.0
    required: true
  
  - name: "organization_id"
    type: "string"
    description: "Organization ID for entity creation"
    required: true
```

### 3. Implement the Simulation

Create `simulation.go`:

```go
package mysimulation

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/picogrid/legion-simulations/pkg/client"
    "github.com/picogrid/legion-simulations/pkg/models"
    "github.com/picogrid/legion-simulations/pkg/simulation"
)

type MySimulation struct {
    // Configuration from parameters
    numEntities     int
    updateInterval  time.Duration
    organizationID  string
    
    // Runtime state
    entities []string
    stopChan chan struct{}
}

func NewMySimulation() simulation.Simulation {
    return &MySimulation{
        stopChan: make(chan struct{}),
    }
}

func (s *MySimulation) Name() string {
    return "My Simulation"
}

func (s *MySimulation) Description() string {
    return "Description of what this simulation does"
}

func (s *MySimulation) Configure(params map[string]interface{}) error {
    // Parse parameters with type checking
    if v, ok := params["num_entities"].(float64); ok {
        s.numEntities = int(v)
    } else {
        return fmt.Errorf("num_entities must be a number")
    }
    
    if v, ok := params["update_interval"].(float64); ok {
        s.updateInterval = time.Duration(v) * time.Second
    } else {
        return fmt.Errorf("update_interval must be a number")
    }
    
    if v, ok := params["organization_id"].(string); ok {
        s.organizationID = v
    } else {
        return fmt.Errorf("organization_id must be a string")
    }
    
    return nil
}

func (s *MySimulation) Run(ctx context.Context, legionClient *client.Legion) error {
    log.Printf("Starting simulation with %d entities", s.numEntities)
    
    // Create entities
    for i := 0; i < s.numEntities; i++ {
        entityID, err := s.createEntity(ctx, legionClient, i)
        if err != nil {
            return fmt.Errorf("failed to create entity %d: %w", i, err)
        }
        s.entities = append(s.entities, entityID)
        log.Printf("Created entity %d: %s", i+1, entityID)
    }
    
    // Update loop
    ticker := time.NewTicker(s.updateInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-s.stopChan:
            return nil
        case <-ticker.C:
            if err := s.updateEntities(ctx, legionClient); err != nil {
                log.Printf("Error updating entities: %v", err)
            }
        }
    }
}

func (s *MySimulation) Stop() error {
    close(s.stopChan)
    return nil
}

// Helper function to create an entity
func (s *MySimulation) createEntity(ctx context.Context, legionClient *client.Legion, index int) (string, error) {
    // Helper to create string pointers (required by the API models)
    strPtr := func(s string) *string { return &s }
    categoryPtr := func(c models.Category) *models.Category { return &c }
    orgID := uuid.MustParse(s.organizationID)
    
    req := &models.CreateEntityRequest{
        Name:           strPtr(fmt.Sprintf("Sim Entity %d", index+1)),
        Type:           strPtr("simulation"),
        Category:       categoryPtr(models.CategoryDEVICE),
        Status:         strPtr("active"),
        OrganizationID: &orgID,
        Metadata:       &metadataRaw,
    }
    
    resp, err := legionClient.CreateEntity(ctx, req)
    if err != nil {
        return "", err
    }
    
    return resp.ID.String(), nil
}

// Helper function to update entity locations
func (s *MySimulation) updateEntities(ctx context.Context, legionClient *client.Legion) error {
    for _, entityID := range s.entities {
        recordedAt := time.Now()
        req := &models.CreateEntityLocationRequest{
            Position: &models.GeomPoint{
                Type:        strPtr("Point"),
                Coordinates: []float64{4517590.878, 832293.160, 4524856.575},
            },
            RecordedAt: &recordedAt,
            Source:     "my-simulation",
        }
        
        _, err := legionClient.CreateEntityLocation(ctx, entityID, req)
        if err != nil {
            log.Printf("Failed to update location for entity %s: %v", entityID, err)
        }
    }
    
    log.Printf("Updated locations for %d entities", len(s.entities))
    return nil
}

func init() {
    simulation.DefaultRegistry.Register("my-simulation", NewMySimulation)
}
```

### 4. Add Configuration Type (Optional)

For better type safety, create `config.go`:

```go
package mysimulation

type Config struct {
    NumEntities     int     `yaml:"num_entities"`
    UpdateInterval  float64 `yaml:"update_interval"`
    OrganizationID  string  `yaml:"organization_id"`
}
```

### 5. Directory Structure

Your complete simulation directory should look like:

```
cmd/my-simulation/
├── simulation.yaml    # Configuration and parameters
├── simulation.go      # Main simulation implementation
└── config.go         # Optional: Type definitions
```

### 6. Register the Simulation

Add import to `cmd/cli/cmd/run.go`:

```go
import (
    // ... other imports ...
    _ "github.com/picogrid/legion-simulations/cmd/my-simulation"
)
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Run specific package tests
go test -v ./pkg/simulation/...
```

### Linting

```bash
# Run linter with auto-fix
make lint-fix

# Run linter without auto-fix
make lint
```

### Building

```bash
# Build all binaries
make build

# Clean build artifacts
make clean

# Clean and rebuild
make rebuild
```

## Technical Details

### Legion Client

The project uses a hand-written Legion API client (`pkg/client/`) on top of OpenAPI-derived raw models in `pkg/models/openapi.gen.go`. This provides:
- Simplified API without complex dependencies
- Easy-to-understand code structure organized by domain
- Context-aware operations with proper authentication
- Clean error handling
- Type safety using the domain models in `pkg/models/`

The client is organized into domain-specific files:
- `client.go` - Core client functionality and HTTP request handling
- `entities.go` - Entity creation, updates, deletion, and search
- `locations.go` - Entity location management
- `users.go` - User profile and authentication
- `organizations.go` - Organization and user management
- `feeds.go` - Feed definitions and data ingestion
- `helpers.go` - Utility functions for API operations

### Working with Legion API

The Legion client provides methods for all major operations:

```go
// Creating entities
entity, err := client.CreateEntity(ctx, &models.CreateEntityRequest{...})

// Updating entity locations (ECEF coordinates)
location, err := client.CreateEntityLocation(ctx, entityID, &models.CreateEntityLocationRequest{...})

// Creating feeds for data ingestion
feed, err := client.CreateFeedDefinition(ctx, &models.CreateFeedDefinitionRequest{...})

// Sending telemetry data
err := client.IngestFeedData(ctx, &models.IngestFeedDataRequest{...})

// Search for entities
entities, err := client.SearchEntities(ctx, params)
```

### Model Generation

`openapi.yaml` is the checked-in source spec. Regenerate the raw OAS3-backed model layer with:

```bash
make generate-models
# or
go generate ./pkg/models
```

The generation flow uses `oapi-codegen` plus `cmd/tools/openapi-normalize` to normalize the OpenAPI 3.1 spec for generation without forking the checked-in source.

### ECEF Coordinates

Entity locations in Legion use ECEF (Earth-Centered, Earth-Fixed) coordinates, not latitude/longitude. Use this conversion function:

```go
// Convert latitude, longitude, altitude to ECEF coordinates
func latLonAltToECEF(lat, lon, alt float64) (x, y, z float64) {
    // WGS84 ellipsoid constants
    a := 6378137.0         // semi-major axis
    f := 1 / 298.257223563 // flattening
    
    latRad := lat * math.Pi / 180
    lonRad := lon * math.Pi / 180
    
    sinLat := math.Sin(latRad)
    cosLat := math.Cos(latRad)
    
    N := a / math.Sqrt(1 - f*(2-f)*sinLat*sinLat)
    
    x = (N + alt) * cosLat * math.Cos(lonRad)
    y = (N + alt) * cosLat * math.Sin(lonRad)
    z = (N*(1-f*(2-f)) + alt) * sinLat
    
    return x, y, z
}
```

### Common Patterns

#### Periodic Updates
```go
ticker := time.NewTicker(5 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-ticker.C:
        // Perform updates
    }
}
```

#### API Helper Pointers
```go
// Helpers for the small number of pointer fields in domain models
strPtr := func(s string) *string { return &s }
categoryPtr := func(c models.Category) *models.Category { return &c }

req := &models.CreateEntityRequest{
    Name: strPtr("My Entity"),
    // ...
}
```

## Environment Variables

### Authentication
- `LEGION_API_KEY` - API key for authentication (when using API key auth)
- `LEGION_URL` - Override Legion API URL

### Configuration via .env File

The CLI supports `.env` files for easier development. Create a `.env` file in your project root:

```bash
# Legion API Configuration
LEGION_URL=https://legion-staging.com
LEGION_API_KEY=your-api-key-here

# OAuth Configuration (if using OAuth instead of API key)
LEGION_EMAIL=your-email@example.com
LEGION_PASSWORD=your-password

# Default Organization ID for simulations
LEGION_ORG_ID=your-organization-id-here

# Simulation Parameter Defaults (shown in prompts)
DEFAULT_NUM_ENTITIES=3
DEFAULT_ENTITY_TYPE=Camera
DEFAULT_UPDATE_INTERVAL=5s
DEFAULT_DURATION=5m

# Override specific simulation parameters (skip prompts)
LEGION_ORGANIZATION_ID=ecc2dce2-b664-4077-b34c-ea89e1fb045e
```

Environment variable precedence:
- `LEGION_URL` and `LEGION_API_KEY` skip the environment selection prompt
- `DEFAULT_*` variables set default values for prompts (user can still change them)
- `LEGION_*` variables override parameters entirely (no prompt shown)

### CLI Flags
- `--log-level` - Set logging level (debug, info, warn, error)
- `--no-color` - Disable colored output

## Contributing

1. Follow the existing code structure and patterns
2. Add appropriate tests for new functionality
3. Run `make lint-fix` before committing
4. Update documentation as needed

## License

[License information here]
