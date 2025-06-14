package controllers

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/cmd/drone-swarm/core"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
)

// SystemController manages Counter-UAS system detection, tracking, and engagement logic
type SystemController struct {
	counterUASSystems map[uuid.UUID]*CounterUASSystem
	uasThreats        map[uuid.UUID]*UASThreat
	detectionGraph    map[uuid.UUID]map[uuid.UUID]float64 // systemID -> threatID -> distance
	engagementQueue   chan *EngagementRequest
	updateBuffer      *core.UpdateBuffer
	mu                sync.RWMutex
}

// EngagementRequest represents a pending engagement
type EngagementRequest struct {
	SystemID  uuid.UUID
	TargetID  uuid.UUID
	Distance  float64
	Priority  float64
	Timestamp time.Time
}

// DetectionData contains detection information
type DetectionData struct {
	SystemID        uuid.UUID
	DetectedThreats []uuid.UUID
	Distances       map[uuid.UUID]float64
	Timestamp       time.Time
}

// SystemMetrics tracks system performance
type SystemMetrics struct {
	TotalDetections       int
	TotalEngagements      int
	SuccessfulEngagements int
	AverageEngagementTime time.Duration
	SystemUtilization     map[uuid.UUID]float64
}

// NewSystemController creates a new system controller
func NewSystemController() *SystemController {
	return &SystemController{
		counterUASSystems: make(map[uuid.UUID]*CounterUASSystem),
		uasThreats:        make(map[uuid.UUID]*UASThreat),
		detectionGraph:    make(map[uuid.UUID]map[uuid.UUID]float64),
		engagementQueue:   make(chan *EngagementRequest, 100),
	}
}

// Initialize sets up the system controller
func (sc *SystemController) Initialize(ctx context.Context) error {
	logger.Info("Initializing system controller...")

	// Initialize detection graph entries for each system
	for systemID := range sc.counterUASSystems {
		sc.detectionGraph[systemID] = make(map[uuid.UUID]float64)
	}

	return nil
}

// SetSystems sets the Counter-UAS systems to manage
func (sc *SystemController) SetSystems(systems map[uuid.UUID]*CounterUASSystem) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.counterUASSystems = systems
}

// SetThreats sets the UAS threats to track
func (sc *SystemController) SetThreats(threats map[uuid.UUID]*UASThreat) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.uasThreats = threats
}

// SetUpdateBuffer sets the update buffer for entity updates
func (sc *SystemController) SetUpdateBuffer(buffer *core.UpdateBuffer) {
	sc.updateBuffer = buffer
}

// UpdateDetectionGraph updates the detection state for all Counter-UAS systems
func (sc *SystemController) UpdateDetectionGraph() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for systemID, system := range sc.counterUASSystems {
		// Skip depleted systems
		if system.Status == CounterUASStatusDepleted {
			continue
		}

		// Clear previous detections
		sc.detectionGraph[systemID] = make(map[uuid.UUID]float64)

		// Check each threat
		for threatID, threat := range sc.uasThreats {
			// Skip eliminated or jammed threats
			if threat.Status == UASStatusEliminated || threat.Status == UASStatusJammed {
				continue
			}

			// Calculate distance
			distance := calculateDistance(system.Position, threat.Position)

			// Check if within detection range
			if distance <= system.DetectionRadiusKm {
				sc.detectionGraph[systemID][threatID] = distance
			}
		}
	}
}

// GetDetectedThreats returns threats detected by a specific Counter-UAS system
func (sc *SystemController) GetDetectedThreats(systemID uuid.UUID) map[uuid.UUID]float64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if detections, exists := sc.detectionGraph[systemID]; exists {
		// Return a copy to prevent external modification
		copy := make(map[uuid.UUID]float64)
		for k, v := range detections {
			copy[k] = v
		}
		return copy
	}

	return make(map[uuid.UUID]float64)
}

// GetHighestPriorityTarget returns the best target for a Counter-UAS system
func (sc *SystemController) GetHighestPriorityTarget(systemID uuid.UUID) (*UASThreat, float64) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	detections, exists := sc.detectionGraph[systemID]
	if !exists || len(detections) == 0 {
		return nil, 0
	}

	system := sc.counterUASSystems[systemID]
	if system == nil {
		return nil, 0
	}

	var bestTarget *UASThreat
	var bestPriority float64

	for threatID, distance := range detections {
		threat := sc.uasThreats[threatID]
		if threat == nil {
			continue
		}

		// Skip if out of engagement range
		if distance > system.EngagementRadiusKm {
			continue
		}

		// Skip if EW system can't affect high autonomy targets
		if system.EngagementType == "electronic_warfare" && threat.AutonomyLevel >= 0.5 {
			continue
		}

		// Calculate priority
		distancePriority := 1.0 - (distance / system.DetectionRadiusKm)
		speedPriority := threat.SpeedKph / 200.0
		rolePriority := 1.0
		switch threat.FormationRole {
		case "leader":
			rolePriority = 1.5
		case "scout":
			rolePriority = 1.2
		}

		priority := distancePriority*0.5 + speedPriority*0.3 + rolePriority*0.2

		if bestTarget == nil || priority > bestPriority {
			bestTarget = threat
			bestPriority = priority
		}
	}

	return bestTarget, bestPriority
}

// QueueEngagement adds an engagement request to the processing queue
func (sc *SystemController) QueueEngagement(req *EngagementRequest) {
	select {
	case sc.engagementQueue <- req:
		// Successfully queued
	default:
		logger.Warn("Engagement queue full, dropping request")
	}
}

// ProcessEngagementQueue processes pending engagement requests
func (sc *SystemController) ProcessEngagementQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-sc.engagementQueue:
			sc.processEngagementRequest(req)
		}
	}
}

// processEngagementRequest handles a single engagement request
func (sc *SystemController) processEngagementRequest(req *EngagementRequest) {
	sc.mu.RLock()
	system := sc.counterUASSystems[req.SystemID]
	threat := sc.uasThreats[req.TargetID]
	sc.mu.RUnlock()

	if system == nil || threat == nil {
		return
	}

	// Verify system can still engage
	if system.Status != CounterUASStatusEngaging ||
		system.CurrentTarget == nil ||
		*system.CurrentTarget != req.TargetID {
		return
	}

	// Log the engagement attempt
	logger.Infof("Processing engagement: %s -> %s at %.2f km",
		system.Name, threat.Name, req.Distance)
}

// GetSystemMetrics returns performance metrics for the system controller
func (sc *SystemController) GetSystemMetrics() SystemMetrics {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	metrics := SystemMetrics{
		SystemUtilization: make(map[uuid.UUID]float64),
	}

	// Calculate system utilization
	for systemID, system := range sc.counterUASSystems {
		switch system.Status {
		case CounterUASStatusDepleted:
			metrics.SystemUtilization[systemID] = 0.0
		case CounterUASStatusIdle:
			metrics.SystemUtilization[systemID] = 0.2
		case CounterUASStatusTracking:
			metrics.SystemUtilization[systemID] = 0.5
		case CounterUASStatusEngaging, CounterUASStatusCooldown:
			metrics.SystemUtilization[systemID] = 1.0
		default:
			metrics.SystemUtilization[systemID] = 0.0
		}

		metrics.TotalEngagements += system.TotalEngagements
		metrics.SuccessfulEngagements += system.SuccessfulEngagements
	}

	return metrics
}

// OptimizeTargetAssignments optimizes target assignments across all systems
func (sc *SystemController) OptimizeTargetAssignments() map[uuid.UUID]uuid.UUID {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	assignments := make(map[uuid.UUID]uuid.UUID)
	assignedThreats := make(map[uuid.UUID]bool)

	// Simple greedy assignment - each system gets its highest priority unassigned target
	for systemID, system := range sc.counterUASSystems {
		if system.Status != CounterUASStatusIdle {
			continue
		}

		detections := sc.detectionGraph[systemID]
		var bestTarget uuid.UUID
		var bestPriority float64

		for threatID, distance := range detections {
			if assignedThreats[threatID] {
				continue
			}

			threat := sc.uasThreats[threatID]
			if threat == nil || distance > system.EngagementRadiusKm {
				continue
			}

			// Skip if EW can't affect high autonomy
			if system.EngagementType == "electronic_warfare" && threat.AutonomyLevel >= 0.5 {
				continue
			}

			priority := 1.0 - (distance / system.DetectionRadiusKm)
			if priority > bestPriority {
				bestTarget = threatID
				bestPriority = priority
			}
		}

		if bestPriority > 0 {
			assignments[systemID] = bestTarget
			assignedThreats[bestTarget] = true
		}
	}

	return assignments
}

// Helper function to calculate distance between two ECEF points
func calculateDistance(p1, p2 *models.GeomPoint) float64 {
	if p1 == nil || p2 == nil || len(p1.Coordinates) < 3 || len(p2.Coordinates) < 3 {
		return math.MaxFloat64
	}

	dx := p2.Coordinates[0] - p1.Coordinates[0]
	dy := p2.Coordinates[1] - p1.Coordinates[1]
	dz := p2.Coordinates[2] - p1.Coordinates[2]
	return math.Sqrt(dx*dx+dy*dy+dz*dz) / 1000.0 // Convert to km
}
