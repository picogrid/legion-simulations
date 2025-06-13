# Simple Entity Test Simulation

A basic simulation that demonstrates the Legion simulation framework by creating and managing entities with periodic location updates.

## Overview

The Simple Entity Test simulation creates a specified number of entities in Legion and continuously updates their locations at regular intervals. This simulation serves as a foundational example for building more complex simulations.

## Features

- Creates 1-5 entities of a specified type (Camera, Drone, Vehicle, or Sensor)
- Updates entity locations at configurable intervals
- Simulates movement in a circular pattern around a center point
- Runs for a specified duration
- Proper cleanup of entities when simulation stops

## Configuration

The simulation accepts the following parameters:

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `num_entities` | integer | Number of entities to create (1-5) | 3 |
| `entity_type` | string | Type of entity (Camera, Drone, Vehicle, Sensor) | Camera |
| `update_interval` | duration | How often to update entity locations | 5s |
| `duration` | duration | How long to run the simulation | 5m |
| `organization_id` | string | Organization ID for entity creation | (required) |

## Usage

### Interactive Mode

```bash
./bin/legion-sim run
```

Then select "Simple Entity Test" from the menu and follow the prompts.

### Command Line Mode

```bash
./bin/legion-sim run --simulation "Simple Entity Test"
```

### With Parameters File

Create a YAML file with your parameters:

```yaml
num_entities: 5
entity_type: "Drone"
update_interval: "10s"
duration: "10m"
organization_id: "your-org-id-here"
```

Then run:

```bash
./bin/legion-sim run --simulation "Simple Entity Test" --params params.yaml
```

### Using Environment Variables

You can use a `.env` file or environment variables to set defaults or override parameters:

```bash
# Copy the example file
cp .env.example .env

# Edit .env with your values:
# LEGION_URL=https://legion-staging.com
# LEGION_API_KEY=your-api-key
# LEGION_ORG_ID=your-org-id
# DEFAULT_NUM_ENTITIES=3
# DEFAULT_ENTITY_TYPE=Camera

# Run the simulation - it will use the defaults from .env
./bin/legion-sim run --simulation "Simple Entity Test"
```

Environment variable precedence:
1. `LEGION_ORGANIZATION_ID` - Skips the organization ID prompt entirely
2. `DEFAULT_*` variables - Set default values in prompts (user can change)
3. Command line parameters override everything

## Location Updates

The simulation updates entity locations using Earth-Centered, Earth-Fixed (ECEF) coordinates. Entities move in a circular pattern around a central point (New York City by default) with the following behavior:

- Each entity has a unique angular offset based on its index
- Entities rotate around the center point at a constant angular velocity
- The radius of movement is 1000 meters from the center
- Updates are sent to Legion at the configured interval

## Implementation Details

### Entity Creation

Entities are created with:
- A unique name following the pattern `{type}-{index}` (e.g., `Camera-001`)
- Category set to "device"
- Status set to "active"
- Metadata containing the entity index and simulation name

### Organization ID Header

The simulation properly sets the `X-ORG-ID` header for all API calls using the Legion client's context support:

```go
ctx = client.WithOrgID(ctx, s.config.OrganizationID)
```

### Coordinate System

The simulation uses the following coordinate transformations:
- **Center Point**: 40.7128�N, 74.0060�W (New York City)
- **Movement**: Circular pattern at 1000m radius
- **Conversion**: Latitude/Longitude/Altitude to ECEF coordinates

### Error Handling

The simulation includes robust error handling:
- Validates configuration parameters
- Handles entity creation failures
- Manages context cancellation
- Performs cleanup on shutdown

## Example Output

```
21:37:34 INFO  Starting Simple Entity Test simulation with 3 Camera entities
21:37:35 INFO  Created entity: Camera-001 (ID: abc123...)
21:37:35 INFO  Created entity: Camera-002 (ID: def456...)
21:37:35 INFO  Created entity: Camera-003 (ID: ghi789...)
21:37:40 INFO  Updated locations for 3 entities
21:37:45 INFO  Updated locations for 3 entities
...
21:42:34 INFO  Simulation completed after 5m0s
```

## Development

### Adding New Entity Types

To add support for new entity types:

1. Update the `simulation.yaml` file to include the new type in the allowed values
2. Consider if the new type requires different movement patterns or behaviors

### Extending Functionality

This simulation can be extended to:
- Support different movement patterns (linear, random, following paths)
- Include entity attributes or telemetry data
- Simulate entity interactions
- Add environmental factors

### Testing

When testing the simulation:
1. Start with a small number of entities (1-2)
2. Use short update intervals to quickly see results
3. Monitor the Legion dashboard to verify entities are created and updated
4. Check logs for any errors or warnings

## Troubleshooting

### Common Issues

1. **"Organization ID is required" error**
   - Ensure you've provided a valid organization ID when prompted
   - Verify the Legion client is properly setting the X-ORG-ID header

2. **"Failed to create entity" errors**
   - Check your authentication credentials
   - Verify the organization ID has permission to create entities
   - Ensure the entity type is valid

3. **No location updates visible**
   - Check that the update interval isn't too long
   - Verify entities were created successfully
   - Look for errors in the update loop

### Debug Mode

Run with debug logging to see detailed information:

```bash
./bin/legion-sim run --log-level debug --simulation "Simple Entity Test"
```

## See Also

- [Legion Simulations README](../../README.md)
- [Simulation Framework](../../pkg/simulation/README.md)
- [Legion API Client](../../pkg/client/README.md)