# Drone Swarm Combat - Quick Start Guide

## Running Your First Simulation

### Option 1: Interactive Mode (Recommended)
```bash
./bin/legion-sim run -s "Drone Swarm Combat"
```
You'll be prompted for each parameter with defaults shown.

### Option 2: Use Example Menu
```bash
./cmd/drone-swarm/examples/run-examples.sh
```
Select from pre-configured scenarios.

### Option 3: Environment Variables
```bash
export LEGION_NUM_UAS_THREATS=100
export LEGION_NUM_COUNTER_UAS_SYSTEMS=20
./bin/legion-sim run -s "Drone Swarm Combat"
```

## What You'll See

1. **Entity Creation**: Systems and threats are created in Legion
2. **Deployment**: Forces positioned on the battlefield
3. **Real-time Combat**:
   - 👁️ Detection events
   - 🎯 Engagement attempts
   - 💥 Successful eliminations
   - ❌ Missed shots
4. **Statistics**: Live updates on engagements and active units
5. **After Action Report**: Generated in `reports/` directory

## Key Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| Counter-UAS Systems | 10 | Defensive units |
| UAS Threats | 50 | Attacking drones |
| Waves | 5 | Attack waves |
| Duration | 2m | Max simulation time |
| Update Interval | 1s | Update frequency |

## Tips

- Start with defaults to see a balanced battle
- Increase threats for overwhelming attack scenarios
- Enable debug logging for detailed mechanics: `export LEGION_LOG_LEVEL=debug`
- Use `cleanup_existing: true` to avoid entity conflicts