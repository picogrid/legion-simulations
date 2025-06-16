package controllers

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/cmd/drone-swarm/core"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
)

// SwarmController manages UAS threat coordination and wave management
type SwarmController struct {
	uasThreats     map[uuid.UUID]*UASThreat
	waves          []*WaveState
	formations     map[string]Formation
	targetLocation *models.GeomPoint
	updateBuffer   *core.UpdateBuffer
	mu             sync.RWMutex
}

// WaveState represents the state of a UAS wave
type WaveState struct {
	WaveNumber    int
	Threats       []uuid.UUID
	FormationType string // "distributed", "concentrated", "line"
	Status        string // "forming", "launched", "engaged", "complete"
	CenterOfMass  Vector3D
	LeaderID      *uuid.UUID
	LaunchTime    time.Time
}

// Vector3D represents a 3D vector
type Vector3D struct {
	X, Y, Z float64
}

// DroneState represents the current state of a drone
type DroneState struct {
	ID         uuid.UUID
	Type       string
	Position   Vector3D
	Velocity   Vector3D
	Health     float64
	Ammunition int
	FuelLevel  float64
	TargetID   *uuid.UUID
	Role       string // "leader", "follower", "scout", etc.
	TeamName   string
}

// Formation defines swarm formations
type Formation interface {
	GetTargetPosition(droneIndex int, centerOfMass Vector3D, heading float64) Vector3D
	GetSpacing() float64
}

// Objective defines team objectives
type Objective interface {
	GetPriority() int
	GetTargetLocation() *Vector3D
	IsComplete() bool
	Update(wave *WaveState)
}

// Common formations
type VFormation struct {
	Spacing float64
	Angle   float64
}

func (f *VFormation) GetSpacing() float64 { return f.Spacing }

type WedgeFormation struct {
	Spacing float64
	Depth   float64
}

func (f *WedgeFormation) GetSpacing() float64 { return f.Spacing }

type LineFormation struct {
	Spacing float64
}

func (f *LineFormation) GetSpacing() float64 { return f.Spacing }

type DistributedFormation struct {
	MinSpacing float64
	MaxSpacing float64
}

func (f *DistributedFormation) GetSpacing() float64 { return f.MinSpacing }

// NewSwarmController creates a new swarm controller
func NewSwarmController() *SwarmController {
	sc := &SwarmController{
		uasThreats: make(map[uuid.UUID]*UASThreat),
		waves:      make([]*WaveState, 0),
		formations: make(map[string]Formation),
	}

	// Register default formations
	sc.formations["v"] = &VFormation{Spacing: 50.0, Angle: 45.0}
	sc.formations["wedge"] = &WedgeFormation{Spacing: 40.0, Depth: 30.0}
	sc.formations["line"] = &LineFormation{Spacing: 60.0}
	sc.formations["distributed"] = &DistributedFormation{MinSpacing: 100.0, MaxSpacing: 200.0}

	return sc
}

// Initialize sets up the swarm controller
func (sc *SwarmController) Initialize(ctx context.Context, teams []string) error {
	logger.Info("Initializing swarm controller...")
	return nil
}

// SetThreats sets the UAS threats managed by the swarm controller
func (sc *SwarmController) SetThreats(threats map[uuid.UUID]*UASThreat) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.uasThreats = threats
}

// SetTargetLocation sets the target location for the swarm
func (sc *SwarmController) SetTargetLocation(location *models.GeomPoint) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.targetLocation = location
}

// SetUpdateBuffer sets the update buffer for entity updates
func (sc *SwarmController) SetUpdateBuffer(buffer *core.UpdateBuffer) {
	sc.updateBuffer = buffer
}

// InitializeWave initializes a new wave of threats
func (sc *SwarmController) InitializeWave(waveNumber int, threatIDs []uuid.UUID, formationType string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	wave := &WaveState{
		WaveNumber:    waveNumber,
		Threats:       threatIDs,
		FormationType: formationType,
		Status:        "forming",
	}

	// Select leader (first threat in wave)
	if len(threatIDs) > 0 {
		wave.LeaderID = &threatIDs[0]
	}

	sc.waves = append(sc.waves, wave)
}

// Update performs swarm behavior updates
func (sc *SwarmController) Update(ctx context.Context, deltaTime float64) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Update each wave
	for _, wave := range sc.waves {
		if wave.Status != "launched" {
			continue
		}

		// Update wave center of mass
		sc.updateWaveCenterOfMass(wave)

		// Apply formation to wave
		if formation, exists := sc.formations[wave.FormationType]; exists {
			sc.applyFormation(wave, formation, deltaTime)
		}

		// Apply swarm behavior
		sc.applySwarmBehavior(wave, deltaTime)
	}

	return nil
}

// LaunchWave transitions a wave from forming to launched
func (sc *SwarmController) LaunchWave(waveNumber int) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, wave := range sc.waves {
		if wave.WaveNumber == waveNumber && wave.Status == "forming" {
			wave.Status = "launched"
			wave.LaunchTime = time.Now()
			logger.Infof("Wave %d launched with %d threats", waveNumber+1, len(wave.Threats))
			break
		}
	}
}

// AddDrone adds a drone to the swarm controller
func (sc *SwarmController) AddDrone(teamName string, drone *DroneState) error {
	// In this implementation, drones are managed through the UASThreat map
	// This method is kept for interface compatibility
	return nil
}

// updateWaveCenterOfMass calculates the center of mass for a wave
func (sc *SwarmController) updateWaveCenterOfMass(wave *WaveState) {
	if len(wave.Threats) == 0 {
		return
	}

	var sumX, sumY, sumZ float64
	var count int

	for _, threatID := range wave.Threats {
		threat, exists := sc.uasThreats[threatID]
		if !exists || threat.Status == UASStatusEliminated || threat.Status == UASStatusJammed {
			continue
		}

		if threat.Position != nil && len(threat.Position.Coordinates) >= 3 {
			sumX += threat.Position.Coordinates[0]
			sumY += threat.Position.Coordinates[1]
			sumZ += threat.Position.Coordinates[2]
			count++
		}
	}

	if count > 0 {
		wave.CenterOfMass = Vector3D{
			X: sumX / float64(count),
			Y: sumY / float64(count),
			Z: sumZ / float64(count),
		}
	}
}

// applyFormation applies formation keeping forces to a wave
func (sc *SwarmController) applyFormation(wave *WaveState, formation Formation, deltaTime float64) {
	if len(wave.Threats) < 2 {
		return
	}

	// Calculate average heading
	var avgVelX, avgVelY float64
	var count int

	for _, threatID := range wave.Threats {
		threat, exists := sc.uasThreats[threatID]
		if !exists || threat.Status == UASStatusEliminated || threat.Status == UASStatusJammed {
			continue
		}

		if threat.Velocity != nil && len(threat.Velocity.Coordinates) >= 2 {
			avgVelX += threat.Velocity.Coordinates[0]
			avgVelY += threat.Velocity.Coordinates[1]
			count++
		}
	}

	if count == 0 {
		return
	}

	heading := math.Atan2(avgVelY/float64(count), avgVelX/float64(count))

	// Apply formation positions
	validThreats := 0
	for i, threatID := range wave.Threats {
		threat, exists := sc.uasThreats[threatID]
		if !exists || threat.Status == UASStatusEliminated || threat.Status == UASStatusJammed {
			continue
		}

		// Get target position in formation
		targetPos := formation.GetTargetPosition(validThreats, wave.CenterOfMass, heading)
		validThreats++

		// Apply formation keeping force
		if threat.Position != nil && len(threat.Position.Coordinates) >= 3 {
			forceX := targetPos.X - threat.Position.Coordinates[0]
			forceY := targetPos.Y - threat.Position.Coordinates[1]
			forceZ := targetPos.Z - threat.Position.Coordinates[2]

			// Scale force
			forceMagnitude := 5.0 // Formation keeping strength

			// Update velocity
			if threat.Velocity != nil && len(threat.Velocity.Coordinates) >= 3 {
				threat.Velocity.Coordinates[0] += forceX * forceMagnitude * deltaTime / 1000.0
				threat.Velocity.Coordinates[1] += forceY * forceMagnitude * deltaTime / 1000.0
				threat.Velocity.Coordinates[2] += forceZ * forceMagnitude * deltaTime / 1000.0
			}
		}

		// Special handling for leader
		if wave.LeaderID != nil && threatID == *wave.LeaderID && i == 0 {
			// Leader maintains course toward target
			if sc.targetLocation != nil && threat.Position != nil {
				dirX := sc.targetLocation.Coordinates[0] - threat.Position.Coordinates[0]
				dirY := sc.targetLocation.Coordinates[1] - threat.Position.Coordinates[1]
				dirZ := sc.targetLocation.Coordinates[2] - threat.Position.Coordinates[2]

				// Normalize direction
				mag := math.Sqrt(dirX*dirX + dirY*dirY + dirZ*dirZ)
				if mag > 0 {
					dirX /= mag
					dirY /= mag
					dirZ /= mag

					// Set leader velocity
					speed := threat.SpeedKph / 3.6 // Convert to m/s
					threat.Velocity.Coordinates[0] = dirX * speed
					threat.Velocity.Coordinates[1] = dirY * speed
					threat.Velocity.Coordinates[2] = dirZ * speed
				}
			}
		}
	}
}

// applySwarmBehavior applies swarm intelligence behaviors
func (sc *SwarmController) applySwarmBehavior(wave *WaveState, deltaTime float64) {
	// Implement basic flocking behavior
	for _, threatID := range wave.Threats {
		threat, exists := sc.uasThreats[threatID]
		if !exists || threat.Status == UASStatusEliminated || threat.Status == UASStatusJammed {
			continue
		}

		// Apply evasion if under fire
		if threat.Status == UASStatusUnderFire && threat.EvasionCapability {
			sc.applyEvasionBehavior(threat, deltaTime)
		}

		// Apply separation to avoid collisions
		sc.applySeparation(threat, wave, deltaTime)

		// Apply cohesion to stay together
		sc.applyCohesion(threat, wave, deltaTime)

		// Apply alignment to match velocities
		sc.applyAlignment(threat, wave, deltaTime)
	}
}

// applyEvasionBehavior applies evasive maneuvers
func (sc *SwarmController) applyEvasionBehavior(threat *UASThreat, deltaTime float64) {
	if threat.Velocity == nil || len(threat.Velocity.Coordinates) < 3 {
		return
	}

	// Random evasive maneuver
	evasionForce := 10.0
	threat.Velocity.Coordinates[0] += (rand.Float64()*2 - 1) * evasionForce
	threat.Velocity.Coordinates[1] += (rand.Float64()*2 - 1) * evasionForce
	threat.Velocity.Coordinates[2] += (rand.Float64()*0.5 - 0.25) * evasionForce // Less vertical evasion

	// Update status
	threat.Status = UASStatusEvading
	if sc.updateBuffer != nil {
		sc.updateBuffer.QueueStatusUpdate(threat.ID, threat.Status)
	}
}

// applySeparation applies separation force to avoid collisions
func (sc *SwarmController) applySeparation(threat *UASThreat, wave *WaveState, deltaTime float64) {
	if threat.Position == nil || threat.Velocity == nil {
		return
	}

	separationRadius := 50.0 // meters
	var forceX, forceY, forceZ float64

	for _, otherID := range wave.Threats {
		if otherID == threat.ID {
			continue
		}

		other, exists := sc.uasThreats[otherID]
		if !exists || other.Position == nil {
			continue
		}

		dx := threat.Position.Coordinates[0] - other.Position.Coordinates[0]
		dy := threat.Position.Coordinates[1] - other.Position.Coordinates[1]
		dz := threat.Position.Coordinates[2] - other.Position.Coordinates[2]
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)

		if dist < separationRadius && dist > 0 {
			// Apply repulsive force
			force := (separationRadius - dist) / separationRadius
			forceX += (dx / dist) * force * 5.0
			forceY += (dy / dist) * force * 5.0
			forceZ += (dz / dist) * force * 5.0
		}
	}

	// Apply separation force
	threat.Velocity.Coordinates[0] += forceX * deltaTime
	threat.Velocity.Coordinates[1] += forceY * deltaTime
	threat.Velocity.Coordinates[2] += forceZ * deltaTime
}

// applyCohesion applies cohesion force to stay together
func (sc *SwarmController) applyCohesion(threat *UASThreat, wave *WaveState, deltaTime float64) {
	if threat.Position == nil || threat.Velocity == nil {
		return
	}

	// Pull towards center of mass
	dx := wave.CenterOfMass.X - threat.Position.Coordinates[0]
	dy := wave.CenterOfMass.Y - threat.Position.Coordinates[1]
	dz := wave.CenterOfMass.Z - threat.Position.Coordinates[2]

	// Apply cohesion force (weaker than separation)
	cohesionStrength := 0.5
	threat.Velocity.Coordinates[0] += dx * cohesionStrength * deltaTime / 1000.0
	threat.Velocity.Coordinates[1] += dy * cohesionStrength * deltaTime / 1000.0
	threat.Velocity.Coordinates[2] += dz * cohesionStrength * deltaTime / 1000.0
}

// applyAlignment applies alignment force to match velocities
func (sc *SwarmController) applyAlignment(threat *UASThreat, wave *WaveState, deltaTime float64) {
	if threat.Velocity == nil {
		return
	}

	var avgVelX, avgVelY, avgVelZ float64
	var count int

	for _, otherID := range wave.Threats {
		if otherID == threat.ID {
			continue
		}

		other, exists := sc.uasThreats[otherID]
		if !exists || other.Velocity == nil {
			continue
		}

		avgVelX += other.Velocity.Coordinates[0]
		avgVelY += other.Velocity.Coordinates[1]
		avgVelZ += other.Velocity.Coordinates[2]
		count++
	}

	if count > 0 {
		avgVelX /= float64(count)
		avgVelY /= float64(count)
		avgVelZ /= float64(count)

		// Apply alignment force
		alignmentStrength := 0.1
		threat.Velocity.Coordinates[0] += (avgVelX - threat.Velocity.Coordinates[0]) * alignmentStrength * deltaTime
		threat.Velocity.Coordinates[1] += (avgVelY - threat.Velocity.Coordinates[1]) * alignmentStrength * deltaTime
		threat.Velocity.Coordinates[2] += (avgVelZ - threat.Velocity.Coordinates[2]) * alignmentStrength * deltaTime
	}
}

// Formation implementations
func (f *VFormation) GetTargetPosition(index int, center Vector3D, heading float64) Vector3D {
	row := index / 2
	side := index % 2

	offsetX := float64(row) * f.Spacing * math.Cos(heading)
	offsetY := float64(row) * f.Spacing * math.Sin(heading)

	if side == 1 {
		offsetX += f.Spacing * math.Cos(heading+f.Angle*math.Pi/180)
		offsetY += f.Spacing * math.Sin(heading+f.Angle*math.Pi/180)
	} else if index > 0 {
		offsetX += f.Spacing * math.Cos(heading-f.Angle*math.Pi/180)
		offsetY += f.Spacing * math.Sin(heading-f.Angle*math.Pi/180)
	}

	return Vector3D{
		X: center.X + offsetX,
		Y: center.Y + offsetY,
		Z: center.Z,
	}
}

func (f *WedgeFormation) GetTargetPosition(index int, center Vector3D, heading float64) Vector3D {
	// Wedge formation - wider at the back
	row := int(math.Sqrt(float64(index * 2)))
	posInRow := index - (row * (row + 1) / 2)

	offsetForward := float64(row) * f.Depth
	offsetSide := float64(posInRow-(row/2)) * f.Spacing

	offsetX := offsetForward*math.Cos(heading) - offsetSide*math.Sin(heading)
	offsetY := offsetForward*math.Sin(heading) + offsetSide*math.Cos(heading)

	return Vector3D{
		X: center.X + offsetX,
		Y: center.Y + offsetY,
		Z: center.Z,
	}
}

func (f *LineFormation) GetTargetPosition(index int, center Vector3D, heading float64) Vector3D {
	// Line formation perpendicular to heading
	offset := float64(index) * f.Spacing

	offsetX := -offset * math.Sin(heading)
	offsetY := offset * math.Cos(heading)

	return Vector3D{
		X: center.X + offsetX,
		Y: center.Y + offsetY,
		Z: center.Z,
	}
}

func (f *DistributedFormation) GetTargetPosition(index int, center Vector3D, heading float64) Vector3D {
	// Random distributed formation
	angle := rand.Float64() * 2 * math.Pi
	distance := f.MinSpacing + rand.Float64()*(f.MaxSpacing-f.MinSpacing)

	return Vector3D{
		X: center.X + distance*math.Cos(angle),
		Y: center.Y + distance*math.Sin(angle),
		Z: center.Z + (rand.Float64()*100 - 50), // Random altitude variation
	}
}

// GetWaveStatus returns the current status of all waves
func (sc *SwarmController) GetWaveStatus() map[int]string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	status := make(map[int]string)
	for _, wave := range sc.waves {
		status[wave.WaveNumber] = wave.Status
	}
	return status
}

// GetActiveThreatsCount returns the number of active threats
func (sc *SwarmController) GetActiveThreatsCount() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	count := 0
	for _, threat := range sc.uasThreats {
		if threat.Status != UASStatusEliminated &&
			threat.Status != UASStatusJammed &&
			threat.Status != UASStatusMissionComplete {
			count++
		}
	}
	return count
}
