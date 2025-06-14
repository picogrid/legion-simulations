package core

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EngagementCalculator handles Counter-UAS vs UAS engagement calculations
type EngagementCalculator struct {
	kineticSuccessRange [2]float64 // min, max success rates for kinetic
	ewSuccessRange      [2]float64 // min, max success rates for electronic warfare
	autonomyThreshold   float64    // autonomy level below which EW is effective
	mu                  sync.RWMutex
}

// EngagementResult represents the outcome of an engagement
type EngagementResult struct {
	AttackerID        uuid.UUID
	TargetID          uuid.UUID
	Success           bool
	EngagementType    string
	Distance          float64
	TargetAutonomy    float64
	TargetNeutralized bool
	Timestamp         time.Time
}

// NewEngagementCalculator creates a new engagement calculator
func NewEngagementCalculator() *EngagementCalculator {
	return &EngagementCalculator{
		kineticSuccessRange: [2]float64{0.7, 0.9}, // 70-90% success rate for kinetic
		ewSuccessRange:      [2]float64{0.5, 0.7}, // 50-70% success rate for EW
		autonomyThreshold:   0.5,                  // EW only affects drones with autonomy < 0.5
	}
}

// CalculateEngagement determines the outcome of a Counter-UAS engagement against a UAS threat
func (ec *EngagementCalculator) CalculateEngagement(
	attacker CounterUASInfo,
	target UASInfo,
	distance float64,
	environmental Modifiers,
) *EngagementResult {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	// Check if within engagement range
	if distance > attacker.EngagementRangeKm {
		return &EngagementResult{
			AttackerID:     attacker.ID,
			TargetID:       target.ID,
			Success:        false,
			EngagementType: attacker.EngagementType,
			Distance:       distance,
			TargetAutonomy: target.AutonomyLevel,
			Timestamp:      time.Now(),
		}
	}

	// Calculate base success probability based on engagement type
	var baseSuccessProb float64

	switch attacker.EngagementType {
	case "kinetic":
		// Kinetic engagement: 70-90% success rate
		baseSuccessProb = attacker.SuccessRate
	case "electronic_warfare":
		// EW engagement: Check if target can be jammed
		if target.AutonomyLevel >= ec.autonomyThreshold {
			// High autonomy drones are immune to jamming
			return &EngagementResult{
				AttackerID:     attacker.ID,
				TargetID:       target.ID,
				Success:        false,
				EngagementType: attacker.EngagementType,
				Distance:       distance,
				TargetAutonomy: target.AutonomyLevel,
				Timestamp:      time.Now(),
			}
		}
		// EW is effective against low autonomy targets
		baseSuccessProb = attacker.SuccessRate
	}

	// Apply environmental and distance modifiers
	successProb := ec.applyModifiers(baseSuccessProb, distance, attacker.EngagementRangeKm, environmental)

	// Roll for success
	success := rand.Float64() < successProb

	result := &EngagementResult{
		AttackerID:        attacker.ID,
		TargetID:          target.ID,
		Success:           success,
		EngagementType:    attacker.EngagementType,
		Distance:          distance,
		TargetAutonomy:    target.AutonomyLevel,
		TargetNeutralized: success,
		Timestamp:         time.Now(),
	}

	return result
}

// applyModifiers applies environmental and distance modifiers to success probability
func (ec *EngagementCalculator) applyModifiers(
	baseProb float64,
	distance float64,
	maxRange float64,
	environmental Modifiers,
) float64 {
	prob := baseProb

	// Distance modifier (linear falloff)
	distanceRatio := distance / maxRange
	distanceMod := 1.0 - (distanceRatio * 0.3) // 30% reduction at max range
	prob *= distanceMod

	// Environmental modifiers
	prob *= environmental.Visibility
	prob *= environmental.Weather
	prob *= environmental.Terrain

	// Speed modifier (harder to hit fast-moving targets)
	if environmental.TargetSpeed > 0 {
		// Convert km/h to m/s for calculation
		speedMs := environmental.TargetSpeed / 3.6
		speedMod := 1.0 / (1.0 + speedMs/50.0) // Significant reduction for fast targets
		prob *= speedMod
	}

	// Evasion modifier
	if environmental.TargetEvading {
		prob *= 0.7 // 30% reduction for evading targets
	}

	// Clamp between 0 and 1
	return math.Max(0.0, math.Min(1.0, prob))
}

// CounterUASInfo contains Counter-UAS system information for engagement calculations
type CounterUASInfo struct {
	ID                uuid.UUID
	EngagementType    string
	EngagementRangeKm float64
	SuccessRate       float64
	AmmoRemaining     int
	CooldownRemaining int
}

// UASInfo contains UAS threat information for engagement calculations
type UASInfo struct {
	ID                uuid.UUID
	AutonomyLevel     float64
	SpeedKph          float64
	EvasionCapability bool
	Status            string
}

// Modifiers contains environmental and situational modifiers
type Modifiers struct {
	Visibility    float64 // 0.0 to 1.0 (1.0 = perfect visibility)
	Weather       float64 // 0.0 to 1.0 (1.0 = clear weather)
	Terrain       float64 // 0.0 to 1.0 (1.0 = open terrain)
	TargetSpeed   float64 // km/h
	TargetEvading bool    // Whether target is actively evading
}

// CanEngage checks if a Counter-UAS system can engage a UAS threat
func (ec *EngagementCalculator) CanEngage(attacker CounterUASInfo, target UASInfo, distance float64) bool {
	// Check range
	if distance > attacker.EngagementRangeKm {
		return false
	}

	// Check if Counter-UAS is ready
	if attacker.CooldownRemaining > 0 {
		return false
	}

	// Check ammo for kinetic systems
	if attacker.EngagementType == "kinetic" && attacker.AmmoRemaining <= 0 {
		return false
	}

	// Check if EW can affect this target
	if attacker.EngagementType == "electronic_warfare" && target.AutonomyLevel >= ec.autonomyThreshold {
		return false
	}

	return true
}

// GetSuccessRate returns a random success rate within the configured range for the engagement type
func (ec *EngagementCalculator) GetSuccessRate(engagementType string) float64 {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	var minRate, maxRate float64
	switch engagementType {
	case "kinetic":
		minRate, maxRate = ec.kineticSuccessRange[0], ec.kineticSuccessRange[1]
	case "electronic_warfare":
		minRate, maxRate = ec.ewSuccessRange[0], ec.ewSuccessRange[1]
	default:
		return 0.5 // Default for unknown types
	}

	return minRate + rand.Float64()*(maxRate-minRate)
}

// UpdateConfiguration allows updating the engagement calculator configuration
func (ec *EngagementCalculator) UpdateConfiguration(kineticMin, kineticMax, ewMin, ewMax, autonomyThreshold float64) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.kineticSuccessRange = [2]float64{kineticMin, kineticMax}
	ec.ewSuccessRange = [2]float64{ewMin, ewMax}
	ec.autonomyThreshold = autonomyThreshold
}
