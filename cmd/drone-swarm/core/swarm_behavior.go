package core

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SwarmBehaviorEngine manages UAS threat swarm behaviors for attacking defended positions
type SwarmBehaviorEngine struct {
	behaviors       map[string]Behavior
	activeBehaviors map[string]string // team -> active behavior
	behaviorWeights map[string]float64
	waveStatus      map[int]*WaveStatus // wave number -> status
	mu              sync.RWMutex
}

// WaveStatus tracks the status of each attack wave
type WaveStatus struct {
	WaveNumber   int
	LaunchTime   time.Time
	DronesInWave []*Drone
	Status       string // "forming", "launched", "engaged", "complete"
}

// Behavior represents a swarm behavior pattern
type Behavior interface {
	Calculate(swarm *Swarm, environment *Environment) []Force
	GetPriority() float64
	IsApplicable(swarm *Swarm, environment *Environment) bool
}

// Swarm represents a coordinated group of UAS threats
type Swarm struct {
	ID             string
	TeamName       string
	Drones         []*Drone
	Objective      *Objective // The defended position to attack
	Formation      string     // "distributed", "concentrated", "waves"
	CenterMass     Vector3D
	AverageVel     Vector3D
	WaveDelay      time.Duration // Delay between waves (30-60 seconds)
	CurrentWave    int           // Current wave being launched
	BehaviorEngine interface{}   // Reference to the behavior engine
}

// Drone represents an individual UAS threat in the swarm
type Drone struct {
	ID             uuid.UUID
	Type           string
	Position       Vector3D
	Velocity       Vector3D
	Health         float64
	Status         string     // FORMING, INBOUND, DETECTED, TARGETED, UNDER_FIRE, JAMMED, EVADING, ELIMINATED, MISSION_COMPLETE
	TargetID       *uuid.UUID // ID of Counter-UAS system targeting this drone
	Role           string     // "leader", "follower", "scout"
	WaveNumber     int        // 1-3 for coordinated wave attacks
	AttackVector   float64    // Approach angle in radians
	SpeedKPH       float64    // Speed in km/h (50-200)
	AutonomyLevel  float64    // 0.0-1.0 (affects jamming resistance)
	EvasionCapable bool       // Can perform evasive maneuvers
	IsJammed       bool       // Currently affected by EW
	Neighbors      []*Drone
	LastUpdate     time.Time
	mu             sync.RWMutex
}

// Vector3D represents a 3D vector (ECEF coordinates)
type Vector3D struct {
	X, Y, Z float64
}

// Force represents a behavioral force applied to drones
type Force struct {
	DroneID  uuid.UUID
	Force    Vector3D
	Priority float64
	Behavior string
}

// Environment contains information about the battlefield
type Environment struct {
	DefendedPosition  Vector3D            // The position being defended
	CounterUASSystems []*CounterUASSystem // Defensive systems
	Threats           []Threat            // Active threat zones (engagement areas)
	JammingZones      []JammingZone       // EW affected areas
	TerrainHeight     func(x, y float64) float64
}

// CounterUASSystem represents a defensive system
type CounterUASSystem struct {
	ID               uuid.UUID
	Position         Vector3D
	EngagementType   string  // "kinetic" or "electronic_warfare"
	EngagementRadius float64 // in meters
	Status           string  // IDLE, TRACKING, ENGAGING, COOLDOWN, DEPLETED
}

// Threat represents an active engagement zone
type Threat struct {
	Position Vector3D
	Radius   float64
	Severity float64
	Type     string // "kinetic_fire", "ew_jamming"
}

// JammingZone represents an area affected by electronic warfare
type JammingZone struct {
	Position Vector3D
	Radius   float64
	Strength float64 // 0.0-1.0
}

// Objective represents the defended position to attack
type Objective struct {
	ID       string
	Position Vector3D
	Type     string
	Priority float64
	Status   string
}

// NewSwarmBehaviorEngine creates a new behavior engine for UAS threat swarms
func NewSwarmBehaviorEngine() *SwarmBehaviorEngine {
	engine := &SwarmBehaviorEngine{
		behaviors:       make(map[string]Behavior),
		activeBehaviors: make(map[string]string),
		behaviorWeights: make(map[string]float64),
		waveStatus:      make(map[int]*WaveStatus),
	}

	// Register UAS threat-specific behaviors
	engine.registerThreatBehaviors()

	return engine
}

// registerThreatBehaviors sets up UAS threat swarm behaviors
func (e *SwarmBehaviorEngine) registerThreatBehaviors() {
	// Core swarm coordination behaviors
	e.behaviors["separation"] = &SeparationBehavior{Weight: 1.5, MinDistance: 30.0}
	e.behaviors["cohesion"] = &CohesionBehavior{Weight: 1.0}
	e.behaviors["alignment"] = &AlignmentBehavior{Weight: 1.2}

	// Attack-specific behaviors
	e.behaviors["wave_attack"] = &WaveAttackBehavior{Weight: 3.0}
	e.behaviors["attack_vector"] = &AttackVectorBehavior{Weight: 2.5}
	e.behaviors["objective_approach"] = &ObjectiveApproachBehavior{Weight: 2.0}

	// Defensive behaviors
	e.behaviors["evasion"] = &EvasionBehavior{Weight: 4.0}
	e.behaviors["jamming_response"] = &JammingResponseBehavior{Weight: 3.5}

	// Formation behaviors
	e.behaviors["formation"] = &FormationBehavior{Weight: 2.0}
	e.behaviors["role_based"] = &RoleBasedBehavior{Weight: 2.2}

	// Set default weights
	e.behaviorWeights["separation"] = 1.0
	e.behaviorWeights["cohesion"] = 1.0
	e.behaviorWeights["alignment"] = 1.0
	e.behaviorWeights["wave_attack"] = 1.5
	e.behaviorWeights["attack_vector"] = 1.3
	e.behaviorWeights["objective_approach"] = 1.2
	e.behaviorWeights["evasion"] = 2.0
	e.behaviorWeights["jamming_response"] = 1.8
	e.behaviorWeights["formation"] = 1.1
	e.behaviorWeights["role_based"] = 1.2
}

// CalculateForces computes all behavioral forces for a swarm
func (e *SwarmBehaviorEngine) CalculateForces(swarm *Swarm, environment *Environment) []Force {
	e.mu.Lock()
	defer e.mu.Unlock()

	var allForces []Force

	// Update wave status
	e.updateWaveStatus(swarm)

	// Update swarm metrics
	e.UpdateSwarmMetrics(swarm)

	// Update neighbor information
	e.updateNeighbors(swarm)

	// Calculate forces from each applicable behavior
	for name, behavior := range e.behaviors {
		if behavior.IsApplicable(swarm, environment) {
			forces := behavior.Calculate(swarm, environment)

			// Tag forces with behavior name
			for i := range forces {
				forces[i].Behavior = name
			}

			allForces = append(allForces, forces...)
		}
	}

	// Combine and prioritize forces
	return e.combineForces(allForces)
}

// updateWaveStatus manages the wave attack coordination
func (e *SwarmBehaviorEngine) updateWaveStatus(swarm *Swarm) {
	now := time.Now()

	// Initialize wave status if needed
	for waveNum := 1; waveNum <= 3; waveNum++ {
		if _, exists := e.waveStatus[waveNum]; !exists {
			e.waveStatus[waveNum] = &WaveStatus{
				WaveNumber:   waveNum,
				Status:       "forming",
				DronesInWave: make([]*Drone, 0),
			}

			// Assign drones to waves
			for _, drone := range swarm.Drones {
				if drone.WaveNumber == waveNum {
					e.waveStatus[waveNum].DronesInWave = append(e.waveStatus[waveNum].DronesInWave, drone)
				}
			}
		}
	}

	// Check wave launch times
	for waveNum, status := range e.waveStatus {
		if status.Status == "forming" {
			// Wave 1 launches immediately
			if waveNum == 1 {
				status.LaunchTime = now
				status.Status = "launched"
				swarm.CurrentWave = 1
			} else {
				// Check if previous wave has launched and delay has passed
				prevWave := e.waveStatus[waveNum-1]
				if prevWave.Status != "forming" && now.Sub(prevWave.LaunchTime) >= swarm.WaveDelay {
					status.LaunchTime = now
					status.Status = "launched"
					swarm.CurrentWave = waveNum
				}
			}
		}
	}
}

// IsWaveLaunched checks if a specific wave has been launched
func (e *SwarmBehaviorEngine) IsWaveLaunched(waveNumber int) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if status, exists := e.waveStatus[waveNumber]; exists {
		return status.Status != "forming"
	}
	return false
}

// updateNeighbors updates neighbor lists for all drones
func (e *SwarmBehaviorEngine) updateNeighbors(swarm *Swarm) {
	neighborRadius := 100.0 // meters

	for i, drone := range swarm.Drones {
		drone.Neighbors = nil

		for j, other := range swarm.Drones {
			if i != j {
				dist := drone.Position.DistanceTo(other.Position)
				if dist < neighborRadius {
					drone.Neighbors = append(drone.Neighbors, other)
				}
			}
		}
	}
}

// combineForces merges and weights multiple forces
func (e *SwarmBehaviorEngine) combineForces(forces []Force) []Force {
	// Group forces by drone
	droneForces := make(map[uuid.UUID][]Force)
	for _, force := range forces {
		droneForces[force.DroneID] = append(droneForces[force.DroneID], force)
	}

	// Combine forces for each drone
	var combined []Force
	for droneID, forceList := range droneForces {
		if len(forceList) == 0 {
			continue
		}

		// Weight and sum forces
		var totalForce Vector3D
		var totalWeight float64

		for _, f := range forceList {
			weight := f.Priority * e.behaviorWeights[f.Behavior]
			totalForce = totalForce.Add(f.Force.Scale(weight))
			totalWeight += weight
		}

		// Normalize by total weight
		if totalWeight > 0 {
			totalForce = totalForce.Scale(1.0 / totalWeight)
		}

		combined = append(combined, Force{
			DroneID:  droneID,
			Force:    totalForce,
			Priority: 1.0,
			Behavior: "combined",
		})
	}

	return combined
}

// SeparationBehavior avoids crowding between drones
type SeparationBehavior struct {
	Weight      float64
	MinDistance float64
}

func (b *SeparationBehavior) GetPriority() float64 { return b.Weight }

func (b *SeparationBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	return true // Always active
}

func (b *SeparationBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	for _, drone := range swarm.Drones {
		// Skip eliminated or jammed drones
		drone.mu.RLock()
		if drone.Status == "ELIMINATED" || (drone.Status == "JAMMED" && !drone.EvasionCapable) {
			drone.mu.RUnlock()
			continue
		}
		drone.mu.RUnlock()

		var separationForce Vector3D

		for _, neighbor := range drone.Neighbors {
			diff := drone.Position.Subtract(neighbor.Position)
			dist := diff.Magnitude()

			if dist > 0 && dist < b.MinDistance {
				// Repulsion force inversely proportional to distance
				force := diff.Normalize().Scale(b.MinDistance / dist)
				separationForce = separationForce.Add(force)
			}
		}

		if separationForce.Magnitude() > 0 {
			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    separationForce.Normalize(),
				Priority: b.Weight,
			})
		}
	}

	return forces
}

// Cohesion behavior - steer towards center of neighbors
type CohesionBehavior struct {
	Weight float64
}

func (b *CohesionBehavior) GetPriority() float64 { return b.Weight }

func (b *CohesionBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	return true // Always active
}

func (b *CohesionBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	for _, drone := range swarm.Drones {
		if len(drone.Neighbors) == 0 {
			continue
		}

		// Calculate center of neighbors
		var center Vector3D
		for _, neighbor := range drone.Neighbors {
			center = center.Add(neighbor.Position)
		}
		center = center.Scale(1.0 / float64(len(drone.Neighbors)))

		// Steer towards center
		cohesionForce := center.Subtract(drone.Position)

		if cohesionForce.Magnitude() > 0 {
			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    cohesionForce.Normalize(),
				Priority: b.Weight,
			})
		}
	}

	return forces
}

// Alignment behavior - match velocity with neighbors
type AlignmentBehavior struct {
	Weight float64
}

func (b *AlignmentBehavior) GetPriority() float64 { return b.Weight }

func (b *AlignmentBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	return true // Always active
}

func (b *AlignmentBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	for _, drone := range swarm.Drones {
		if len(drone.Neighbors) == 0 {
			continue
		}

		// Calculate average velocity of neighbors
		var avgVelocity Vector3D
		for _, neighbor := range drone.Neighbors {
			avgVelocity = avgVelocity.Add(neighbor.Velocity)
		}
		avgVelocity = avgVelocity.Scale(1.0 / float64(len(drone.Neighbors)))

		// Steer towards average velocity
		alignmentForce := avgVelocity.Subtract(drone.Velocity)

		if alignmentForce.Magnitude() > 0 {
			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    alignmentForce.Normalize(),
				Priority: b.Weight,
			})
		}
	}

	return forces
}

// WaveAttackBehavior coordinates multi-wave attacks on defended position
type WaveAttackBehavior struct {
	Weight float64
}

func (b *WaveAttackBehavior) GetPriority() float64 { return b.Weight }

func (b *WaveAttackBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	return swarm.Formation == "waves"
}

func (b *WaveAttackBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		waveNum := drone.WaveNumber
		status := drone.Status
		drone.mu.RUnlock()

		// Only apply forces to drones in launched waves
		if status == "ELIMINATED" || status == "MISSION_COMPLETE" {
			continue
		}

		// Check if this drone's wave has launched
		engine := swarm.BehaviorEngine
		if engine != nil && !engine.(*SwarmBehaviorEngine).IsWaveLaunched(waveNum) {
			continue // Stay in formation until wave launches
		}

		// Move towards objective
		force := env.DefendedPosition.Subtract(drone.Position)
		if force.Magnitude() > 0 {
			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    force.Normalize(),
				Priority: b.Weight,
			})
		}
	}

	return forces
}

// AttackVectorBehavior maintains assigned attack vectors for distributed approach
type AttackVectorBehavior struct {
	Weight float64
}

func (b *AttackVectorBehavior) GetPriority() float64 { return b.Weight }

func (b *AttackVectorBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	return true
}

func (b *AttackVectorBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		vector := drone.AttackVector
		status := drone.Status
		drone.mu.RUnlock()

		if status == "FORMING" || status == "ELIMINATED" {
			continue
		}

		// Calculate lateral offset based on attack vector
		objectiveDir := env.DefendedPosition.Subtract(drone.Position).Normalize()

		// Create perpendicular vector for curved approach
		perpX := -objectiveDir.Y
		perpY := objectiveDir.X
		perpZ := 0.0
		perp := Vector3D{X: perpX, Y: perpY, Z: perpZ}.Normalize()

		// Blend direct and curved approach based on distance
		dist := drone.Position.DistanceTo(env.DefendedPosition)
		curveFactor := math.Min(dist/5000.0, 1.0) // More curve when farther away

		lateralForce := perp.Scale(math.Sin(vector) * curveFactor)
		directForce := objectiveDir.Scale(1.0 - curveFactor*0.3)

		totalForce := directForce.Add(lateralForce)

		if totalForce.Magnitude() > 0 {
			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    totalForce.Normalize(),
				Priority: b.Weight,
			})
		}
	}

	return forces
}

// EvasionBehavior performs evasive maneuvers when under fire
type EvasionBehavior struct {
	Weight float64
}

func (b *EvasionBehavior) GetPriority() float64 { return b.Weight }

func (b *EvasionBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	// Check if any drones are under fire
	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		underFire := drone.Status == "UNDER_FIRE" || drone.Status == "TARGETED"
		capable := drone.EvasionCapable
		drone.mu.RUnlock()

		if underFire && capable {
			return true
		}
	}
	return false
}

func (b *EvasionBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		status := drone.Status
		capable := drone.EvasionCapable
		drone.mu.RUnlock()

		// Only evade if capable and under threat
		if !capable || (status != "UNDER_FIRE" && status != "TARGETED") {
			continue
		}

		var evadeForce Vector3D

		// Evade from known threats
		for _, threat := range env.Threats {
			if threat.Type == "kinetic_fire" {
				dist := drone.Position.DistanceTo(threat.Position)

				if dist < threat.Radius*2 { // Start evading before entering kill zone
					// Lateral evasion perpendicular to threat
					toThreat := threat.Position.Subtract(drone.Position).Normalize()

					// Random evasion direction
					evadeDir := Vector3D{
						X: -toThreat.Y + (rand.Float64()-0.5)*0.5,
						Y: toThreat.X + (rand.Float64()-0.5)*0.5,
						Z: (rand.Float64() - 0.5) * 0.3, // Some vertical evasion
					}.Normalize()

					// Stronger evasion when closer
					strength := threat.Severity * (threat.Radius*2 - dist) / (threat.Radius * 2)
					evadeForce = evadeForce.Add(evadeDir.Scale(strength))
				}
			}
		}

		// Add some randomness for unpredictability
		if status == "UNDER_FIRE" {
			randomForce := Vector3D{
				X: (rand.Float64() - 0.5),
				Y: (rand.Float64() - 0.5),
				Z: (rand.Float64() - 0.5) * 0.5,
			}.Normalize().Scale(0.3)
			evadeForce = evadeForce.Add(randomForce)
		}

		if evadeForce.Magnitude() > 0 {
			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    evadeForce.Normalize(),
				Priority: b.Weight,
			})
		}
	}

	return forces
}

// JammingResponseBehavior handles drone behavior when jammed
type JammingResponseBehavior struct {
	Weight float64
}

func (b *JammingResponseBehavior) GetPriority() float64 { return b.Weight }

func (b *JammingResponseBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	// Check if any drones are jammed
	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		jammed := drone.IsJammed || drone.Status == "JAMMED"
		drone.mu.RUnlock()

		if jammed {
			return true
		}
	}
	return false
}

func (b *JammingResponseBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		jammed := drone.IsJammed || drone.Status == "JAMMED"
		autonomy := drone.AutonomyLevel
		drone.mu.RUnlock()

		if !jammed {
			continue
		}

		// Low autonomy drones spiral or hover
		if autonomy < 0.5 {
			// Spiral pattern
			spiralForce := Vector3D{
				X: math.Cos(float64(time.Now().Unix())*0.5) * 0.3,
				Y: math.Sin(float64(time.Now().Unix())*0.5) * 0.3,
				Z: -0.1, // Slight descent
			}

			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    spiralForce,
				Priority: b.Weight,
			})
		} else if env.DefendedPosition.Magnitude() > 0 {
			// High autonomy drones try to continue mission
			continueForce := env.DefendedPosition.Subtract(drone.Position).Normalize().Scale(0.5)
			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    continueForce,
				Priority: b.Weight * 0.5, // Reduced effectiveness
			})
		}
	}

	return forces
}

// ObjectiveApproachBehavior moves drones towards the defended position
type ObjectiveApproachBehavior struct {
	Weight float64
}

func (b *ObjectiveApproachBehavior) GetPriority() float64 { return b.Weight }

func (b *ObjectiveApproachBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	return swarm.Objective != nil && env.DefendedPosition.Magnitude() > 0
}

func (b *ObjectiveApproachBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		status := drone.Status
		drone.mu.RUnlock()

		// Skip drones that can't move
		if status == "ELIMINATED" || status == "MISSION_COMPLETE" || status == "FORMING" {
			continue
		}

		// Direct approach to objective
		objectiveForce := env.DefendedPosition.Subtract(drone.Position)

		if objectiveForce.Magnitude() > 0 {
			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    objectiveForce.Normalize(),
				Priority: b.Weight,
			})
		}
	}

	return forces
}

// RoleBasedBehavior adjusts behavior based on drone role (leader/follower/scout)
type RoleBasedBehavior struct {
	Weight float64
}

func (b *RoleBasedBehavior) GetPriority() float64 { return b.Weight }

func (b *RoleBasedBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	return true
}

func (b *RoleBasedBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	// Find leaders and their followers
	leaders := make(map[int]*Drone) // wave -> leader
	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		if drone.Role == "leader" && drone.Status != "ELIMINATED" {
			leaders[drone.WaveNumber] = drone
		}
		drone.mu.RUnlock()
	}

	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		role := drone.Role
		waveNum := drone.WaveNumber
		status := drone.Status
		drone.mu.RUnlock()

		if status == "ELIMINATED" || status == "MISSION_COMPLETE" {
			continue
		}

		var roleForce Vector3D

		switch role {
		case "follower":
			// Follow the leader of your wave
			if leader, exists := leaders[waveNum]; exists {
				// Maintain formation relative to leader
				idealOffset := Vector3D{
					X: (rand.Float64() - 0.5) * 50, // 50m spread
					Y: (rand.Float64() - 0.5) * 50,
					Z: (rand.Float64() - 0.5) * 10,
				}
				idealPos := leader.Position.Add(idealOffset)
				roleForce = idealPos.Subtract(drone.Position).Scale(0.3)
			}

		case "scout":
			// Scouts move ahead and to the sides
			if swarm.CenterMass.Magnitude() > 0 {
				// Move ahead of the swarm
				toObjective := env.DefendedPosition.Subtract(swarm.CenterMass).Normalize()
				scoutOffset := toObjective.Scale(200) // 200m ahead

				// Add lateral offset
				lateral := Vector3D{
					X: -toObjective.Y,
					Y: toObjective.X,
					Z: 0,
				}.Scale((rand.Float64() - 0.5) * 100)

				idealPos := swarm.CenterMass.Add(scoutOffset).Add(lateral)
				roleForce = idealPos.Subtract(drone.Position).Scale(0.4)
			}

		case "leader":
			// Leaders push forward more aggressively
			if env.DefendedPosition.Magnitude() > 0 {
				roleForce = env.DefendedPosition.Subtract(drone.Position).Normalize().Scale(0.2)
			}
		}

		if roleForce.Magnitude() > 0 {
			forces = append(forces, Force{
				DroneID:  drone.ID,
				Force:    roleForce,
				Priority: b.Weight,
			})
		}
	}

	return forces
}

// FormationBehavior maintains swarm formations during approach
type FormationBehavior struct {
	Weight float64
}

func (b *FormationBehavior) GetPriority() float64 { return b.Weight }

func (b *FormationBehavior) IsApplicable(swarm *Swarm, env *Environment) bool {
	return swarm.Formation != "" && swarm.Formation != "waves" // Waves handled separately
}

func (b *FormationBehavior) Calculate(swarm *Swarm, env *Environment) []Force {
	var forces []Force

	switch swarm.Formation {
	case "distributed":
		// Maintain spacing while advancing
		for _, drone := range swarm.Drones {
			drone.mu.RLock()
			status := drone.Status
			drone.mu.RUnlock()

			if status == "ELIMINATED" || status == "FORMING" {
				continue
			}

			// Maintain minimum distance from neighbors
			var spacingForce Vector3D
			neighborCount := 0

			for _, neighbor := range drone.Neighbors {
				dist := drone.Position.DistanceTo(neighbor.Position)
				if dist < 100 && dist > 0 { // 100m ideal spacing
					// Push away if too close
					away := drone.Position.Subtract(neighbor.Position).Normalize()
					spacingForce = spacingForce.Add(away.Scale((100 - dist) / 100))
					neighborCount++
				}
			}

			if neighborCount > 0 && spacingForce.Magnitude() > 0 {
				forces = append(forces, Force{
					DroneID:  drone.ID,
					Force:    spacingForce.Normalize(),
					Priority: b.Weight,
				})
			}
		}

	case "concentrated":
		// Tight formation, minimal spacing
		if swarm.CenterMass.Magnitude() > 0 {
			for _, drone := range swarm.Drones {
				drone.mu.RLock()
				status := drone.Status
				drone.mu.RUnlock()

				if status == "ELIMINATED" || status == "FORMING" {
					continue
				}

				// Pull towards swarm center
				toCenter := swarm.CenterMass.Subtract(drone.Position)
				dist := toCenter.Magnitude()

				if dist > 50 { // Keep within 50m of center
					forces = append(forces, Force{
						DroneID:  drone.ID,
						Force:    toCenter.Normalize(),
						Priority: b.Weight * 0.5,
					})
				}
			}
		}
	}

	return forces
}

// Vector3D methods
func (v Vector3D) Add(other Vector3D) Vector3D {
	return Vector3D{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

func (v Vector3D) Subtract(other Vector3D) Vector3D {
	return Vector3D{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

func (v Vector3D) Scale(s float64) Vector3D {
	return Vector3D{X: v.X * s, Y: v.Y * s, Z: v.Z * s}
}

func (v Vector3D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vector3D) Normalize() Vector3D {
	mag := v.Magnitude()
	if mag == 0 {
		return v
	}
	return v.Scale(1.0 / mag)
}

func (v Vector3D) DistanceTo(other Vector3D) float64 {
	return v.Subtract(other).Magnitude()
}

// UpdateSwarmMetrics calculates center of mass and average velocity
func (e *SwarmBehaviorEngine) UpdateSwarmMetrics(swarm *Swarm) {
	if len(swarm.Drones) == 0 {
		return
	}

	var centerMass Vector3D
	var avgVel Vector3D
	activeCount := 0

	for _, drone := range swarm.Drones {
		drone.mu.RLock()
		status := drone.Status
		pos := drone.Position
		vel := drone.Velocity
		drone.mu.RUnlock()

		if status != "ELIMINATED" && status != "MISSION_COMPLETE" {
			centerMass = centerMass.Add(pos)
			avgVel = avgVel.Add(vel)
			activeCount++
		}
	}

	if activeCount > 0 {
		swarm.CenterMass = centerMass.Scale(1.0 / float64(activeCount))
		swarm.AverageVel = avgVel.Scale(1.0 / float64(activeCount))
	}
}

// GetWaveStatus returns the status of a specific wave
func (e *SwarmBehaviorEngine) GetWaveStatus(waveNumber int) *WaveStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.waveStatus[waveNumber]
}

// SetStatus safely updates a drone's status
func (d *Drone) SetStatus(status string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.Status = status
	d.LastUpdate = time.Now()
}

// GetStatus safely retrieves a drone's status
func (d *Drone) GetStatus() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.Status
}

// ApplyForce updates drone velocity based on combined forces
func (d *Drone) ApplyForce(force Vector3D, deltaTime float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Skip if drone can't move
	if d.Status == "ELIMINATED" || d.Status == "MISSION_COMPLETE" {
		return
	}

	// Apply force as acceleration (F=ma, assuming mass=1)
	acceleration := force

	// Update velocity
	d.Velocity = d.Velocity.Add(acceleration.Scale(deltaTime))

	// Limit to max speed
	maxSpeed := d.SpeedKPH / 3.6 // Convert km/h to m/s
	if d.Velocity.Magnitude() > maxSpeed {
		d.Velocity = d.Velocity.Normalize().Scale(maxSpeed)
	}

	// Update position
	d.Position = d.Position.Add(d.Velocity.Scale(deltaTime))
	d.LastUpdate = time.Now()
}

// DistanceToObjective calculates distance to the defended position
func (d *Drone) DistanceToObjective(objective Vector3D) float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.Position.DistanceTo(objective)
}
