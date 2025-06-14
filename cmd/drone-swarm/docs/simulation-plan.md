# Drone Swarm Simulation Plan

## Objective
Simulate a coordinated swarm of UAS threats attacking a position defended by multiple Counter-UAS systems. The simulation models realistic engagement mechanics, swarm behaviors, and system limitations, running until all threats are neutralized or defenses are overwhelmed.

## Core Components
- **Counter-UAS Systems (Friendly Entities)**: Static, ground-based defensive systems that detect and engage threats using kinetic or electronic warfare methods.
- **UAS Threats (Hostile Entities)**: Mobile, airborne entities that coordinate attacks using swarm intelligence and varied approach tactics.

## Simulation Parameters
- **User-Defined**:
  - `num_counter_uas_systems`: Number of Counter-UAS defensive systems.
  - `num_uas_threats`: Number of drones in the swarm.
  - `engagement_type_mix`: Percentage of kinetic vs electronic warfare systems (default: 70% kinetic, 30% EW).
- **System-Defined**:
  - `update_interval`: 3 seconds.
  - `detection_radius_km`: 10 km.
  - `engagement_radius_km`: 5 km.
  - `kinetic_success_rate`: 0.7-0.9 (randomly assigned per system).
  - `ew_success_rate`: 0.5-0.7 (randomly assigned per system).
  - `cooldown_seconds`: 5-10 (varies by engagement type).

## Entity Definitions & Lifecycles

### Counter-UAS System
- **Type**: `CounterUAS`
- **Location**: Placed in defensive positions around the protected area.
- **Metadata**:
  - `detection_radius_km`: 10
  - `engagement_radius_km`: 5
  - `engagement_type`: 'kinetic' or 'electronic_warfare'
  - `ammo_capacity`: 5 (for kinetic systems only)
  - `success_rate`: 0.7-0.9 (kinetic) or 0.5-0.7 (EW)
  - `cooldown_remaining`: 0 (seconds until next engagement)
  - `total_engagements`: 0
  - `successful_engagements`: 0
- **Status Lifecycle**:
  - `IDLE`: Operational and scanning for targets.
  - `TRACKING`: Target detected, waiting for engagement range.
  - `ENGAGING`: Actively engaging a target.
  - `COOLDOWN`: Post-engagement cooldown period.
  - `DEPLETED`: Out of ammo (kinetic) or disabled.

### UAS Threat
- **Type**: `UAS`
- **Location**: Initial placement in attack formations 12km from target, using coordinated entry vectors.
- **Metadata**:
  - `speed_kph`: 50-200 (varies by drone type)
  - `autonomy_level`: 0.0-1.0 (affects jamming resistance)
  - `evasion_capability`: true/false
  - `formation_role`: 'leader', 'follower', or 'scout'
  - `attack_vector`: Assigned approach angle
  - `wave_number`: 1-3 (for coordinated wave attacks)
- **Status Lifecycle**:
  - `FORMING`: Initial swarm coordination phase.
  - `INBOUND`: Moving towards target using assigned vector.
  - `DETECTED`: Entered detection radius, may begin evasion.
  - `TARGETED`: Being tracked by a Counter-UAS system.
  - `UNDER_FIRE`: Actively being engaged.
  - `JAMMED`: Affected by electronic warfare (if autonomy < 0.5).
  - `EVADING`: Performing evasive maneuvers.
  - `ELIMINATED`: Successfully destroyed.
  - `MISSION_COMPLETE`: Reached target (simulation failure).

## Simulation Logic (Executed every 3 seconds)

1. **Swarm Coordination Phase**:
   - UAS drones coordinate into attack formations based on wave assignments.
   - Leaders designated for each wave, followers maintain formation.
   - Scouts move independently for reconnaissance.

2. **Movement Phase**:
   - Drones move according to formation role and wave timing.
   - Wave 1 launches immediately, subsequent waves delayed by 30-60 seconds.
   - Evasive maneuvers triggered when under fire (if capable).
   - Jammed drones with low autonomy spiral or hover in place.

3. **Detection Phase**:
   - Counter-UAS systems scan detection radius (10km).
   - Prioritize targets by: proximity, speed, and formation role.
   - Update drone status from `INBOUND` to `DETECTED`.
   - Systems transition from `IDLE` to `TRACKING`.

4. **Engagement Phase**:
   - Systems with targets in engagement range (5km) and not in cooldown engage.
   - Kinetic systems: Direct fire with success probability.
   - EW systems: Attempt jamming based on target autonomy level.
   - Multiple systems may engage high-priority targets.

5. **Resolution Phase**:
   - Calculate engagement success based on system type and success rate.
   - Successful kinetic hits: Target `ELIMINATED`.
   - Successful jamming: Low-autonomy drones become `JAMMED`.
   - Failed engagements: Target continues mission.
   - Update system cooldowns and ammo counts.
   - Systems return to `IDLE` or `COOLDOWN` state.

## Termination Conditions
- **Success**: All UAS entities have status `ELIMINATED` or `JAMMED`.
- **Failure**: Any UAS entity reaches status `MISSION_COMPLETE` (breached defenses).
- **Stalemate**: All Counter-UAS systems are `DEPLETED` with active threats remaining.

## Implementation Details

- **Location Representation**: Use ECEF coordinates for all entity positions.
- **Distance Calculations**: 3D Euclidean distance for detection and engagement ranges.
- **Engagement Probability**: Randomized success based on system-specific rates.
- **Swarm Intelligence**: 
  - Coordinated wave attacks with 30-60 second intervals.
  - Formation maintenance with leader-follower dynamics.
  - Distributed approach vectors to overwhelm defenses.
- **Realistic Constraints**:
  - Kinetic systems limited by ammunition.
  - All systems subject to cooldown periods.
  - EW effectiveness depends on target autonomy.
- **Performance Metrics**:
  - Track engagement success rates per system type.
  - Monitor time to neutralize threats.
  - Record any defensive breaches.

## Technical Architecture

### Concurrent Processing Architecture
- **Goroutine-based Design**: Each Counter-UAS system operates independently in its own goroutine
- **Thread-Safe Updates**: Mutex-protected entity states with channel-based coordination
- **Batched API Updates**: Buffered updates to prevent overwhelming Legion API
- **Worker Pools**: Configurable goroutine pools for parallel processing

### Core Modules
1. **simulation.go**: Main orchestrator with concurrent system management
2. **entities.go**: Thread-safe entity definitions and state management
3. **controllers/**:
   - `simulation_controller.go`: Central coordination and API management
   - `system_controller.go`: Independent Counter-UAS system logic
   - `swarm_controller.go`: Drone swarm coordination and formations
4. **core/**:
   - `engagement_calculator.go`: Thread-safe engagement resolution
   - `swarm_behavior.go`: Formation and movement algorithms
   - `update_buffer.go`: Batched API update management
5. **config/**:
   - `config.go`: Configuration structures and validation
   - `loader.go`: Multi-source configuration loading
6. **reporting/**:
   - `logger.go`: Real-time console logging with levels
   - `aar_generator.go`: After Action Report generation

### Configuration System

#### Config File (config.yaml)
```yaml
simulation:
  name: "drone-swarm"
  update_interval: 3s
  
performance:
  worker_pool_size: 10
  batch_size: 50
  api_rate_limit: 100
  
swarm_config:
  formation_type: "distributed"  # distributed, concentrated, waves
  wave_delay: 45s
  autonomy_distribution: "mixed"
  evasion_probability: 0.7
  
defense_config:
  placement_pattern: "ring"  # ring, cluster, line
  engagement_rules: "closest"  # closest, highest_threat, distributed
  kinetic_ratio: 0.7
  
logging:
  console_level: "info"  # debug, info, warn, error
  enable_aar: true
  aar_format: "detailed"  # summary, detailed, full
```

#### CLI Overrides
```bash
./bin/legion-sim run drone-swarm \
  --config ./custom-config.yaml \
  --set performance.worker_pool_size=20 \
  --set logging.console_level=debug
```

### Logging and Reporting

#### Real-time Console Logging
- **Levels**: DEBUG, INFO, WARN, ERROR with color coding
- **Event Types**:
  - System initialization and configuration
  - Wave launches and formation changes
  - Detection events with threat assessment
  - Engagement attempts with outcomes
  - System status changes (cooldown, depletion)
  - Critical events (breaches, eliminations)

#### After Action Report (AAR)
Generated at simulation end with sections:
1. **Executive Summary**
   - Final outcome (Success/Failure/Stalemate)
   - Duration and forces involved
   - Key statistics

2. **Engagement Timeline**
   - Chronological event log with timestamps
   - Wave arrival times
   - Critical engagement moments

3. **System Performance Analysis**
   - Per-system statistics:
     - Total/successful engagements
     - Hit rate percentage
     - Ammunition usage
     - Time in each state

4. **Threat Analysis**
   - Wave composition and timing
   - Successful penetrations
   - Evasion effectiveness
   - Jamming resistance rates

5. **Detailed Event Log**
   - Complete timestamped event history
   - System state transitions
   - Engagement details

6. **Summary Statistics**
   - Total threats neutralized
   - Defense effectiveness rate
   - Average engagement time
   - Resource utilization

### Key Algorithms
- **Target Priority Score**: `distance_weight * (1/distance) + speed_weight * speed + role_weight`
- **Engagement Success**: `random() < system.success_rate`
- **Jamming Effectiveness**: `success if (target.autonomy < 0.5 && random() < system.success_rate)`
- **Concurrent Engagement Resolution**: Mutex-protected target assignment with channel-based coordination