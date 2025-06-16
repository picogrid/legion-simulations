# Simulation Development Template

When creating a new simulation, organize it following this structure:

## Directory Layout

```
cmd/<simulation-name>/
├── README.md              # Detailed simulation documentation
├── simulation.yaml        # Parameter definitions and metadata
├── main.go               # Entry point with init() registration
├── simulation/           # Core simulation logic
│   └── simulation.go     # Main simulation implementation
├── examples/             # Example configurations and scripts
│   ├── params-example.yaml
│   ├── scenario-1.yaml
│   ├── scenario-2.yaml
│   ├── demo-interactive.sh
│   └── run-examples.sh
├── docs/                 # Additional documentation
│   ├── QUICK_START.md
│   └── TECHNICAL.md
└── [optional directories]
    ├── controllers/      # If using controller pattern
    ├── core/            # Core mechanics/algorithms
    ├── reporting/       # Custom reporting/AAR
    └── config/          # Configuration structures
```

## File Templates

### simulation.yaml
```yaml
name: "Simulation Name"
description: "Brief description"
author: "Your Name"
version: "1.0.0"
category: "category"

parameters:
  - name: "organization_id"
    type: "string"
    description: "Legion organization ID"
    required: true
    env: "LEGION_ORGANIZATION_ID"
  
  - name: "param1"
    type: "integer"
    description: "Description"
    default: 10
    env: "LEGION_PARAM1"
  # ... more parameters
```

### examples/run-examples.sh
```bash
#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR/../../.."

echo "======================================"
echo "Simulation Name Examples"
echo "======================================"
# ... menu implementation
```

### examples/demo-interactive.sh
```bash
#!/bin/bash
# Interactive demonstration of configuration options
```

## Best Practices

1. **Organization**
   - Keep all simulation-specific files in the simulation directory
   - Use `examples/` for parameter files and demo scripts
   - Use `docs/` for additional documentation

2. **Configuration**
   - Support both interactive prompts and parameter files
   - Use environment variables for defaults, not overrides
   - Document all parameters clearly

3. **Examples**
   - Provide at least 2-3 example scenarios
   - Include an interactive demo script
   - Create a run-examples.sh menu for easy access

4. **Documentation**
   - README.md with complete overview
   - QUICK_START.md for getting started fast
   - Include troubleshooting section

5. **Paths**
   - Use relative paths in scripts: `../../../bin/legion-sim`
   - Make scripts executable: `chmod +x`

## Example: Drone Swarm

See `cmd/drone-swarm/` for a complete implementation following this template.