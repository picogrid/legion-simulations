# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Legion-simulations is a Go-based simulation framework by PicoGrid for modeling various scenarios including drone swarms, weather events, satellite operations, and more. Built on OpenAPI-driven architecture with a complete infrastructure setup.

## Common Development Commands

### Environment Setup
```bash
make dev-env          # Complete macOS setup (brew, pre-commit, Go tools)
make dev-brew         # Install golangci-lint, pre-commit
make dev-precommit    # Configure git hooks
make dev-gotooling    # Install Go development tools

# Optional: Set up .env file for easier development
cp .env.example .env
# Edit .env with your credentials and preferences
```

### Environment Variables (.env support)

The CLI automatically loads `.env` files from the current directory. You can use this to set default values and credentials:

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

When environment variables are set:
- `LEGION_URL` and `LEGION_API_KEY` skip the environment selection prompt
- `DEFAULT_*` variables set default values for prompts (user can still change them)
- `LEGION_*` variables override parameters entirely (no prompt shown)

### Build and Run
```bash
make build            # Build the CLI binary
make run              # Run simulations interactively
make list             # List available simulations
make rebuild          # Clean and rebuild
make deps             # Update Go dependencies

# Direct CLI usage
./bin/legion-sim run
./bin/legion-sim list
./bin/legion-sim env add      # Add Legion environment
./bin/legion-sim env list     # List configured environments
./bin/legion-sim env remove   # Remove environment
```

### Testing
```bash
make test             # Run all unit tests (required by pre-push hook)
make test-verbose     # Run tests with verbose output
make unit-test        # Run unit tests with JSON output

# Run specific tests
go test -v -cover -run TestName ./pkg/...
go test -v -cover ./...  # All tests with coverage
```

### Code Quality
```bash
make lint             # Run linter without fixes
make lint-fix         # Run linter with auto-fix (used by pre-commit)
make check            # Run both lint and tests

# Direct golangci-lint usage
golangci-lint run --fix --timeout 5m --config=.golangci.yaml
golangci-lint run ./pkg/...  # Specific package
```

### Model Generation
```bash
# Generate models from OpenAPI spec
swagger generate model -f openapi.yaml -t pkg/ --model-package=models --strict-responders

# Generate client from OpenAPI spec (if using generated client)
swagger generate client -f openapi.yaml -t pkg/ --client-package=legion --skip-models
```

## Architecture

### High-Level Structure
```
/cmd/               # Executable applications
  cli/              # Main CLI tool for running simulations
  simple/           # Example simulation implementation
  drone-swarm/      # Drone coordination simulation (planned)
  weather-events/   # Weather pattern modeling (planned)
  satellite-ops/    # Satellite operations simulation (planned)

/pkg/               # Shared packages
  client/           # Hand-written Legion API client
  auth/             # Authentication (Keycloak, token management)
  models/           # Generated OpenAPI models (DO NOT EDIT)
  legion/           # Generated OpenAPI client (optional)
  simulation/       # Core simulation framework and registry
  config/           # Environment configuration management
  utils/            # Shared utilities and helpers
  logger/           # Structured logging utilities
```

### Authentication Flow

The system supports two authentication methods:

1. **OAuth2 (Keycloak)**
   - Dynamically fetches auth URL from Legion API endpoint `/v3/integrations/oauth/authorization-url`
   - Falls back to hardcoded URL if API call fails
   - Handles interactive login with email/password
   - Automatic token refresh via TokenManager

2. **API Key**
   - Environment variable based (e.g., `LEGION_DEV_API_KEY`)
   - Suitable for CI/CD and automation

Key files:
- `pkg/auth/legion_auth.go` - Fetches auth config from Legion API
- `pkg/auth/keycloak.go` - Keycloak client implementation
- `pkg/auth/token_manager.go` - Token lifecycle management
- `pkg/auth/helpers.go` - Authentication utilities

### Simulation Framework

All simulations implement the `simulation.Simulation` interface:
```go
type Simulation interface {
    Name() string
    Description() string
    Configure(params map[string]interface{}) error
    Run(ctx context.Context, client *client.Legion) error
    Stop() error
}
```

Simulations are:
- Discovered via YAML configuration files
- Registered in init() functions
- Run through the CLI with parameter prompting
- Support graceful shutdown via context cancellation

### Legion API Client

The project uses a hand-written client (`pkg/client/`) organized by domain:
- `client.go` - Core HTTP client with auth support
- `entities.go` - Entity CRUD operations
- `locations.go` - Entity location management (ECEF coordinates)
- `organizations.go` - Organization management
- `feeds.go` - Feed definitions and data ingestion
- `users.go` - User operations
- `helpers.go` - Helper utilities for API operations

## Key Technical Details

### ECEF Coordinates
Entity locations use Earth-Centered, Earth-Fixed (ECEF) coordinates, not lat/lon. Use the conversion function in `cmd/simple/simulation.go`:
```go
func latLonAltToECEF(lat, lon, alt float64) (x, y, z float64)
```

### Pre-commit Hooks
- **pre-commit**: Runs `golangci-lint run --fix` on staged files
- **pre-push**: Runs `make unit-test` to ensure tests pass

### Generated vs Hand-written Code
- Models are generated from `openapi.yaml` using go-swagger
- Client is hand-written for simplicity and maintainability
- Never edit files in `pkg/models/` - they're regenerated

### Environment Configuration
Environments are stored in `~/.legion/config.yaml` with structure:
```yaml
environments:
  - name: "dev"
    url: "https://api.legion-dev.com"
    apiKey: "LEGION_DEV_API_KEY"  # Name of env var
    authMethod: "oauth"  # or "apikey"
```

## Important Implementation Notes

1. **Authorization URL**: The CLI now fetches the Keycloak auth URL from Legion's `/v3/integrations/oauth/authorization-url` endpoint instead of hardcoding it. See `pkg/auth/legion_auth.go`.

2. **Simulation Registration**: New simulations must be imported in `cmd/cli/cmd/run.go` to trigger their init() functions.

3. **Error Handling**: The client returns wrapped errors with context. Always wrap errors when propagating them up.

4. **Logging**: Use the `pkg/logger` package for structured logging with proper levels (Info, Error, Warn, Progress, Success).

5. **Model Generation**: After updating `openapi.yaml`, regenerate models and handle any breaking changes in the hand-written client.

## CI/CD Pipeline

GitHub Actions workflows execute in sequence:
1. Linting (`.github/workflows/lint.yml`)
2. Unit Tests (`.github/workflows/unit-tests.yml`)
3. Pull Request orchestration (`.github/workflows/pull_request.yml`)

## Enabled Linters

From `.golangci.yaml`:
- gocritic - Opinionated checks
- govet - Suspicious constructs
- ineffassign - Unused assignments
- staticcheck - Advanced static analysis
- unconvert - Unnecessary conversions
- gofmt/goimports - Formatting and imports