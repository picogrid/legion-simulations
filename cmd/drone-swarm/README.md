# Drone Swarm Combat Simulation

A sophisticated Counter-UAS simulation featuring coordinated drone swarms attacking defensive systems.

## Overview

This simulation models:
- **Counter-UAS Systems**: Defensive units with kinetic and electronic warfare capabilities
- **UAS Threats**: Attacking drones with swarm behaviors and evasion capabilities
- **Engagement Mechanics**: Detection, tracking, targeting, and engagement with realistic success rates
- **Wave Attacks**: Coordinated multi-wave assault patterns
- **After Action Reports**: Detailed analysis of simulation outcomes

## Quick Start

```bash
# Run with interactive prompts (recommended for first time)
./bin/legion-sim run -s "Drone Swarm Combat"

# Run with example configurations
./cmd/drone-swarm/examples/run-examples.sh

# Run with specific parameter file
./bin/legion-sim run -s "Drone Swarm Combat" -p cmd/drone-swarm/examples/large-scale-battle.yaml
```

## Configuration

### Interactive Configuration (Default)
The simulation will prompt for all parameters with sensible defaults:
- Counter-UAS Systems: 10
- UAS Threats: 50
- Waves: 5
- Update Interval: 1s
- Duration: 2m

### Environment Variables
Set defaults for prompts:
```bash
export LEGION_NUM_COUNTER_UAS_SYSTEMS=20
export LEGION_NUM_UAS_THREATS=100
export LEGION_WAVES=10
export LEGION_DURATION=5m
```

### Parameter Files
See `examples/` directory for pre-configured scenarios:
- `params-example.yaml` - Basic configuration example
- `large-scale-battle.yaml` - 100 threats vs 20 defenders
- `defensive-test.yaml` - Testing defensive capabilities

### Automation Mode
Skip all prompts for CI/CD:
```bash
export LEGION_SKIP_PROMPTS=true
export LEGION_NUM_COUNTER_UAS_SYSTEMS=10
export LEGION_NUM_UAS_THREATS=50
./bin/legion-sim run -s "Drone Swarm Combat"
```

## Simulation Mechanics

### Counter-UAS Systems
- **Detection Radius**: 10km
- **Engagement Radius**: 5km
- **Types**:
  - Kinetic: Higher success rate (70-90%), limited ammo
  - Electronic Warfare: Lower success rate (50-70%), unlimited uses

### UAS Threats
- **Speed**: 50-200 kph (randomized)
- **Autonomy Level**: 0.0-1.0 (affects targeting difficulty)
- **Evasion**: 70% have evasion capabilities
- **Formation Roles**: Leader, Scout, Follower

### Engagement Phases
1. **Swarm Coordination**: Formation keeping and wave coordination
2. **Movement**: Threats advance toward base, evasive maneuvers when under fire
3. **Detection**: Counter-UAS systems detect threats within range
4. **Engagement**: Systems engage targets within range with success probability
5. **Resolution**: Update statistics, check victory conditions

## Output

### Real-time Updates
- System status (IDLE, TRACKING, ENGAGING, COOLDOWN, DEPLETED)
- Threat status (FORMING, INBOUND, DETECTED, TARGETED, UNDER_FIRE, ELIMINATED)
- Engagement results with hit/miss indicators
- Running statistics

### After Action Report
Generated in `reports/` directory:
- Engagement statistics
- System performance metrics
- Threat analysis
- Timeline of events
- Recommendations

## Examples

### Interactive Demo
```bash
./cmd/drone-swarm/examples/demo-interactive.sh
```

### Run Examples Menu
```bash
./cmd/drone-swarm/examples/run-examples.sh
```

### Custom Scenario
```bash
# Set your parameters
export LEGION_NUM_COUNTER_UAS_SYSTEMS=15
export LEGION_NUM_UAS_THREATS=75
export LEGION_WAVES=5
export LEGION_UPDATE_INTERVAL=500ms
export LEGION_DURATION=3m

# Run simulation
./bin/legion-sim run -s "Drone Swarm Combat"
```

## Development

### Directory Structure
```
cmd/drone-swarm/
├── README.md              # This file
├── simulation.yaml        # Simulation configuration
├── main.go               # Entry point
├── simulation/           # Core simulation logic
├── controllers/          # Simulation controllers
├── core/                # Core mechanics (engagement, swarm behavior)
├── reporting/           # AAR generation
├── examples/            # Example configurations and scripts
└── docs/                # Additional documentation
```

### Adding New Features
1. Modify engagement mechanics in `core/engagement_calculator.go`
2. Add new behaviors in `core/swarm_behavior.go`
3. Update entity definitions in `simulation/entities.go`
4. Extend reporting in `reporting/aar_generator.go`

## Troubleshooting

### Common Issues

**409 Entity Already Exists**
- Solution: Set `cleanup_existing: true` or use unique names

**No Engagements Occurring**
- Check spawn distance vs engagement radius
- Verify threats are moving toward base
- Enable debug logging: `export LEGION_LOG_LEVEL=debug`

**Simulation Running Too Long**
- Reduce duration: `export LEGION_DURATION=1m`
- Increase update interval: `export LEGION_UPDATE_INTERVAL=2s`

### Debug Mode
```bash
export LEGION_LOG_LEVEL=debug
export LEGION_ENABLE_AAR=true
./bin/legion-sim run -s "Drone Swarm Combat"
```