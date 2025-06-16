# Simulation Configuration Guide

The Legion Simulations framework provides flexible configuration options through multiple methods:

## Configuration Hierarchy

1. **Interactive CLI Prompts** (default)
   - User is prompted for each parameter
   - Shows current defaults
   - Allows modification of any value

2. **Environment Variables** 
   - Set defaults for CLI prompts
   - Use `LEGION_<PARAMETER_NAME>` format
   - Values appear as defaults in prompts

3. **Skip Prompts Mode**
   - Set `LEGION_SKIP_PROMPTS=true` for automation
   - Uses environment variables or defaults
   - No interactive prompts

## Environment Variable Reference

### Authentication & Organization
- `LEGION_URL` - Legion API URL
- `LEGION_API_KEY` - API key for authentication
- `LEGION_EMAIL` - Email for OAuth authentication
- `LEGION_PASSWORD` - Password for OAuth authentication
- `LEGION_ORG_ID` / `LEGION_ORGANIZATION_ID` - Organization ID

### Drone Swarm Simulation Parameters
- `LEGION_NUM_COUNTER_UAS_SYSTEMS` - Number of defensive systems (default: 10)
- `LEGION_NUM_UAS_THREATS` - Number of attacking drones (default: 50)
- `LEGION_WAVES` - Number of attack waves (default: 5)
- `LEGION_UPDATE_INTERVAL` - Update frequency (default: 1s)
- `LEGION_DURATION` - Maximum duration (default: 2m)
- `LEGION_CLEANUP_EXISTING` - Clean up entities before start (default: true)
- `LEGION_LOG_LEVEL` - Logging level (default: info)
- `LEGION_ENABLE_AAR` - Generate After Action Report (default: true)

## Usage Examples

### Interactive Mode (Default)
```bash
# Run with prompts, using defaults from simulation.yaml
./bin/legion-sim run -s "Drone Swarm Combat"

# Environment variables set defaults for prompts
export LEGION_NUM_UAS_THREATS=100
./bin/legion-sim run -s "Drone Swarm Combat"
# The prompt will show 100 as the default for threats
```

### Using Parameter Files
```bash
# Run with specific configuration
./bin/legion-sim run -s "Drone Swarm Combat" -p cmd/drone-swarm/examples/large-scale-battle.yaml

# Parameter files override all prompts
```

### Automated Mode (CI/CD)
```bash
# Skip all prompts, use environment variables
export LEGION_SKIP_PROMPTS=true
export LEGION_NUM_COUNTER_UAS_SYSTEMS=20
export LEGION_NUM_UAS_THREATS=100
export LEGION_WAVES=10
export LEGION_UPDATE_INTERVAL=500ms
export LEGION_DURATION=5m

./bin/legion-sim run -s "Drone Swarm Combat"
```

### Using .env File
Create a `.env` file in your project directory:
```env
# Authentication
LEGION_URL=https://legion-staging.com
LEGION_EMAIL=user@example.com
LEGION_PASSWORD=your-password
LEGION_ORG_ID=your-org-id

# Simulation defaults (will appear in prompts)
LEGION_NUM_COUNTER_UAS_SYSTEMS=15
LEGION_NUM_UAS_THREATS=75
LEGION_WAVES=5

# Skip prompts for automation
# LEGION_SKIP_PROMPTS=true
```

## Best Practices

1. **Development**: Use interactive mode to experiment with different configurations
2. **Testing**: Set environment variables for consistent test scenarios
3. **CI/CD**: Use `LEGION_SKIP_PROMPTS=true` with full environment configuration
4. **Documentation**: Always document the configuration used for important simulations

## Configuration Priority

When `LEGION_SKIP_PROMPTS=true`:
1. Environment variable value (if set)
2. Default from simulation.yaml
3. Error if required and no value available

When prompting (default):
1. User input from prompt
2. Default shown in prompt is from environment variable (if set)
3. Default shown in prompt is from simulation.yaml (if no env var)