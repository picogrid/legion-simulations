# Counter-UAS Configuration System

This package provides a comprehensive configuration system for the Counter-UAS vs UAS threat simulation. It supports loading from YAML files, environment variable overrides, and CLI parameter overrides with a clear hierarchy of precedence.

## Configuration Hierarchy

Configuration values are applied in this order (later values override earlier ones):

1. **Default values** (hardcoded defaults)
2. **YAML file** (config.yaml)
3. **Environment variables** (runtime overrides)
4. **CLI parameters** (command-line overrides)

## Usage Examples

### Basic Configuration Loading

```go
import "github.com/picogrid/legion-simulations/cmd/drone-swarm/config"

// Load with automatic fallback to defaults
config, err := config.LoadConfigOrDefault("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Load with both environment and CLI overrides
cliOverrides := map[string]interface{}{
    "num_counter_uas_systems": 10,
    "formation_type": "waves",
}
config, err := config.LoadConfigWithOverrides("config.yaml", cliOverrides)
```

### Environment Variable Overrides

| Environment Variable | Description | Example |
|---------------------|-------------|---------|
| `LEGION_ORGANIZATION_ID` | Legion organization ID | `org-123-456` |
| `NUM_COUNTER_UAS_SYSTEMS` | Number of Counter-UAS systems | `8` |
| `NUM_UAS_THREATS` | Number of UAS threats | `25` |
| `CENTER_LATITUDE` | Center latitude for simulation | `37.7749` |
| `CENTER_LONGITUDE` | Center longitude for simulation | `-122.4194` |
| `ENGAGEMENT_TYPE_MIX` | Kinetic vs EW ratio (0.0-1.0) | `0.8` |
| `SWARM_FORMATION_TYPE` | Formation pattern | `distributed`, `concentrated`, `waves` |
| `DEFENSE_PLACEMENT_PATTERN` | Defense placement | `ring`, `cluster`, `line` |
| `LOG_LEVEL` | Logging level | `debug`, `info`, `warn`, `error` |
| `ENABLE_AAR` | Enable After Action Reports | `true`, `false` |
| `AAR_OUTPUT_PATH` | AAR output directory | `./reports/` |
| `VERBOSE_LOGGING` | Enable verbose logging | `true`, `false` |

### CLI Override Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `num_counter_uas_systems` | int | Number of Counter-UAS systems |
| `num_uas_threats` | int | Number of UAS threats |
| `engagement_type_mix` | float64 | Kinetic ratio (0.0-1.0) |
| `formation_type` | string | `distributed`, `concentrated`, `waves` |
| `placement_pattern` | string | `ring`, `cluster`, `line` |
| `wave_count` | int | Number of attack waves |
| `wave_delay` | time.Duration | Delay between waves |
| `autonomy_distribution` | string | `low`, `mixed`, `high` |
| `evasion_probability` | float64 | Evasion probability (0.0-1.0) |
| `success_rate_modifier` | float64 | Difficulty modifier |
| `verbose_logging` | bool | Enable verbose logging |
| `enable_aar` | bool | Enable After Action Reports |
| `log_level` | string | Logging level |

## Configuration Structure

### Core Simulation Settings
- **Name**: Simulation identifier
- **Description**: Human-readable description
- **Update Interval**: How often to update entities

### Entity Configuration
- **Counter-UAS Systems**: Number of defensive systems
- **UAS Threats**: Number of attacking drones

### Swarm Configuration
- **Formation Type**: How UAS threats are organized
- **Wave Count/Delay**: Attack wave parameters
- **Autonomy Distribution**: AI capability levels
- **Evasion Probability**: Chance to avoid engagement
- **Speed Range**: Min/max speeds in kph

### Defense Configuration
- **Placement Pattern**: How Counter-UAS systems are positioned
- **Engagement Rules**: Target selection strategy
- **Kinetic Ratio**: Proportion of kinetic vs EW systems
- **Detection/Engagement Radius**: Sensor and weapon ranges
- **Cooldown Ranges**: Reload/recharge times

### Engagement Parameters
- **Success Rate Ranges**: Effectiveness of kinetic and EW systems
- **Ammo Capacity**: Kinetic system ammunition
- **Jamming Threshold**: Autonomy level susceptible to EW

### Target Prioritization
- **Distance Weight**: Importance of target proximity
- **Speed Weight**: Importance of target velocity
- **Role Weight**: Importance of target type
- **Role Multipliers**: Priority modifiers by role

### Termination Conditions
- **Success**: All threats neutralized
- **Failure**: Defensive breach (any UAS reaches objective)
- **Stalemate**: All systems depleted with active threats

## Validation

The configuration system includes comprehensive validation:

```go
if err := config.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}
```

Validation checks:
- Required fields are present
- Numeric values are within valid ranges
- Probability values are between 0.0 and 1.0
- Enum values are from valid sets
- Range minimums are less than maximums

## Example Configuration Files

See `config.yaml` for a complete example configuration that matches the Counter-UAS simulation plan. The configuration includes:

- Performance tuning parameters
- Swarm behavior settings
- Defense system configuration
- Engagement mechanics
- Logging and reporting options
- Victory/defeat conditions

## Testing

Run the configuration tests:

```bash
go test -v ./cmd/drone-swarm/config/
```

This validates:
- YAML file loading
- Default configuration
- Environment variable overrides
- CLI parameter overrides
- Configuration validation
- Error handling