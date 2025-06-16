package simulation

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/cmd/drone-swarm/controllers"
	"github.com/picogrid/legion-simulations/cmd/drone-swarm/core"
	"github.com/picogrid/legion-simulations/cmd/drone-swarm/reporting"
	"github.com/picogrid/legion-simulations/pkg/client"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
	"github.com/picogrid/legion-simulations/pkg/simulation"
)

// Track number counter for generating military-style track numbers
var trackNumberCounter uint32 = 0

// generateTrackNumber creates a military-style track number
func generateTrackNumber() string {
	num := atomic.AddUint32(&trackNumberCounter, 1)
	return fmt.Sprintf("TK-%04d", num)
}

// generateUniqueTrackNumber creates a track number with timestamp for uniqueness
func generateUniqueTrackNumber() string {
	num := atomic.AddUint32(&trackNumberCounter, 1)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("TK-%04d-%d", num, timestamp)
}

// DroneSwarmSimulation implements a multi-team drone swarm combat simulation
type DroneSwarmSimulation struct {
	// Configuration
	config SimulationConfig

	// Controllers
	simController    *controllers.SimulationController
	systemController *controllers.SystemController
	swarmController  *controllers.SwarmController

	// Core systems
	engagementCalculator *core.EngagementCalculator
	swarmBehavior        *core.SwarmBehaviorEngine
	updateBuffer         *core.UpdateBuffer

	// Reporting
	simLogger    *reporting.SimulationLogger
	aarGenerator *reporting.AARGenerator

	// Entity tracking
	counterUASSystems map[uuid.UUID]*CounterUASSystem
	uasThreats        map[uuid.UUID]*UASThreat

	// Legion client
	legionClient *client.Legion

	// Synchronization
	mu       sync.RWMutex
	stopChan chan struct{}

	// Statistics
	stats SimulationStats
}

// SimulationConfig holds configuration parameters
type SimulationConfig struct {
	OrganizationID       string
	NumCounterUASSystems int
	NumUASThreats        int
	NumWaves             int
	SimDuration          time.Duration
	UpdateInterval       time.Duration
	BaseLocation         Location
	SimulationRadius     float64 // km
	EnableDebugLogging   bool
	CleanupExisting      bool
	UseUniqueNames       bool // Add timestamp to entity names for uniqueness
}

// SimulationStats tracks simulation statistics
type SimulationStats struct {
	TotalEngagements      int
	SuccessfulEngagements int
	UASEliminated         int
	UASPenetrated         int
	CounterUASLosses      int
	SimulationOutcome     string
	mu                    sync.RWMutex
}

// NewDroneSwarmSimulation creates a new instance of the drone swarm simulation
func NewDroneSwarmSimulation() simulation.Simulation {
	return &DroneSwarmSimulation{
		counterUASSystems: make(map[uuid.UUID]*CounterUASSystem),
		uasThreats:        make(map[uuid.UUID]*UASThreat),
		stopChan:          make(chan struct{}),
	}
}

// Name returns the simulation name
func (s *DroneSwarmSimulation) Name() string {
	return "Drone Swarm Combat"
}

// Description returns the simulation description
func (s *DroneSwarmSimulation) Description() string {
	return "Multi-team drone swarm simulation with engagement, targeting, and complex behaviors"
}

// Configure sets up the simulation with provided parameters
func (s *DroneSwarmSimulation) Configure(params map[string]interface{}) error {
	logger.Info("Configuring drone swarm simulation...")

	// Set defaults
	s.config = SimulationConfig{
		NumCounterUASSystems: 10,
		NumUASThreats:        50,
		NumWaves:             5,
		SimDuration:          2 * time.Minute,
		UpdateInterval:       1 * time.Second,
		BaseLocation:         Location{Lat: 37.7749, Lon: -122.4194, Alt: 100},
		SimulationRadius:     15.0, // km
		EnableDebugLogging:   true,
		CleanupExisting:      true,
	}

	// Parse configuration parameters
	if val, ok := params["organization_id"].(string); ok {
		s.config.OrganizationID = val
	}

	if val, ok := params["num_counter_uas_systems"].(float64); ok {
		s.config.NumCounterUASSystems = int(val)
	}

	if val, ok := params["num_uas_threats"].(float64); ok {
		s.config.NumUASThreats = int(val)
	}

	if val, ok := params["waves"].(float64); ok {
		s.config.NumWaves = int(val)
	}

	if val, ok := params["duration"].(time.Duration); ok {
		s.config.SimDuration = val
	}

	if val, ok := params["update_interval"].(time.Duration); ok {
		s.config.UpdateInterval = val
	}

	if val, ok := params["debug_logging"].(bool); ok {
		s.config.EnableDebugLogging = val
	}

	if val, ok := params["cleanup_existing"].(bool); ok {
		s.config.CleanupExisting = val
	}

	// Validate configuration
	if s.config.NumCounterUASSystems < 1 {
		return fmt.Errorf("must have at least 1 Counter-UAS system")
	}

	if s.config.NumUASThreats < 1 {
		return fmt.Errorf("must have at least 1 UAS threat")
	}

	logger.Infof("Configuration: %d Counter-UAS systems vs %d UAS threats in %d waves",
		s.config.NumCounterUASSystems, s.config.NumUASThreats, s.config.NumWaves)

	return nil
}

// Run executes the simulation
func (s *DroneSwarmSimulation) Run(ctx context.Context, legionClient *client.Legion) error {
	logger.Infof("Starting %s simulation", s.Name())
	s.legionClient = legionClient

	// Initialize controllers and systems
	if err := s.initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize simulation: %w", err)
	}

	// Clean up existing entities if requested
	if s.config.CleanupExisting {
		if err := s.cleanupExistingEntities(ctx); err != nil {
			logger.Warnf("Failed to cleanup existing entities: %v", err)
			// Enable unique names as fallback
			s.config.UseUniqueNames = true
		}
	}

	// Create entities
	if err := s.createEntities(ctx); err != nil {
		// If we get a conflict error, retry with unique names
		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "already exists") {
			logger.Warn("Entity name conflict detected, retrying with unique names...")
			s.config.UseUniqueNames = true
			// Clear any partially created entities
			s.counterUASSystems = make(map[uuid.UUID]*CounterUASSystem)
			s.uasThreats = make(map[uuid.UUID]*UASThreat)
			// Retry with unique names
			if err := s.createEntities(ctx); err != nil {
				return fmt.Errorf("failed to create entities with unique names: %w", err)
			}
		} else {
			return fmt.Errorf("failed to create entities: %w", err)
		}
	}

	// Deploy entities to initial positions
	if err := s.deployEntities(ctx); err != nil {
		return fmt.Errorf("failed to deploy entities: %w", err)
	}

	// Start simulation loop
	return s.runSimulationLoop(ctx)
}

// initialize sets up controllers and systems
func (s *DroneSwarmSimulation) initialize(ctx context.Context) error {
	logger.Info("Initializing simulation controllers and systems...")

	// Initialize simulation logger
	s.simLogger = reporting.NewSimulationLogger("counter-uas-simulation")

	// Initialize AAR generator
	aarConfig := reporting.AARConfig{
		OutputDir:     "./reports",
		Format:        "json",
		IncludeGraphs: true,
		DetailLevel:   "detailed",
	}
	s.aarGenerator = reporting.NewAARGenerator(s.simLogger, aarConfig)

	// Initialize core systems
	s.engagementCalculator = core.NewEngagementCalculator()
	s.swarmBehavior = core.NewSwarmBehaviorEngine()
	s.updateBuffer = core.NewUpdateBuffer(s.legionClient, s.config.OrganizationID, 100, 1*time.Second)

	// Initialize controllers
	simConfig := &controllers.SimulationConfig{
		Duration:       s.config.SimDuration,
		UpdateInterval: s.config.UpdateInterval,
		TickRate:       100 * time.Millisecond,
	}
	s.simController = controllers.NewSimulationController(s.legionClient, s.config.OrganizationID, simConfig)
	s.systemController = controllers.NewSystemController()
	s.swarmController = controllers.NewSwarmController()

	// Initialize controllers
	if err := s.simController.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize simulation controller: %w", err)
	}

	if err := s.systemController.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize system controller: %w", err)
	}

	teams := []string{"Counter-UAS", "UAS-Threats"}
	if err := s.swarmController.Initialize(ctx, teams); err != nil {
		return fmt.Errorf("failed to initialize swarm controller: %w", err)
	}

	return nil
}

// createEntities creates all entities in Legion
func (s *DroneSwarmSimulation) createEntities(ctx context.Context) error {
	logger.Info("Creating entities in Legion...")

	// Create Counter-UAS systems (BLUE FORCE)
	for i := 0; i < s.config.NumCounterUASSystems; i++ {
		// Alternate between kinetic and EW systems
		engagementType := EngagementTypeKinetic
		if i%2 == 1 {
			engagementType = EngagementTypeEW
		}

		name := fmt.Sprintf("Counter-UAS-%02d", i+1)
		if s.config.UseUniqueNames {
			name = fmt.Sprintf("Counter-UAS-%02d-%d", i+1, time.Now().Unix())
		}
		position := &models.GeomPoint{
			Type:        "Point",
			Coordinates: []float64{0, 0, 0}, // Will be set during deployment
		}

		system := NewCounterUASSystem(name, position, engagementType)
		s.counterUASSystems[system.ID] = system

		// Prepare metadata with full BLUE FORCE visibility
		metadata, err := json.Marshal(system.GetMetadata())
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataRaw := json.RawMessage(metadata)

		// Create entity in Legion
		entityReq := &models.CreateEntityRequest{
			OrganizationID: uuid.MustParse(s.config.OrganizationID),
			Name:           name,
			Category:       models.CATEGORY_DEVICE,
			Type:           EntityTypeCounterUAS,
			Status:         system.Status,
			Metadata:       &metadataRaw,
		}

		// Create context with organization ID
		orgCtx := client.WithOrgID(ctx, s.config.OrganizationID)
		createdEntity, err := s.legionClient.CreateEntity(orgCtx, entityReq)
		if err != nil {
			return fmt.Errorf("failed to create Counter-UAS entity %s: %w", name, err)
		}

		// Update the map with the new Legion ID
		delete(s.counterUASSystems, system.ID) // Remove old entry
		system.ID = createdEntity.ID
		s.counterUASSystems[system.ID] = system // Add with new ID

		logger.Infof("🛡️ Deployed %s (%s) - %s system online", system.Name, system.Callsign, engagementType)
	}

	// Create UAS threats in waves (RED FORCE)
	threatsPerWave := s.config.NumUASThreats / s.config.NumWaves
	for wave := 0; wave < s.config.NumWaves; wave++ {
		for i := 0; i < threatsPerWave; i++ {
			var trackNumber string
			if s.config.UseUniqueNames {
				trackNumber = generateUniqueTrackNumber()
			} else {
				trackNumber = generateTrackNumber()
			}
			position := &models.GeomPoint{
				Type:        "Point",
				Coordinates: []float64{0, 0, 0}, // Will be set during deployment
			}

			threat := NewUASThreat(trackNumber, position, wave+1)
			s.uasThreats[threat.ID] = threat

			// Prepare metadata with only observable RED FORCE data
			metadata, err := json.Marshal(threat.GetMetadata())
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}
			metadataRaw := json.RawMessage(metadata)

			// Create entity in Legion - using track classification
			entityReq := &models.CreateEntityRequest{
				OrganizationID: uuid.MustParse(s.config.OrganizationID),
				Name:           trackNumber, // Use track number as name
				Category:       models.CATEGORY_UXV,
				Type:           EntityTypeUAS,
				Status:         threat.Classification, // Use classification as status
				Metadata:       &metadataRaw,
			}

			// Create context with organization ID
			orgCtx := client.WithOrgID(ctx, s.config.OrganizationID)
			createdEntity, err := s.legionClient.CreateEntity(orgCtx, entityReq)
			if err != nil {
				return fmt.Errorf("failed to create UAS entity %s: %w", trackNumber, err)
			}

			// Update the map with the new Legion ID
			delete(s.uasThreats, threat.ID) // Remove old entry
			threat.ID = createdEntity.ID
			s.uasThreats[threat.ID] = threat // Add with new ID

			logger.Infof("🔴 New air track detected: %s", trackNumber)
		}
	}

	logger.Infof("Successfully created %d Counter-UAS systems and %d UAS threats",
		len(s.counterUASSystems), len(s.uasThreats))

	return nil
}

// deployEntities positions entities at their initial locations
func (s *DroneSwarmSimulation) deployEntities(ctx context.Context) error {
	logger.Info("Deploying entities to initial positions...")

	// Convert base location to ECEF
	baseX, baseY, baseZ := latLonAltToECEF(
		s.config.BaseLocation.Lat,
		s.config.BaseLocation.Lon,
		s.config.BaseLocation.Alt,
	)

	// Deploy Counter-UAS systems in defensive ring
	angleStep := 360.0 / float64(s.config.NumCounterUASSystems)
	defenseRadius := 5000.0 // 5km defensive perimeter

	i := 0
	for _, system := range s.counterUASSystems {
		angle := float64(i) * angleStep * math.Pi / 180.0

		// Calculate position on defensive ring
		offsetX := defenseRadius * math.Cos(angle)
		offsetY := defenseRadius * math.Sin(angle)

		system.Position.Coordinates[0] = baseX + offsetX
		system.Position.Coordinates[1] = baseY + offsetY
		system.Position.Coordinates[2] = baseZ + 50 // 50m elevation

		// Update location in Legion
		locationReq := &models.CreateEntityLocationRequest{
			Position: system.Position,
		}

		orgCtx := client.WithOrgID(ctx, s.config.OrganizationID)
		_, err := s.legionClient.CreateEntityLocation(orgCtx, system.ID.String(), locationReq)
		if err != nil {
			return fmt.Errorf("failed to update Counter-UAS location: %w", err)
		}

		i++
	}

	// Deploy UAS threats at 8km radius with attack vectors
	threatRadius := 8000.0 // 8km initial distance - closer for faster engagement

	for _, threat := range s.uasThreats {
		// Random attack vector
		angle := rand.Float64() * 360.0 * math.Pi / 180.0

		// Calculate initial position
		offsetX := threatRadius * math.Cos(angle)
		offsetY := threatRadius * math.Sin(angle)

		// Vary altitude by wave
		altitude := baseZ + 100 + float64(threat.ActualCapabilities.WaveNumber)*50

		threat.Position.Coordinates[0] = baseX + offsetX
		threat.Position.Coordinates[1] = baseY + offsetY
		threat.Position.Coordinates[2] = altitude

		// Calculate velocity towards base (hidden simulation data)
		dx := baseX - threat.Position.Coordinates[0]
		dy := baseY - threat.Position.Coordinates[1]
		dz := baseZ - threat.Position.Coordinates[2]

		// Normalize direction vector
		distance := math.Sqrt(dx*dx + dy*dy + dz*dz)
		velocityMagnitude := threat.ActualCapabilities.SpeedKph / 3.6 // Convert to m/s

		threat.ActualVelocity.Coordinates[0] = (dx / distance) * velocityMagnitude
		threat.ActualVelocity.Coordinates[1] = (dy / distance) * velocityMagnitude
		threat.ActualVelocity.Coordinates[2] = (dz / distance) * velocityMagnitude

		// Update location in Legion
		locationReq := &models.CreateEntityLocationRequest{
			Position: threat.Position,
		}

		orgCtx := client.WithOrgID(ctx, s.config.OrganizationID)
		_, err := s.legionClient.CreateEntityLocation(orgCtx, threat.ID.String(), locationReq)
		if err != nil {
			return fmt.Errorf("failed to update UAS threat location: %w", err)
		}

		// Threats start as PENDING until detected and classified
		// No need to update status here as they're created with PENDING classification
	}

	logger.Info("All entities deployed to initial positions")
	return nil
}

// runSimulationLoop executes the main simulation loop
func (s *DroneSwarmSimulation) runSimulationLoop(ctx context.Context) error {
	logger.Info("Starting main simulation loop...")

	startTime := time.Now()
	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()

	simulationComplete := false

	for !simulationComplete {
		select {
		case <-ctx.Done():
			logger.Info("Simulation cancelled by context")
			return ctx.Err()

		case <-s.stopChan:
			logger.Info("Simulation stopped by user")
			return nil

		case <-ticker.C:
			// Check if simulation duration exceeded
			if time.Since(startTime) > s.config.SimDuration {
				logger.Info("Simulation duration reached")
				simulationComplete = true
				break
			}

			// Execute simulation phases
			if err := s.executeSimulationPhases(ctx); err != nil {
				logger.Errorf("Error executing simulation phases: %v", err)
			}

			// Check termination conditions
			if s.checkTerminationConditions() {
				simulationComplete = true
			}

			// Log progress
			elapsed := time.Since(startTime)
			logger.Infof("Simulation progress: %s / %s", elapsed.Round(time.Second), s.config.SimDuration)
		}
	}

	// Generate After Action Report
	if err := s.generateAAR(); err != nil {
		logger.Errorf("Failed to generate AAR: %v", err)
	}

	logger.Infof("Simulation completed. Outcome: %s", s.stats.SimulationOutcome)
	return nil
}

// executeSimulationPhases runs the 5 phases of the simulation
func (s *DroneSwarmSimulation) executeSimulationPhases(ctx context.Context) error {
	// Phase 1: Swarm Coordination
	if err := s.executeSwarmCoordination(ctx); err != nil {
		return fmt.Errorf("swarm coordination phase failed: %w", err)
	}

	// Phase 2: Movement
	if err := s.executeMovement(ctx); err != nil {
		return fmt.Errorf("movement phase failed: %w", err)
	}

	// Phase 3: Detection
	if err := s.executeDetection(ctx); err != nil {
		return fmt.Errorf("detection phase failed: %w", err)
	}

	// Phase 4: Engagement
	if err := s.executeEngagement(ctx); err != nil {
		return fmt.Errorf("engagement phase failed: %w", err)
	}

	// Phase 5: Resolution
	if err := s.executeResolution(ctx); err != nil {
		return fmt.Errorf("resolution phase failed: %w", err)
	}

	return nil
}

// Phase 1: Swarm Coordination
func (s *DroneSwarmSimulation) executeSwarmCoordination(_ context.Context) error {
	// Update swarm formations and behaviors
	activeThreats := s.getActiveThreats()

	// Group threats by wave (using hidden simulation data)
	waveGroups := make(map[int][]*UASThreat)
	for _, threat := range activeThreats {
		waveGroups[threat.ActualCapabilities.WaveNumber] = append(waveGroups[threat.ActualCapabilities.WaveNumber], threat)
	}

	// Coordinate each wave
	for wave, threats := range waveGroups {
		if len(threats) < 2 {
			continue
		}

		// Calculate center of mass for the wave
		var sumX, sumY, sumZ float64
		for _, threat := range threats {
			sumX += threat.Position.Coordinates[0]
			sumY += threat.Position.Coordinates[1]
			sumZ += threat.Position.Coordinates[2]
		}

		centerX := sumX / float64(len(threats))
		centerY := sumY / float64(len(threats))
		_ = sumZ / float64(len(threats)) // centerZ - not currently used

		// Apply swarm behavior if they're close enough to be identified as a swarm
		for _, threat := range threats {
			// Calculate desired position relative to center
			dx := threat.Position.Coordinates[0] - centerX
			dy := threat.Position.Coordinates[1] - centerY

			desiredDistance := 100.0 // meters
			currentDistance := math.Sqrt(dx*dx + dy*dy)

			if currentDistance > desiredDistance*2 {
				// Apply correction force while maintaining general direction
				// Don't just reduce velocity - add a force towards the swarm center
				correctionFactor := 0.05 // Reduced from 0.1 to be less aggressive

				// Add force towards swarm center
				forceX := -(dx / currentDistance) * correctionFactor * 10.0
				forceY := -(dy / currentDistance) * correctionFactor * 10.0

				threat.ActualVelocity.Coordinates[0] += forceX
				threat.ActualVelocity.Coordinates[1] += forceY
			}
		}

		// Mark threats as part of swarm if close enough (observable)
		if len(threats) >= 3 {
			swarmID := fmt.Sprintf("SWARM-%02d", wave)
			for _, threat := range threats {
				threat.mu.Lock()
				threat.IsPartOfSwarm = true
				threat.SwarmID = &swarmID
				threat.mu.Unlock()
			}
		}

		if s.config.EnableDebugLogging {
			logger.Debugf("Wave %d coordination: %d active threats", wave, len(threats))
		}
	}

	return nil
}

// Phase 2: Movement
func (s *DroneSwarmSimulation) executeMovement(_ context.Context) error {
	// Update UAS threat positions using hidden actual velocity
	for _, threat := range s.uasThreats {
		if threat.Classification == TrackStatusDestroyed || threat.Classification == TrackStatusLost {
			continue
		}

		// Update position based on actual velocity (simulation physics)
		deltaTime := s.config.UpdateInterval.Seconds()

		// Log velocity for debugging if it's too low
		speed := math.Sqrt(
			threat.ActualVelocity.Coordinates[0]*threat.ActualVelocity.Coordinates[0] +
				threat.ActualVelocity.Coordinates[1]*threat.ActualVelocity.Coordinates[1] +
				threat.ActualVelocity.Coordinates[2]*threat.ActualVelocity.Coordinates[2])

		if speed < 5.0 { // Less than 5 m/s is too slow for a drone
			logger.Warnf("Threat %s has very low speed: %.2f m/s, recalculating velocity", threat.TrackNumber, speed)

			// Recalculate velocity towards base
			baseX, baseY, baseZ := latLonAltToECEF(
				s.config.BaseLocation.Lat,
				s.config.BaseLocation.Lon,
				s.config.BaseLocation.Alt,
			)

			dx := baseX - threat.Position.Coordinates[0]
			dy := baseY - threat.Position.Coordinates[1]
			dz := baseZ - threat.Position.Coordinates[2]

			distance := math.Sqrt(dx*dx + dy*dy + dz*dz)
			if distance > 100 { // Only if not already at base
				velocityMagnitude := threat.ActualCapabilities.SpeedKph / 3.6 // Convert to m/s
				threat.ActualVelocity.Coordinates[0] = (dx / distance) * velocityMagnitude
				threat.ActualVelocity.Coordinates[1] = (dy / distance) * velocityMagnitude
				threat.ActualVelocity.Coordinates[2] = (dz / distance) * velocityMagnitude
			}
		}

		threat.Position.Coordinates[0] += threat.ActualVelocity.Coordinates[0] * deltaTime
		threat.Position.Coordinates[1] += threat.ActualVelocity.Coordinates[1] * deltaTime
		threat.Position.Coordinates[2] += threat.ActualVelocity.Coordinates[2] * deltaTime

		// Apply evasion if showing evasive behavior
		if threat.ObservedBehavior == BehaviorEvasive && threat.ActualCapabilities.EvasionCapability {
			s.applyEvasiveManeuvers(threat)
		}

		// Update observed kinematics if being tracked
		if threat.Classification != TrackStatusPending {
			threat.UpdateObservedKinematics(threat.Position)
		}

		// Only queue location update if threat is still active
		if threat.Classification != TrackStatusDestroyed && threat.Classification != TrackStatusLost {
			s.updateBuffer.QueuePositionUpdate(threat.ID, threat.Position)
		}

		threat.LastUpdateTime = time.Now()
	}

	// Counter-UAS systems may update their sensor modes
	for _, system := range s.counterUASSystems {
		// Update heading to track primary target
		if system.EngagedTarget != nil {
			if target, exists := s.uasThreats[*system.EngagedTarget]; exists {
				dx := target.Position.Coordinates[0] - system.Position.Coordinates[0]
				dy := target.Position.Coordinates[1] - system.Position.Coordinates[1]
				system.Heading = math.Atan2(dy, dx) * 180 / math.Pi
				if system.Heading < 0 {
					system.Heading += 360
				}
			}
		}
	}

	return nil
}

// Phase 3: Detection
func (s *DroneSwarmSimulation) executeDetection(_ context.Context) error {
	// For each Counter-UAS system, check for threats in detection range
	for _, system := range s.counterUASSystems {
		if system.Status == CounterUASStatusOffline {
			continue
		}

		detectedThreats := s.detectThreats(system)

		if len(detectedThreats) > 0 {
			if system.Status == CounterUASStatusIdle {
				system.UpdateStatus(CounterUASStatusSearching)
			}

			// Update tracking list
			system.CurrentTargets = make([]uuid.UUID, 0)
			for _, threat := range detectedThreats {
				system.CurrentTargets = append(system.CurrentTargets, threat.ID)
			}

			// Queue status and metadata updates
			s.updateBuffer.QueueStatusUpdate(system.ID, system.Status)
			metadata, _ := json.Marshal(system.GetMetadata())
			s.updateBuffer.QueueMetadataUpdate(system.ID, "metadata", json.RawMessage(metadata))

			// Log detection events and update threat classifications
			for _, threat := range detectedThreats {
				// More aggressive classification based on proximity and behavior
				distance := calculateDistanceKm(system.Position, threat.Position)

				if threat.Classification == TrackStatusPending {
					threat.UpdateClassification(TrackStatusUnknown)
					logger.Infof("🔵 Track %s classification: UNKNOWN - New contact detected at %.1fkm", threat.TrackNumber, distance)
				} else if threat.Classification == TrackStatusUnknown {
					// Within engagement range = definitely hostile
					if distance <= system.EffectiveRange {
						threat.UpdateClassification(TrackStatusHostile)
						logger.Errorf("🔴 Track %s classification: HOSTILE - Within weapons range (%.1fkm)", threat.TrackNumber, distance)
					} else if threat.EstimatedSpeed > 50 || threat.ObservedBehavior == BehaviorAggressive {
						threat.UpdateClassification(TrackStatusSuspected)
						logger.Warnf("🟡 Track %s classification: SUSPECTED - Approaching at %.0f kph", threat.TrackNumber, threat.EstimatedSpeed)
					}
				} else if threat.Classification == TrackStatusSuspected {
					// Upgrade to hostile if getting closer or if engaged
					if distance <= system.EffectiveRange*1.5 || threat.TimesTargeted > 0 {
						threat.UpdateClassification(TrackStatusHostile)
						logger.Errorf("🔴 Track %s classification: HOSTILE - Confirmed enemy asset", threat.TrackNumber)
					}
				}

				// Update observable metadata
				threatMetadata, _ := json.Marshal(threat.GetMetadata())
				s.updateBuffer.QueueStatusUpdate(threat.ID, threat.Classification)
				s.updateBuffer.QueueMetadataUpdate(threat.ID, "metadata", json.RawMessage(threatMetadata))

				// Log detection
				s.simLogger.LogDetection(system.ID, threat.ID,
					"Counter-UAS", "UAS",
					calculateDistanceKm(system.Position, threat.Position)*1000)
			}
		}

		// Clear targets if nothing detected
		if len(detectedThreats) == 0 && len(system.CurrentTargets) > 0 {
			system.CurrentTargets = make([]uuid.UUID, 0)
			if system.Status == CounterUASStatusSearching {
				system.UpdateStatus(CounterUASStatusIdle)
			}
		}
	}

	return nil
}

// Phase 4: Engagement
func (s *DroneSwarmSimulation) executeEngagement(ctx context.Context) error {
	// Use goroutines for concurrent Counter-UAS processing
	var wg sync.WaitGroup
	engagementChan := make(chan *EngagementResult, len(s.counterUASSystems))

	engagementCount := 0
	for _, system := range s.counterUASSystems {
		if system.Status == CounterUASStatusIdle || system.Status == CounterUASStatusOffline ||
			system.Status == CounterUASStatusDegraded || len(system.CurrentTargets) == 0 {
			continue
		}
		engagementCount++

		wg.Add(1)
		go func(sys *CounterUASSystem) {
			defer wg.Done()

			// Find best target
			target := s.selectTarget(sys)
			if target == nil {
				return
			}

			// Check engagement range
			distance := calculateDistanceKm(sys.Position, target.Position)
			if distance > sys.EffectiveRange {
				if s.config.EnableDebugLogging {
					logger.Debugf("%s: Track %s beyond effective range: %.1fkm (max: %.1fkm)",
						sys.Callsign, target.TrackNumber, distance, sys.EffectiveRange)
				}
				return
			}

			// Log engagement attempt
			logger.Infof("🎯 %s (%s) engaging track %s at %.1fkm", sys.Callsign, sys.Name, target.TrackNumber, distance)

			// Engage target
			result := s.engageTarget(sys, target)
			if result == nil {
				logger.Error("engageTarget returned nil result")
				return
			}
			logger.Debugf("Engagement result created: %v", result)
			engagementChan <- result
		}(system)
	}

	logger.Debugf("Started %d engagement goroutines", engagementCount)

	// Process results in a separate goroutine
	resultsChan := make(chan bool)
	go func() {
		for result := range engagementChan {
			if result == nil {
				logger.Error("Received nil engagement result")
				continue
			}
			logger.Infof("📋 Processing engagement result: SystemID=%s, TargetID=%s, success=%v",
				result.SystemID, result.TargetID, result.Success)
			s.processEngagementResult(ctx, result)
		}
		resultsChan <- true
	}()

	// Wait for all engagements to complete
	wg.Wait()
	close(engagementChan)

	// Wait for all results to be processed
	<-resultsChan

	return nil
}

// Phase 5: Resolution
func (s *DroneSwarmSimulation) executeResolution(ctx context.Context) error {
	// Update cooldowns
	for _, system := range s.counterUASSystems {
		if system.CooldownRemaining > 0 {
			system.mu.Lock()
			system.CooldownRemaining--
			if system.CooldownRemaining == 0 && system.Status == CounterUASStatusCooldown {
				system.Status = CounterUASStatusIdle
			}
			system.mu.Unlock()
		}

		// Check ammo depletion
		if system.EngagementType == EngagementTypeKinetic && system.AmmoRemaining == 0 {
			system.UpdateStatus(CounterUASStatusOffline)
			logger.Warnf("⚠️ %s (%s) ammunition depleted - system offline", system.Callsign, system.Name)
		}

		// Check if system is overwhelmed (too many threats in close proximity)
		threatsInRange := 0
		for _, threat := range s.uasThreats {
			if threat.Classification == TrackStatusHostile || threat.Classification == TrackStatusSuspected {
				distance := calculateDistanceKm(system.Position, threat.Position)
				if distance <= system.EffectiveRange*1.2 {
					threatsInRange++
				}
			}
		}

		// System becomes degraded if overwhelmed
		if threatsInRange > 5 && system.Status != CounterUASStatusOffline {
			system.mu.Lock()
			system.SystemHealth = 0.5
			if rand.Float64() < 0.1 { // 10% chance of going offline when overwhelmed
				system.Status = CounterUASStatusOffline
				logger.Errorf("💥 %s (%s) OVERWHELMED - system offline!", system.Callsign, system.Name)
				s.stats.mu.Lock()
				s.stats.CounterUASLosses++
				s.stats.mu.Unlock()
			} else if system.Status != CounterUASStatusDegraded {
				system.Status = CounterUASStatusDegraded
				logger.Warnf("⚠️ %s (%s) under heavy attack - system degraded", system.Callsign, system.Name)
			}
			system.mu.Unlock()
		}

		// Queue status updates for systems
		s.updateBuffer.QueueStatusUpdate(system.ID, system.Status)
		metadata, _ := json.Marshal(system.GetMetadata())
		s.updateBuffer.QueueMetadataUpdate(system.ID, "metadata", json.RawMessage(metadata))
	}

	// Check for mission complete threats
	baseX, baseY, baseZ := latLonAltToECEF(
		s.config.BaseLocation.Lat,
		s.config.BaseLocation.Lon,
		s.config.BaseLocation.Alt,
	)
	basePos := &models.GeomPoint{
		Type:        "Point",
		Coordinates: []float64{baseX, baseY, baseZ},
	}

	for _, threat := range s.uasThreats {
		if threat.Classification == TrackStatusDestroyed || threat.Classification == TrackStatusLost {
			continue
		}

		// Check if threat reached target
		distance := calculateDistanceKm(threat.Position, basePos)
		if distance < 0.5 { // Within 500m of target
			threat.UpdateClassification(TrackStatusLost) // Lost track once it reaches target

			s.stats.mu.Lock()
			s.stats.UASPenetrated++
			s.stats.mu.Unlock()

			// Log mission complete
			logger.Errorf("💥 Track %s reached protected area - MISSION FAILURE", threat.TrackNumber)
			s.simLogger.LogObjective("UAS", "reached_target", "complete", map[string]interface{}{
				"track_id":     threat.ID.String(),
				"track_number": threat.TrackNumber,
			})
		}
	}

	// Flush any pending updates
	if err := s.updateBuffer.Flush(ctx); err != nil {
		logger.Errorf("Failed to flush updates: %v", err)
	}

	// Update statistics
	s.updateStatistics()

	return nil
}

// Helper methods

// getActiveThreats returns all non-eliminated threats
func (s *DroneSwarmSimulation) getActiveThreats() []*UASThreat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	active := make([]*UASThreat, 0)
	for _, threat := range s.uasThreats {
		if threat.Classification != TrackStatusDestroyed && threat.Classification != TrackStatusLost {
			active = append(active, threat)
		}
	}
	return active
}

// detectThreats returns threats within detection range
func (s *DroneSwarmSimulation) detectThreats(system *CounterUASSystem) []*UASThreat {
	detected := make([]*UASThreat, 0)

	for _, threat := range s.uasThreats {
		if threat.Classification == TrackStatusDestroyed || threat.Classification == TrackStatusLost {
			continue
		}

		distance := calculateDistanceKm(system.Position, threat.Position)

		// Different sensors have different ranges
		var detectionRange float64
		switch {
		case threat.RFEmitting && distance <= system.RFDetectionRange:
			detectionRange = system.RFDetectionRange
		case distance <= system.RadarRange:
			detectionRange = system.RadarRange
		case distance <= system.EOIRRange && threat.ThermalSignature:
			detectionRange = system.EOIRRange
		default:
			continue // Not detected
		}

		if distance <= detectionRange {
			// Update track quality based on distance
			threat.mu.Lock()
			threat.TrackQuality = 1.0 - (distance/detectionRange)*0.5
			threat.LastSeenTime = time.Now()
			threat.mu.Unlock()

			detected = append(detected, threat)
		}
	}

	return detected
}

// selectTarget chooses the best target for a Counter-UAS system
func (s *DroneSwarmSimulation) selectTarget(system *CounterUASSystem) *UASThreat {
	threats := s.detectThreats(system)
	if len(threats) == 0 {
		return nil
	}

	// Prioritize by:
	// 1. Already targeted threats (continue engagement)
	// 2. Closest threat
	// 3. Highest autonomy level (more dangerous)

	var bestTarget *UASThreat
	bestScore := -1.0

	for _, threat := range threats {
		score := 0.0

		// Distance factor (closer = higher priority)
		distance := calculateDistanceKm(system.Position, threat.Position)
		distanceScore := 1.0 - (distance / system.RadarRange)
		score += distanceScore * 0.4

		// Threat level factor
		score += float64(threat.ThreatLevel) / 5.0 * 0.3

		// Classification factor (prioritize confirmed hostiles)
		switch threat.Classification {
		case TrackStatusHostile:
			score += 0.3
		case TrackStatusSuspected:
			score += 0.2
		case TrackStatusUnknown:
			score += 0.1
		}

		// Already engaged bonus
		if system.EngagedTarget != nil && *system.EngagedTarget == threat.ID {
			score += 0.2
		}

		if score > bestScore {
			bestScore = score
			bestTarget = threat
		}
	}

	return bestTarget
}

// EngagementResult represents the outcome of an engagement
type EngagementResult struct {
	SystemID   uuid.UUID
	TargetID   uuid.UUID
	Success    bool
	Distance   float64
	EngageType string
}

// engageTarget attempts to engage a threat
func (s *DroneSwarmSimulation) engageTarget(system *CounterUASSystem, target *UASThreat) *EngagementResult {
	system.mu.Lock()
	defer system.mu.Unlock()

	result := &EngagementResult{
		SystemID:   system.ID,
		TargetID:   target.ID,
		Distance:   calculateDistanceKm(system.Position, target.Position),
		EngageType: system.EngagementType,
	}

	// Update status
	system.Status = CounterUASStatusEngaging
	system.EngagedTarget = &target.ID

	// Update threat engagement history
	target.mu.Lock()
	target.TimesTargeted++
	if system.EngagementType == EngagementTypeKinetic {
		target.KineticAttempts++
	} else {
		target.JammingAttempts++
	}
	target.mu.Unlock()

	// Calculate hit probability
	baseProbability := system.SuccessRate

	// Distance modifier
	rangeFactor := 1.0 - (result.Distance / system.EffectiveRange)

	// Evasion modifier (based on observed behavior)
	evasionModifier := 1.0
	if target.ObservedBehavior == BehaviorEvasive {
		evasionModifier = 0.7
	}

	// Size modifier (smaller = harder to hit)
	sizeModifier := 1.0
	switch target.SizeClass {
	case UASSizeGroup1:
		sizeModifier = 0.7
	case UASSizeGroup2:
		sizeModifier = 0.8
	case UASSizeGroup3:
		sizeModifier = 0.9
	}

	// Jamming resistance (for EW attacks)
	jamResistanceModifier := 1.0
	if system.EngagementType == EngagementTypeEW && target.ShowsJamResistance {
		jamResistanceModifier = 0.5
	}

	finalProbability := baseProbability * rangeFactor * evasionModifier * sizeModifier * jamResistanceModifier

	// Roll for success
	if rand.Float64() < finalProbability {
		result.Success = true
		system.SuccessfulEngagements++
	}

	// Update counters
	system.TotalEngagements++

	// Consume ammo for kinetic systems
	if system.EngagementType == EngagementTypeKinetic && system.AmmoRemaining > 0 {
		system.AmmoRemaining--
	}

	// Set cooldown based on reload time
	cooldownTicks := system.ReloadTimeSeconds / int(s.config.UpdateInterval.Seconds())
	if cooldownTicks < 1 {
		cooldownTicks = 1
	}
	system.CooldownRemaining = cooldownTicks

	return result
}

// processEngagementResult handles the outcome of an engagement
func (s *DroneSwarmSimulation) processEngagementResult(_ context.Context, result *EngagementResult) {
	// Get entities with proper locking
	s.mu.RLock()
	threat, threatExists := s.uasThreats[result.TargetID]
	system, systemExists := s.counterUASSystems[result.SystemID]
	s.mu.RUnlock()

	if !threatExists || !systemExists {
		logger.Errorf("Failed to find entities for engagement result: threat=%v, system=%v", threatExists, systemExists)
		return
	}

	s.stats.mu.Lock()
	s.stats.TotalEngagements++
	if result.Success {
		s.stats.SuccessfulEngagements++
		s.stats.UASEliminated++
	}
	s.stats.mu.Unlock()

	if result.Success {
		threat.UpdateClassification(TrackStatusDestroyed)
		logger.Infof("💥 %s (%s) destroyed track %s - SPLASH ONE!", system.Callsign, system.Name, threat.TrackNumber)

		// Update status in Legion to show destroyed
		s.updateBuffer.QueueStatusUpdate(threat.ID, TrackStatusDestroyed)

		// Log elimination
		s.simLogger.LogDestruction(
			result.TargetID,
			"UAS-Threats",
			fmt.Sprintf("destroyed by %s at %.1fkm (%s)",
				system.Callsign,
				result.Distance,
				result.EngageType),
		)
	} else {
		logger.Infof("❌ %s (%s) missed track %s", system.Callsign, system.Name, threat.TrackNumber)

		// Update behavior based on engagement
		threat.mu.Lock()
		if threat.ActualCapabilities.EvasionCapability && rand.Float64() > 0.3 {
			threat.ObservedBehavior = BehaviorEvasive
		}

		// Check for jam resistance
		if result.EngageType == EngagementTypeEW && !result.Success {
			threat.ShowsJamResistance = true
		}
		threat.mu.Unlock()
	}

	// Update system status
	if system.CooldownRemaining > 0 {
		system.UpdateStatus(CounterUASStatusReloading)
	} else {
		system.UpdateStatus(CounterUASStatusTracking)
	}
	system.EngagedTarget = nil // Clear engaged target

	// Log engagement
	s.simLogger.LogEngagement(
		result.SystemID,
		result.TargetID,
		fmt.Sprintf("%s engagement", result.EngageType),
		map[string]interface{}{
			"distance_km": result.Distance,
			"hit":         result.Success,
			"type":        result.EngageType,
		},
	)

	// Queue metadata updates
	s.updateBuffer.QueueStatusUpdate(system.ID, system.Status)
	s.updateBuffer.QueueMetadataUpdate(system.ID, "total_engagements", system.TotalEngagements)
	s.updateBuffer.QueueMetadataUpdate(system.ID, "successful_engagements", system.SuccessfulEngagements)

	if system.EngagementType == EngagementTypeKinetic {
		s.updateBuffer.QueueMetadataUpdate(system.ID, "ammo_remaining", system.AmmoRemaining)
	}

	// Update threat status
	s.updateBuffer.QueueStatusUpdate(threat.ID, threat.Classification)
	threatMetadata, _ := json.Marshal(threat.GetMetadata())
	s.updateBuffer.QueueMetadataUpdate(threat.ID, "metadata", json.RawMessage(threatMetadata))
}

// applyEvasiveManeuvers modifies threat velocity for evasion
func (s *DroneSwarmSimulation) applyEvasiveManeuvers(threat *UASThreat) {
	// Random direction change
	angleChange := (rand.Float64() - 0.5) * 60 * math.Pi / 180 // ±30 degrees

	// Current velocity magnitude
	vMag := math.Sqrt(threat.ActualVelocity.Coordinates[0]*threat.ActualVelocity.Coordinates[0] +
		threat.ActualVelocity.Coordinates[1]*threat.ActualVelocity.Coordinates[1])

	// Current angle
	currentAngle := math.Atan2(threat.ActualVelocity.Coordinates[1], threat.ActualVelocity.Coordinates[0])

	// New angle
	newAngle := currentAngle + angleChange

	// Apply new velocity to actual (hidden) velocity
	threat.ActualVelocity.Coordinates[0] = vMag * math.Cos(newAngle)
	threat.ActualVelocity.Coordinates[1] = vMag * math.Sin(newAngle)

	// Random altitude change
	threat.ActualVelocity.Coordinates[2] = (rand.Float64() - 0.5) * 10 // ±5 m/s vertical
}

// updateStatistics updates simulation statistics
func (s *DroneSwarmSimulation) updateStatistics() {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	// Count active systems
	activeSystems := 0
	for _, system := range s.counterUASSystems {
		if system.Status != CounterUASStatusOffline {
			activeSystems++
		}
	}

	// Count active threats
	activeThreats := len(s.getActiveThreats())

	// Log current status
	logger.Infof("Status: Systems %d/%d active, Threats %d/%d active, Engagements: %d (%d successful)",
		activeSystems, s.config.NumCounterUASSystems,
		activeThreats, s.config.NumUASThreats,
		s.stats.TotalEngagements, s.stats.SuccessfulEngagements)
}

// checkTerminationConditions checks if simulation should end
func (s *DroneSwarmSimulation) checkTerminationConditions() bool {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	// Success: All threats eliminated
	activeThreats := len(s.getActiveThreats())
	if activeThreats == 0 {
		s.stats.SimulationOutcome = "SUCCESS - All threats eliminated"
		logger.Info("Termination condition met: All threats eliminated")
		return true
	}

	// Failure: Too many threats penetrated defenses
	penetrationRate := float64(s.stats.UASPenetrated) / float64(s.config.NumUASThreats)
	if penetrationRate > 0.5 {
		s.stats.SimulationOutcome = "FAILURE - Defenses overwhelmed"
		logger.Info("Termination condition met: Defenses overwhelmed")
		return true
	}

	// Stalemate: All systems depleted but threats remain
	activeSystems := 0
	for _, system := range s.counterUASSystems {
		if system.Status != CounterUASStatusOffline {
			activeSystems++
		}
	}

	if activeSystems == 0 && activeThreats > 0 {
		s.stats.SimulationOutcome = "FAILURE - All defensive systems offline"
		logger.Error("Termination condition met: All defensive systems offline")
		return true
	}

	return false
}

// generateAAR creates the After Action Report
func (s *DroneSwarmSimulation) generateAAR() error {
	logger.Info("Generating After Action Report...")

	// Generate report
	aar, err := s.aarGenerator.GenerateAAR()
	if err != nil {
		return fmt.Errorf("failed to generate AAR: %w", err)
	}

	// Save report
	if err := s.aarGenerator.SaveAAR(aar); err != nil {
		return fmt.Errorf("failed to save AAR: %w", err)
	}

	logger.Info("After Action Report generated successfully")
	return nil
}

// Stop gracefully shuts down the simulation
func (s *DroneSwarmSimulation) Stop() error {
	close(s.stopChan)

	// Cleanup
	if s.updateBuffer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.updateBuffer.Flush(ctx)
	}

	if s.simController != nil {
		_ = s.simController.Stop()
	}

	return nil
}

// Helper functions

// cleanupExistingEntities removes any existing entities with our naming patterns
func (s *DroneSwarmSimulation) cleanupExistingEntities(ctx context.Context) error {
	logger.Info("Cleaning up existing entities...")

	// Reset track number counter to ensure clean start
	atomic.StoreUint32(&trackNumberCounter, 0)

	// Create organization context
	orgCtx := client.WithOrgID(ctx, s.config.OrganizationID)

	// List of entity name patterns to clean up
	patterns := []string{
		"Counter-UAS-",
		"UAS-W",
		"TK-",
	}

	deletedCount := 0

	// Search for each pattern separately to avoid overwhelming the API
	for _, pattern := range patterns {
		searchReq := &models.SearchEntitiesRequest{
			OrganizationID: uuid.MustParse(s.config.OrganizationID),
			Filters: &models.SearchFilters{
				Name: pattern, // This will do a prefix match
			},
		}

		result, err := s.legionClient.SearchEntities(orgCtx, searchReq)
		if err != nil {
			logger.Warnf("Failed to search for entities with pattern %s: %v", pattern, err)
			continue
		}

		// Delete matching entities
		for _, entity := range result.Results {
			if strings.HasPrefix(entity.Name, pattern) {
				if err := s.legionClient.DeleteEntity(orgCtx, entity.ID.String()); err != nil {
					logger.Debugf("Failed to delete entity %s (%s): %v", entity.Name, entity.ID, err)
				} else {
					logger.Debugf("Deleted entity: %s", entity.Name)
					deletedCount++
				}
			}
		}
	}

	if deletedCount > 0 {
		logger.Infof("Cleaned up %d existing entities", deletedCount)
	} else {
		logger.Info("No existing entities found to clean up")
	}

	return nil
}

// init registers the simulation
func init() {
	err := simulation.DefaultRegistry.Register("Drone Swarm Combat", NewDroneSwarmSimulation)
	if err != nil {
		logger.Errorf("Failed to register drone swarm simulation: %v", err)
		return
	}
}
