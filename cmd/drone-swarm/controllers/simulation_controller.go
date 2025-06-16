package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/cmd/drone-swarm/core"
	"github.com/picogrid/legion-simulations/cmd/drone-swarm/reporting"
	"github.com/picogrid/legion-simulations/pkg/client"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
)

// Entity types
const (
	EntityTypeCounterUAS = "CounterUAS"
	EntityTypeUAS        = "UAS"
)

// Counter-UAS System Status Lifecycle
const (
	CounterUASStatusIdle     = "IDLE"
	CounterUASStatusTracking = "TRACKING"
	CounterUASStatusEngaging = "ENGAGING"
	CounterUASStatusCooldown = "COOLDOWN"
	CounterUASStatusDepleted = "DEPLETED"
)

// UAS Threat Status Lifecycle
const (
	UASStatusForming         = "FORMING"
	UASStatusInbound         = "INBOUND"
	UASStatusDetected        = "DETECTED"
	UASStatusTargeted        = "TARGETED"
	UASStatusUnderFire       = "UNDER_FIRE"
	UASStatusJammed          = "JAMMED"
	UASStatusEvading         = "EVADING"
	UASStatusEliminated      = "ELIMINATED"
	UASStatusMissionComplete = "MISSION_COMPLETE"
)

// SimulationController manages the overall simulation lifecycle
type SimulationController struct {
	legionClient      *client.Legion
	organizationID    string
	config            *SimulationConfig
	counterUASSystems map[uuid.UUID]*CounterUASSystem
	uasThreats        map[uuid.UUID]*UASThreat
	systemController  *SystemController
	swarmController   *SwarmController
	updateBuffer      *core.UpdateBuffer
	engagementCalc    *core.EngagementCalculator
	simLogger         *reporting.SimulationLogger
	startTime         time.Time
	endTime           time.Time
	isRunning         atomic.Bool
	mu                sync.RWMutex
	stopChan          chan struct{}
	wg                sync.WaitGroup

	// Metrics
	totalEngagements      atomic.Int64
	successfulEngagements atomic.Int64
	uasEliminated         atomic.Int64
	uasReachedTarget      atomic.Int64
}

// SimulationConfig holds the simulation configuration
type SimulationConfig struct {
	Duration             time.Duration
	UpdateInterval       time.Duration
	TickRate             time.Duration
	NumCounterUASSystems int
	NumUASThreats        int
	EngagementTypeMix    float64 // Percentage of kinetic systems
	CenterLocation       Location
	SpawnRadiusKm        float64
	TargetRadiusKm       float64
	WaveCount            int
	WaveDelay            time.Duration
	DefensePlacement     string // "ring", "cluster", "line"
	FormationType        string // "distributed", "concentrated", "waves"
}

// Location represents a geographic location
type Location struct {
	Lat float64
	Lon float64
	Alt float64
}

// CounterUASSystem represents a defensive Counter-UAS system
type CounterUASSystem struct {
	ID                    uuid.UUID
	Name                  string
	Status                string
	Position              *models.GeomPoint
	DetectionRadiusKm     float64
	EngagementRadiusKm    float64
	EngagementType        string
	AmmoCapacity          int
	AmmoRemaining         int
	SuccessRate           float64
	CooldownRemaining     int
	TotalEngagements      int
	SuccessfulEngagements int
	CurrentTarget         *uuid.UUID
	LastUpdateTime        time.Time
	mu                    sync.RWMutex
}

// UASThreat represents a drone threat in the swarm
type UASThreat struct {
	ID                uuid.UUID
	Name              string
	Status            string
	Position          *models.GeomPoint
	Velocity          *models.GeomPoint
	SpeedKph          float64
	AutonomyLevel     float64
	EvasionCapability bool
	FormationRole     string
	AttackVector      float64
	WaveNumber        int
	LastUpdateTime    time.Time
	mu                sync.RWMutex
}

// TerminationCondition represents a condition that ends the simulation
type TerminationCondition struct {
	Type        string // "all_threats_neutralized", "defensive_breach", "all_systems_depleted"
	Description string
	Met         bool
	Timestamp   time.Time
}

// NewSimulationController creates a new simulation controller
func NewSimulationController(client *client.Legion, organizationID string, config *SimulationConfig) *SimulationController {
	return &SimulationController{
		legionClient:      client,
		organizationID:    organizationID,
		config:            config,
		counterUASSystems: make(map[uuid.UUID]*CounterUASSystem),
		uasThreats:        make(map[uuid.UUID]*UASThreat),
		stopChan:          make(chan struct{}),
	}
}

// Initialize sets up the simulation
func (sc *SimulationController) Initialize(ctx context.Context) error {
	logger.Info("Initializing simulation controller...")
	sc.startTime = time.Now()

	// Initialize components
	sc.systemController = NewSystemController()
	sc.swarmController = NewSwarmController()
	sc.engagementCalc = core.NewEngagementCalculator()
	sc.updateBuffer = core.NewUpdateBuffer(sc.legionClient, sc.organizationID, 50, time.Second)

	// Initialize logger
	sc.simLogger = reporting.NewSimulationLogger("counter-uas-simulation")

	// Start update buffer
	sc.updateBuffer.Start(ctx)

	// Initialize system controller
	if err := sc.systemController.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize system controller: %w", err)
	}

	// Initialize swarm controller
	if err := sc.swarmController.Initialize(ctx, []string{"UAS"}); err != nil {
		return fmt.Errorf("failed to initialize swarm controller: %w", err)
	}

	return nil
}

// Start begins the simulation
func (sc *SimulationController) Start(ctx context.Context) error {
	logger.Info("Starting simulation...")

	// Create Counter-UAS systems
	if err := sc.createCounterUASSystems(ctx); err != nil {
		return fmt.Errorf("failed to create Counter-UAS systems: %w", err)
	}

	// Create UAS threats
	if err := sc.createUASThreats(ctx); err != nil {
		return fmt.Errorf("failed to create UAS threats: %w", err)
	}

	// Mark as running
	sc.isRunning.Store(true)

	// Start simulation loop in goroutine
	sc.wg.Add(1)
	go sc.runSimulationLoop(ctx)

	// Start Counter-UAS system goroutines
	for _, system := range sc.counterUASSystems {
		sc.wg.Add(1)
		go sc.runCounterUASSystem(ctx, system)
	}

	return nil
}

// Stop gracefully stops the simulation
func (sc *SimulationController) Stop() error {
	logger.Info("Stopping simulation...")

	// Mark as not running
	sc.isRunning.Store(false)

	// Signal stop
	close(sc.stopChan)

	// Wait for all goroutines to finish
	sc.wg.Wait()

	// Stop components
	sc.updateBuffer.Stop()

	// Force flush any remaining updates
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sc.updateBuffer.ForceFlush(ctx); err != nil {
		logger.Errorf("Error flushing final updates: %v", err)
	}

	sc.endTime = time.Now()

	// Generate AAR
	if err := sc.generateAAR(); err != nil {
		logger.Errorf("Error generating AAR: %v", err)
	}

	// Print summary
	sc.simLogger.PrintSummary()

	logger.Infof("Simulation stopped. Duration: %v", sc.endTime.Sub(sc.startTime))
	return nil
}

// GetStatus returns the current simulation status
func (sc *SimulationController) GetStatus() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Count active systems and threats
	activeSystems := 0
	depletedSystems := 0
	for _, system := range sc.counterUASSystems {
		if system.Status != CounterUASStatusDepleted {
			activeSystems++
		} else {
			depletedSystems++
		}
	}

	activeThreats := 0
	neutralizedThreats := 0
	for _, threat := range sc.uasThreats {
		if threat.Status == UASStatusEliminated || threat.Status == UASStatusJammed {
			neutralizedThreats++
		} else if threat.Status != UASStatusMissionComplete {
			activeThreats++
		}
	}

	status := map[string]interface{}{
		"start_time":             sc.startTime,
		"duration":               time.Since(sc.startTime).String(),
		"is_running":             sc.isRunning.Load(),
		"counter_uas_systems":    sc.config.NumCounterUASSystems,
		"active_systems":         activeSystems,
		"depleted_systems":       depletedSystems,
		"total_uas_threats":      sc.config.NumUASThreats,
		"active_threats":         activeThreats,
		"neutralized_threats":    neutralizedThreats,
		"total_engagements":      sc.totalEngagements.Load(),
		"successful_engagements": sc.successfulEngagements.Load(),
		"uas_eliminated":         sc.uasEliminated.Load(),
		"uas_reached_target":     sc.uasReachedTarget.Load(),
	}

	return status
}

// checkTerminationConditions checks if any termination conditions are met
func (sc *SimulationController) checkTerminationConditions() *TerminationCondition {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Check if all threats are neutralized
	allThreatsNeutralized := true
	for _, threat := range sc.uasThreats {
		if threat.Status != UASStatusEliminated && threat.Status != UASStatusJammed {
			allThreatsNeutralized = false
			break
		}
	}

	if allThreatsNeutralized {
		return &TerminationCondition{
			Type:        "all_threats_neutralized",
			Description: "All UAS threats have been eliminated or jammed",
			Met:         true,
			Timestamp:   time.Now(),
		}
	}

	// Check if any UAS reached target (defensive breach)
	if sc.uasReachedTarget.Load() > 0 {
		return &TerminationCondition{
			Type:        "defensive_breach",
			Description: fmt.Sprintf("%d UAS reached their target", sc.uasReachedTarget.Load()),
			Met:         true,
			Timestamp:   time.Now(),
		}
	}

	// Check if all systems are depleted with active threats
	allSystemsDepleted := true
	hasActiveThreats := false

	for _, system := range sc.counterUASSystems {
		if system.Status != CounterUASStatusDepleted {
			allSystemsDepleted = false
			break
		}
	}

	for _, threat := range sc.uasThreats {
		if threat.Status != UASStatusEliminated && threat.Status != UASStatusJammed && threat.Status != UASStatusMissionComplete {
			hasActiveThreats = true
			break
		}
	}

	if allSystemsDepleted && hasActiveThreats {
		return &TerminationCondition{
			Type:        "all_systems_depleted",
			Description: "All Counter-UAS systems are depleted with active threats remaining",
			Met:         true,
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// generateAAR generates the After Action Report
func (sc *SimulationController) generateAAR() error {
	logger.Info("Generating After Action Report...")

	aarGen := reporting.NewAARGenerator(sc.simLogger, reporting.AARConfig{
		OutputDir:   "./reports/",
		Format:      "json",
		DetailLevel: "detailed",
	})

	aar, err := aarGen.GenerateAAR()
	if err != nil {
		return fmt.Errorf("failed to generate AAR: %w", err)
	}

	return aarGen.SaveAAR(aar)
}

// createCounterUASSystems creates Counter-UAS system entities in Legion
func (sc *SimulationController) createCounterUASSystems(ctx context.Context) error {
	logger.Info("Creating Counter-UAS systems...")

	// Convert center location to ECEF
	centerX, centerY, centerZ := latLonAltToECEF(sc.config.CenterLocation.Lat, sc.config.CenterLocation.Lon, sc.config.CenterLocation.Alt)

	// Determine number of kinetic vs EW systems
	numKinetic := int(float64(sc.config.NumCounterUASSystems) * sc.config.EngagementTypeMix)

	// Create systems based on placement pattern
	for i := 0; i < sc.config.NumCounterUASSystems; i++ {
		// Determine engagement type
		engagementType := "electronic_warfare"
		if i < numKinetic {
			engagementType = "kinetic"
		}

		// Calculate position based on placement pattern
		var position *models.GeomPoint
		switch sc.config.DefensePlacement {
		case "ring":
			angle := float64(i) * (360.0 / float64(sc.config.NumCounterUASSystems)) * math.Pi / 180.0
			offsetX := sc.config.TargetRadiusKm * 1000 * math.Cos(angle)
			offsetY := sc.config.TargetRadiusKm * 1000 * math.Sin(angle)
			position = &models.GeomPoint{
				Type:        "Point",
				Coordinates: []float64{centerX + offsetX, centerY + offsetY, centerZ},
			}
		case "cluster":
			// Random placement within target radius
			angle := rand.Float64() * 2 * math.Pi
			radius := rand.Float64() * sc.config.TargetRadiusKm * 1000
			position = &models.GeomPoint{
				Type:        "Point",
				Coordinates: []float64{centerX + radius*math.Cos(angle), centerY + radius*math.Sin(angle), centerZ},
			}
		case "line":
			// Linear placement
			spacing := (sc.config.TargetRadiusKm * 2000) / float64(sc.config.NumCounterUASSystems-1)
			offset := -sc.config.TargetRadiusKm*1000 + float64(i)*spacing
			position = &models.GeomPoint{
				Type:        "Point",
				Coordinates: []float64{centerX + offset, centerY, centerZ},
			}
		default:
			position = &models.GeomPoint{
				Type:        "Point",
				Coordinates: []float64{centerX, centerY, centerZ},
			}
		}

		// Create system instance
		name := fmt.Sprintf("CUAS-%s-%d", engagementType, i+1)
		system := NewCounterUASSystem(name, position, engagementType)

		// Create entity in Legion
		metadata, err := json.Marshal(system.GetMetadata())
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataRaw := json.RawMessage(metadata)

		entityReq := &models.CreateEntityRequest{
			OrganizationID: uuid.MustParse(sc.organizationID),
			Name:           system.Name,
			Category:       models.CATEGORY_DEVICE,
			Type:           EntityTypeCounterUAS,
			Status:         system.Status,
			Metadata:       &metadataRaw,
		}

		orgCtx := client.WithOrgID(ctx, sc.organizationID)
		createdEntity, err := sc.legionClient.CreateEntity(orgCtx, entityReq)
		if err != nil {
			return fmt.Errorf("failed to create Counter-UAS entity: %w", err)
		}

		system.ID = createdEntity.ID
		sc.counterUASSystems[system.ID] = system

		// Set initial location
		locationReq := &models.CreateEntityLocationRequest{
			Position: position,
		}

		if _, err := sc.legionClient.CreateEntityLocation(orgCtx, system.ID.String(), locationReq); err != nil {
			return fmt.Errorf("failed to set Counter-UAS location: %w", err)
		}

		logger.Infof("Created Counter-UAS system: %s (ID: %s)", system.Name, system.ID)
	}

	return nil
}

// createUASThreats creates UAS threat entities in Legion
func (sc *SimulationController) createUASThreats(ctx context.Context) error {
	logger.Info("Creating UAS threats...")

	// Convert center location to ECEF
	centerX, centerY, centerZ := latLonAltToECEF(sc.config.CenterLocation.Lat, sc.config.CenterLocation.Lon, sc.config.CenterLocation.Alt)

	// Create threats in waves
	threatsPerWave := sc.config.NumUASThreats / sc.config.WaveCount
	remainingThreats := sc.config.NumUASThreats % sc.config.WaveCount

	threatIndex := 0
	for wave := 0; wave < sc.config.WaveCount; wave++ {
		threatsInThisWave := threatsPerWave
		if wave < remainingThreats {
			threatsInThisWave++
		}

		for i := 0; i < threatsInThisWave; i++ {
			// Determine formation role
			formationRole := "follower"
			if i == 0 {
				formationRole = "leader"
			} else if i < 3 {
				formationRole = "scout"
			}

			// Random spawn position outside spawn radius
			angle := rand.Float64() * 2 * math.Pi
			spawnDistance := sc.config.SpawnRadiusKm * 1000
			position := &models.GeomPoint{
				Type: "Point",
				Coordinates: []float64{
					centerX + spawnDistance*math.Cos(angle),
					centerY + spawnDistance*math.Sin(angle),
					centerZ + 100 + rand.Float64()*400, // 100-500m altitude
				},
			}

			// Create threat instance
			name := fmt.Sprintf("UAS-W%d-%d", wave+1, i+1)
			threat := NewUASThreat(name, position, wave, formationRole)

			// Create entity in Legion
			metadata, err := json.Marshal(threat.GetMetadata())
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}
			metadataRaw := json.RawMessage(metadata)

			entityReq := &models.CreateEntityRequest{
				OrganizationID: uuid.MustParse(sc.organizationID),
				Name:           threat.Name,
				Category:       models.CATEGORY_UXV,
				Type:           EntityTypeUAS,
				Status:         threat.Status,
				Metadata:       &metadataRaw,
			}

			orgCtx := client.WithOrgID(ctx, sc.organizationID)
			createdEntity, err := sc.legionClient.CreateEntity(orgCtx, entityReq)
			if err != nil {
				return fmt.Errorf("failed to create UAS entity: %w", err)
			}

			threat.ID = createdEntity.ID
			sc.uasThreats[threat.ID] = threat

			// Set initial location
			locationReq := &models.CreateEntityLocationRequest{
				Position: position,
			}

			if _, err := sc.legionClient.CreateEntityLocation(orgCtx, threat.ID.String(), locationReq); err != nil {
				return fmt.Errorf("failed to set UAS location: %w", err)
			}

			// Add to swarm controller
			droneState := &DroneState{
				ID:         threat.ID,
				Type:       EntityTypeUAS,
				Position:   Vector3D{X: position.Coordinates[0], Y: position.Coordinates[1], Z: position.Coordinates[2]},
				Velocity:   Vector3D{X: threat.Velocity.Coordinates[0], Y: threat.Velocity.Coordinates[1], Z: threat.Velocity.Coordinates[2]},
				Health:     100.0,
				Ammunition: 0,
				FuelLevel:  100.0,
				Role:       formationRole,
				TeamName:   "UAS",
			}

			if err := sc.swarmController.AddDrone("UAS", droneState); err != nil {
				return fmt.Errorf("failed to add drone to swarm: %w", err)
			}

			logger.Infof("Created UAS threat: %s (ID: %s)", threat.Name, threat.ID)
			threatIndex++
		}
	}

	return nil
}

// runSimulationLoop runs the main simulation update loop
func (sc *SimulationController) runSimulationLoop(ctx context.Context) {
	defer sc.wg.Done()

	ticker := time.NewTicker(sc.config.TickRate)
	defer ticker.Stop()

	updateTicker := time.NewTicker(sc.config.UpdateInterval)
	defer updateTicker.Stop()

	waveLaunched := make([]bool, sc.config.WaveCount)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sc.stopChan:
			return
		case <-ticker.C:
			// Update UAS positions and behaviors
			if err := sc.updateUASMovement(ctx); err != nil {
				logger.Errorf("Error updating UAS movement: %v", err)
			}

			// Check for wave launches
			elapsed := time.Since(sc.startTime)
			for wave := 0; wave < sc.config.WaveCount; wave++ {
				if !waveLaunched[wave] && elapsed >= time.Duration(wave)*sc.config.WaveDelay {
					sc.launchWave(wave)
					waveLaunched[wave] = true
				}
			}

			// Check termination conditions
			if condition := sc.checkTerminationConditions(); condition != nil {
				logger.Infof("Termination condition met: %s", condition.Description)
				// Log termination as an objective event
				sc.simLogger.LogObjective("Simulation", "termination", "complete", map[string]interface{}{
					"condition_type": condition.Type,
					"description":    condition.Description,
				})
				go func() {
					_ = sc.Stop()
				}()
				return
			}

		case <-updateTicker.C:
			// Flush updates to Legion
			if err := sc.updateBuffer.Flush(ctx); err != nil {
				logger.Errorf("Error flushing updates: %v", err)
			}

			// Log status
			status := sc.GetStatus()
			logger.Infof("Simulation status: %v active threats, %v neutralized, %v engagements",
				status["active_threats"], status["neutralized_threats"], status["total_engagements"])
		}
	}
}

// runCounterUASSystem runs a single Counter-UAS system's behavior
func (sc *SimulationController) runCounterUASSystem(ctx context.Context, system *CounterUASSystem) {
	defer sc.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond) // 10Hz update rate
	defer ticker.Stop()

	cooldownTicker := time.NewTicker(time.Second)
	defer cooldownTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sc.stopChan:
			return
		case <-ticker.C:
			// Update system behavior based on status
			sc.updateCounterUASBehavior(ctx, system)

		case <-cooldownTicker.C:
			// Decrement cooldown
			system.mu.Lock()
			if system.CooldownRemaining > 0 {
				system.CooldownRemaining--
				if system.CooldownRemaining == 0 && system.Status == CounterUASStatusCooldown {
					system.Status = CounterUASStatusIdle
					sc.updateBuffer.QueueStatusUpdate(system.ID, system.Status)
				}
			}
			system.mu.Unlock()
		}
	}
}

// updateCounterUASBehavior updates a Counter-UAS system's behavior
func (sc *SimulationController) updateCounterUASBehavior(ctx context.Context, system *CounterUASSystem) {
	system.mu.Lock()
	currentStatus := system.Status
	system.mu.Unlock()

	switch currentStatus {
	case CounterUASStatusIdle:
		// Scan for threats
		if threats := sc.detectThreats(system); len(threats) > 0 {
			// Select highest priority target
			target := sc.selectTarget(system, threats)
			if target != nil {
				system.mu.Lock()
				system.CurrentTarget = &target.ID
				system.Status = CounterUASStatusTracking
				system.mu.Unlock()

				sc.updateBuffer.QueueStatusUpdate(system.ID, system.Status)
				// Log detection event
				sc.simLogger.LogDetection(system.ID, target.ID, "Counter-UAS", "UAS",
					calculateDistanceKm(system.Position, target.Position)*1000)
			}
		}

	case CounterUASStatusTracking:
		// Check if target is still valid and in range
		system.mu.RLock()
		targetID := system.CurrentTarget
		system.mu.RUnlock()

		if targetID != nil {
			if threat, exists := sc.uasThreats[*targetID]; exists && sc.canEngage(system, threat) {
				// Transition to engaging
				system.mu.Lock()
				system.Status = CounterUASStatusEngaging
				system.mu.Unlock()

				sc.updateBuffer.QueueStatusUpdate(system.ID, system.Status)

				// Process engagement
				sc.processEngagement(ctx, system, threat)
			} else {
				// Lost target or out of range
				system.mu.Lock()
				system.CurrentTarget = nil
				system.Status = CounterUASStatusIdle
				system.mu.Unlock()

				sc.updateBuffer.QueueStatusUpdate(system.ID, system.Status)
			}
		}

	case CounterUASStatusEngaging:
		// Engagement is processed, transition to cooldown
		system.mu.Lock()
		if system.EngagementType == "kinetic" {
			system.CooldownRemaining = 5 + rand.Intn(3) // 5-7 seconds
		} else {
			system.CooldownRemaining = 8 + rand.Intn(3) // 8-10 seconds
		}
		system.Status = CounterUASStatusCooldown
		system.CurrentTarget = nil
		system.mu.Unlock()

		sc.updateBuffer.QueueStatusUpdate(system.ID, system.Status)

	case CounterUASStatusDepleted:
		// System is out of ammo or otherwise depleted
		// No further actions

	case CounterUASStatusCooldown:
		// Waiting for cooldown to complete
		// Handled by cooldown ticker
	}
}

// Helper functions

// NewCounterUASSystem creates a new Counter-UAS system
func NewCounterUASSystem(name string, position *models.GeomPoint, engagementType string) *CounterUASSystem {
	successRate := 0.7 + rand.Float64()*0.2 // 0.7-0.9 for kinetic
	if engagementType == "electronic_warfare" {
		successRate = 0.5 + rand.Float64()*0.2 // 0.5-0.7 for EW
	}

	ammoCapacity := -1 // Unlimited for EW
	if engagementType == "kinetic" {
		ammoCapacity = 5
	}

	return &CounterUASSystem{
		ID:                    uuid.New(),
		Name:                  name,
		Status:                CounterUASStatusIdle,
		Position:              position,
		DetectionRadiusKm:     10.0,
		EngagementRadiusKm:    5.0,
		EngagementType:        engagementType,
		AmmoCapacity:          ammoCapacity,
		AmmoRemaining:         ammoCapacity,
		SuccessRate:           successRate,
		CooldownRemaining:     0,
		TotalEngagements:      0,
		SuccessfulEngagements: 0,
		CurrentTarget:         nil,
		LastUpdateTime:        time.Now(),
	}
}

// NewUASThreat creates a new UAS threat
func NewUASThreat(name string, position *models.GeomPoint, waveNumber int, formationRole string) *UASThreat {
	speedKph := 50.0 + rand.Float64()*150.0   // 50-200 kph
	autonomyLevel := rand.Float64()           // 0.0-1.0
	evasionCapability := rand.Float64() > 0.3 // 70% have evasion
	attackVector := rand.Float64() * 360.0    // 0-360 degrees

	// Calculate initial velocity based on attack vector
	velocityMagnitude := speedKph / 3.6 // Convert to m/s
	velocityAngleRad := attackVector * math.Pi / 180.0

	velocity := &models.GeomPoint{
		Type: "Point",
		Coordinates: []float64{
			velocityMagnitude * math.Cos(velocityAngleRad),
			velocityMagnitude * math.Sin(velocityAngleRad),
			0,
		},
	}

	return &UASThreat{
		ID:                uuid.New(),
		Name:              name,
		Status:            UASStatusForming,
		Position:          position,
		Velocity:          velocity,
		SpeedKph:          speedKph,
		AutonomyLevel:     autonomyLevel,
		EvasionCapability: evasionCapability,
		FormationRole:     formationRole,
		AttackVector:      attackVector,
		WaveNumber:        waveNumber,
		LastUpdateTime:    time.Now(),
	}
}

// GetMetadata returns the metadata map for a Counter-UAS system
func (c *CounterUASSystem) GetMetadata() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metadata := map[string]interface{}{
		"status":                 c.Status,
		"detection_radius_km":    c.DetectionRadiusKm,
		"engagement_radius_km":   c.EngagementRadiusKm,
		"engagement_type":        c.EngagementType,
		"success_rate":           c.SuccessRate,
		"cooldown_remaining":     c.CooldownRemaining,
		"total_engagements":      c.TotalEngagements,
		"successful_engagements": c.SuccessfulEngagements,
	}

	if c.EngagementType == "kinetic" {
		metadata["ammo_capacity"] = c.AmmoCapacity
		metadata["ammo_remaining"] = c.AmmoRemaining
	}

	if c.CurrentTarget != nil {
		metadata["current_target"] = c.CurrentTarget.String()
	}

	return metadata
}

// GetMetadata returns the metadata map for a UAS threat
func (u *UASThreat) GetMetadata() map[string]interface{} {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return map[string]interface{}{
		"status":             u.Status,
		"speed_kph":          u.SpeedKph,
		"autonomy_level":     u.AutonomyLevel,
		"evasion_capability": u.EvasionCapability,
		"formation_role":     u.FormationRole,
		"attack_vector":      u.AttackVector,
		"wave_number":        u.WaveNumber,
	}
}

// Coordinate conversion helpers
func latLonAltToECEF(lat, lon, alt float64) (x, y, z float64) {
	// WGS84 ellipsoid constants
	a := 6378137.0           // Semi-major axis
	f := 1.0 / 298.257223563 // Flattening
	e2 := 2*f - f*f          // First eccentricity squared

	// Convert degrees to radians
	latRad := lat * math.Pi / 180.0
	lonRad := lon * math.Pi / 180.0

	// Calculate N - radius of curvature
	N := a / math.Sqrt(1-e2*math.Sin(latRad)*math.Sin(latRad))

	// Calculate ECEF coordinates
	x = (N + alt) * math.Cos(latRad) * math.Cos(lonRad)
	y = (N + alt) * math.Cos(latRad) * math.Sin(lonRad)
	z = (N*(1-e2) + alt) * math.Sin(latRad)

	return x, y, z
}

// calculateDistanceKm calculates the distance in kilometers between two ECEF points
func calculateDistanceKm(p1, p2 *models.GeomPoint) float64 {
	dx := p2.Coordinates[0] - p1.Coordinates[0]
	dy := p2.Coordinates[1] - p1.Coordinates[1]
	dz := p2.Coordinates[2] - p1.Coordinates[2]
	return math.Sqrt(dx*dx+dy*dy+dz*dz) / 1000.0
}

// detectThreats detects UAS threats within range of a Counter-UAS system
func (sc *SimulationController) detectThreats(system *CounterUASSystem) []*UASThreat {
	var threats []*UASThreat

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	for _, threat := range sc.uasThreats {
		// Skip eliminated or jammed threats
		if threat.Status == UASStatusEliminated || threat.Status == UASStatusJammed {
			continue
		}

		// Calculate distance
		distance := calculateDistanceKm(system.Position, threat.Position)

		// Check if within detection range
		if distance <= system.DetectionRadiusKm {
			threats = append(threats, threat)
		}
	}

	return threats
}

// selectTarget selects the highest priority target from detected threats
func (sc *SimulationController) selectTarget(system *CounterUASSystem, threats []*UASThreat) *UASThreat {
	if len(threats) == 0 {
		return nil
	}

	var bestTarget *UASThreat
	var bestPriority float64

	for _, threat := range threats {
		// Calculate priority based on:
		// - Distance (closer = higher priority)
		// - Speed (faster = higher priority)
		// - Role (leader/scout = higher priority)

		distance := calculateDistanceKm(system.Position, threat.Position)
		distancePriority := 1.0 - (distance / system.DetectionRadiusKm)

		speedPriority := threat.SpeedKph / 200.0 // Normalize to 0-1

		rolePriority := 1.0
		switch threat.FormationRole {
		case "leader":
			rolePriority = 1.5
		case "scout":
			rolePriority = 1.2
		}

		// Calculate total priority
		priority := distancePriority*0.5 + speedPriority*0.3 + rolePriority*0.2

		if bestTarget == nil || priority > bestPriority {
			bestTarget = threat
			bestPriority = priority
		}
	}

	return bestTarget
}

// canEngage checks if a Counter-UAS system can engage a threat
func (sc *SimulationController) canEngage(system *CounterUASSystem, threat *UASThreat) bool {
	// Calculate distance
	distance := calculateDistanceKm(system.Position, threat.Position)

	// Check range
	if distance > system.EngagementRadiusKm {
		return false
	}

	// Check if system is ready
	if system.CooldownRemaining > 0 {
		return false
	}

	// Check ammo for kinetic systems
	if system.EngagementType == "kinetic" && system.AmmoRemaining <= 0 {
		return false
	}

	// Check if EW can affect this target
	if system.EngagementType == "electronic_warfare" && threat.AutonomyLevel >= 0.5 {
		return false
	}

	return true
}

// processEngagement processes an engagement between a Counter-UAS system and a threat
func (sc *SimulationController) processEngagement(_ context.Context, system *CounterUASSystem, threat *UASThreat) {
	distance := calculateDistanceKm(system.Position, threat.Position)

	// Create engagement info
	attackerInfo := core.CounterUASInfo{
		ID:                system.ID,
		EngagementType:    system.EngagementType,
		EngagementRangeKm: system.EngagementRadiusKm,
		SuccessRate:       system.SuccessRate,
		AmmoRemaining:     system.AmmoRemaining,
		CooldownRemaining: system.CooldownRemaining,
	}

	targetInfo := core.UASInfo{
		ID:                threat.ID,
		AutonomyLevel:     threat.AutonomyLevel,
		SpeedKph:          threat.SpeedKph,
		EvasionCapability: threat.EvasionCapability,
		Status:            threat.Status,
	}

	// Calculate environmental modifiers
	modifiers := core.Modifiers{
		Visibility:    1.0, // Clear visibility
		Weather:       1.0, // Clear weather
		Terrain:       1.0, // Open terrain
		TargetSpeed:   threat.SpeedKph,
		TargetEvading: threat.Status == UASStatusEvading,
	}

	// Calculate engagement outcome
	result := sc.engagementCalc.CalculateEngagement(attackerInfo, targetInfo, distance, modifiers)

	// Update metrics
	sc.totalEngagements.Add(1)
	system.TotalEngagements++

	if result.Success {
		sc.successfulEngagements.Add(1)
		system.SuccessfulEngagements++

		// Update threat status
		if system.EngagementType == "kinetic" {
			threat.Status = UASStatusEliminated
			sc.uasEliminated.Add(1)
		} else {
			threat.Status = UASStatusJammed
		}

		// Update buffers
		sc.updateBuffer.QueueStatusUpdate(threat.ID, threat.Status)
		sc.updateBuffer.QueueMetadataUpdate(threat.ID, "eliminated_by", system.Name)

		// Log engagement
		details := map[string]interface{}{
			"distance_km": distance,
			"autonomy":    threat.AutonomyLevel,
			"hit":         result.Success,
			"type":        system.EngagementType,
		}
		sc.simLogger.LogEngagement(system.ID, threat.ID, fmt.Sprintf("%s engagement", system.EngagementType), details)
	}

	// Consume ammo for kinetic systems
	if system.EngagementType == "kinetic" {
		system.AmmoRemaining--
		if system.AmmoRemaining <= 0 {
			system.Status = CounterUASStatusDepleted
			sc.updateBuffer.QueueStatusUpdate(system.ID, system.Status)
		}
	}
}

// updateUASMovement updates UAS positions and behaviors
func (sc *SimulationController) updateUASMovement(ctx context.Context) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Get center location in ECEF
	centerX, centerY, centerZ := latLonAltToECEF(sc.config.CenterLocation.Lat, sc.config.CenterLocation.Lon, sc.config.CenterLocation.Alt)
	center := &models.GeomPoint{
		Type:        "Point",
		Coordinates: []float64{centerX, centerY, centerZ},
	}

	deltaTime := sc.config.TickRate.Seconds()

	for _, threat := range sc.uasThreats {
		// Skip eliminated or mission complete threats
		if threat.Status == UASStatusEliminated || threat.Status == UASStatusMissionComplete {
			continue
		}

		// Update position based on velocity
		if threat.Status != UASStatusJammed {
			threat.Position.Coordinates[0] += threat.Velocity.Coordinates[0] * deltaTime
			threat.Position.Coordinates[1] += threat.Velocity.Coordinates[1] * deltaTime
			threat.Position.Coordinates[2] += threat.Velocity.Coordinates[2] * deltaTime

			// Queue position update
			sc.updateBuffer.QueuePositionUpdate(threat.ID, threat.Position)
		}

		// Check if reached target
		distance := calculateDistanceKm(threat.Position, center)
		if distance <= sc.config.TargetRadiusKm {
			threat.Status = UASStatusMissionComplete
			sc.uasReachedTarget.Add(1)
			sc.updateBuffer.QueueStatusUpdate(threat.ID, threat.Status)

			// Log mission complete as objective event
			sc.simLogger.LogObjective("UAS", "reached_target", "complete", map[string]interface{}{
				"threat_id":   threat.ID.String(),
				"threat_name": threat.Name,
			})
		}

		// Update status based on detection
		if threat.Status == UASStatusInbound {
			// Check if detected by any Counter-UAS system
			for _, system := range sc.counterUASSystems {
				if calculateDistanceKm(system.Position, threat.Position) <= system.DetectionRadiusKm {
					threat.Status = UASStatusDetected
					sc.updateBuffer.QueueStatusUpdate(threat.ID, threat.Status)
					break
				}
			}
		}
	}

	// Update swarm behavior
	return sc.swarmController.Update(ctx, deltaTime)
}

// launchWave launches a wave of UAS threats
func (sc *SimulationController) launchWave(waveNumber int) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	logger.Infof("Launching wave %d", waveNumber+1)

	// Transition threats in this wave from FORMING to INBOUND
	for _, threat := range sc.uasThreats {
		if threat.WaveNumber == waveNumber && threat.Status == UASStatusForming {
			threat.Status = UASStatusInbound
			sc.updateBuffer.QueueStatusUpdate(threat.ID, threat.Status)
		}
	}

	// Count threats in this wave
	threatCount := 0
	for _, threat := range sc.uasThreats {
		if threat.WaveNumber == waveNumber {
			threatCount++
		}
	}

	sc.simLogger.LogWaveLaunch("UAS", waveNumber+1, threatCount, map[string]interface{}{
		"status": "inbound",
	})
}
