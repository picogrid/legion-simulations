package simulation

import (
	"fmt"
)

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/pkg/models"
)

// Entity types - Blue Force (friendly) vs Red Force (enemy)
const (
	EntityTypeCounterUAS = "CounterUAS" // Blue Force - our defensive systems
	EntityTypeUAS        = "UAS"        // Red Force - enemy threats
)

// Blue Force Status - Complete visibility of our systems
const (
	CounterUASStatusIdle      = "IDLE"      // System ready, no targets
	CounterUASStatusSearching = "SEARCHING" // Active sensor sweep
	CounterUASStatusTracking  = "TRACKING"  // Tracking detected target
	CounterUASStatusEngaging  = "ENGAGING"  // Weapons release authorized
	CounterUASStatusReloading = "RELOADING" // Kinetic system reloading
	CounterUASStatusCooldown  = "COOLDOWN"  // Post-engagement cooldown
	CounterUASStatusDegraded  = "DEGRADED"  // Partial system failure
	CounterUASStatusOffline   = "OFFLINE"   // System down
)

// Red Force Track Classification - What we can determine about enemies
const (
	TrackStatusPending   = "PENDING"   // Initial detection, classifying
	TrackStatusUnknown   = "UNKNOWN"   // Detected but not identified
	TrackStatusSuspected = "SUSPECTED" // Exhibiting hostile behavior
	TrackStatusHostile   = "HOSTILE"   // Confirmed enemy asset
	TrackStatusNeutral   = "NEUTRAL"   // Identified as non-threat
	TrackStatusLost      = "LOST"      // Lost track of target
	TrackStatusDestroyed = "DESTROYED" // Confirmed kill
)

// Engagement types
const (
	EngagementTypeKinetic = "kinetic"
	EngagementTypeEW      = "electronic_warfare"
)

// UAS Size Classifications (DoD Group System)
const (
	UASSizeGroup1 = "GROUP_1" // < 20 lbs, < 1,200 ft AGL
	UASSizeGroup2 = "GROUP_2" // 21-55 lbs, < 3,500 ft AGL
	UASSizeGroup3 = "GROUP_3" // < 1,320 lbs, < 18,000 ft MSL
	UASSizeGroup4 = "GROUP_4" // > 1,320 lbs, < 18,000 ft MSL
	UASSizeGroup5 = "GROUP_5" // > 1,320 lbs, > 18,000 ft MSL
)

// Threat Behavior Patterns (Observable)
const (
	BehaviorSurveillance = "SURVEILLANCE" // Loitering, circling patterns
	BehaviorAggressive   = "AGGRESSIVE"   // Direct approach, high speed
	BehaviorEvasive      = "EVASIVE"      // Erratic movement when targeted
	BehaviorFormation    = "FORMATION"    // Moving in coordinated group
	BehaviorUnknown      = "UNKNOWN"      // No clear pattern
)

// CounterUASSystem represents a BLUE FORCE defensive Counter-UAS system
// We have complete visibility and control of these systems
type CounterUASSystem struct {
	ID       uuid.UUID
	Name     string
	Callsign string // Military callsign
	Status   string
	Position *models.GeomPoint
	Heading  float64 // Degrees

	// Sensor Capabilities
	RadarRange        float64 // Detection range for radar
	EOIRRange         float64 // Electro-optical/infrared range
	RFDetectionRange  float64 // RF signal detection range
	CurrentSensorMode string  // RADAR, EO/IR, RF, MULTI

	// Weapon Systems
	EngagementType    string  // kinetic or electronic_warfare
	EffectiveRange    float64 // Maximum engagement range
	AmmoCapacity      int
	AmmoRemaining     int
	SuccessRate       float64
	ReloadTimeSeconds int
	CooldownRemaining int

	// Operational Data
	SystemHealth          float64 // 0.0 to 1.0
	PowerLevel            float64 // Battery/generator percentage
	TotalEngagements      int
	SuccessfulEngagements int
	CurrentTargets        []uuid.UUID // Can track multiple
	EngagedTarget         *uuid.UUID  // Currently engaging

	// C2 Integration
	DataLinkStatus string // ONLINE, DEGRADED, OFFLINE
	LastC2Update   time.Time
	IFFCode        string // Identification Friend or Foe

	LastUpdateTime time.Time
	mu             sync.RWMutex
}

// UASThreat represents a RED FORCE enemy drone
// We only know what we can observe/detect
type UASThreat struct {
	ID             uuid.UUID // Our tracking ID
	TrackNumber    string    // Military track number (e.g., "TK-4521")
	Classification string    // PENDING, UNKNOWN, SUSPECTED, HOSTILE

	// Observable Characteristics
	Position     *models.GeomPoint // Last known position
	LastSeenTime time.Time         // When last detected
	TrackQuality float64           // 0.0-1.0 confidence in track

	// Estimated from observations
	EstimatedSpeed    float64 // Calculated from position changes
	EstimatedHeading  float64 // Degrees, calculated from movement
	EstimatedAltitude float64 // Meters AGL

	// Size Classification (from radar/visual)
	SizeClass         string  // GROUP_1 through GROUP_5
	RadarCrossSection float64 // Square meters

	// Behavioral Analysis
	ObservedBehavior string  // SURVEILLANCE, AGGRESSIVE, etc.
	ThreatLevel      int     // 1-5 scale
	IsPartOfSwarm    bool    // Multiple tracks moving together
	SwarmID          *string // If part of detected swarm

	// Sensor Detections
	RFEmitting        bool     // Detected RF emissions
	RFFrequency       *float64 // If detected, MHz
	ThermalSignature  bool     // IR detection
	AcousticSignature bool     // Audio detection

	// Engagement History
	TimesTargeted      int  // How many times we've engaged
	JammingAttempts    int  // EW attempts
	KineticAttempts    int  // Kinetic attempts
	ShowsJamResistance bool // Didn't respond to jamming

	// For simulation purposes only (hidden from C2 display)
	ActualVelocity     *models.GeomPoint     // True velocity for physics
	ActualCapabilities SimulatedCapabilities // Hidden true capabilities

	LastUpdateTime time.Time
	mu             sync.RWMutex
}

// SimulatedCapabilities holds the true capabilities of threats (hidden from C2)
type SimulatedCapabilities struct {
	SpeedKph          float64
	AutonomyLevel     float64 // 0.0-1.0 for simulation mechanics
	EvasionCapability bool
	PayloadType       string // For simulation narrative
	WaveNumber        int    // Which attack wave
}

// NewCounterUASSystem creates a new BLUE FORCE Counter-UAS system
func NewCounterUASSystem(name string, position *models.GeomPoint, engagementType string) *CounterUASSystem {
	// Generate military callsign
	callsigns := []string{"HAWK", "EAGLE", "SENTRY", "GUARDIAN", "DEFENDER"}
	callsign := fmt.Sprintf("%s-%02d", callsigns[rand.Intn(len(callsigns))], rand.Intn(99)+1)

	// Assign capabilities based on engagement type
	var successRate float64
	var ammoCapacity int
	var reloadTime int
	var effectiveRange float64

	if engagementType == EngagementTypeKinetic {
		successRate = 0.7 + rand.Float64()*0.2    // 0.7-0.9
		ammoCapacity = 20 + rand.Intn(20)         // 20-40 rounds
		reloadTime = 30 + rand.Intn(30)           // 30-60 seconds
		effectiveRange = 3.0 + rand.Float64()*2.0 // 3-5 km
	} else {
		successRate = 0.5 + rand.Float64()*0.2 // 0.5-0.7
		ammoCapacity = -1                      // Unlimited for EW
		reloadTime = 5                         // Quick reset
		effectiveRange = 2.0 + rand.Float64()  // 2-3 km
	}

	return &CounterUASSystem{
		ID:       uuid.New(),
		Name:     name,
		Callsign: callsign,
		Status:   CounterUASStatusIdle,
		Position: position,
		Heading:  rand.Float64() * 360,

		// Sensor suite
		RadarRange:        12.0, // 12km radar detection
		EOIRRange:         8.0,  // 8km EO/IR
		RFDetectionRange:  15.0, // 15km RF detection
		CurrentSensorMode: "MULTI",

		// Weapons
		EngagementType:    engagementType,
		EffectiveRange:    effectiveRange,
		AmmoCapacity:      ammoCapacity,
		AmmoRemaining:     ammoCapacity,
		SuccessRate:       successRate,
		ReloadTimeSeconds: reloadTime,
		CooldownRemaining: 0,

		// System status
		SystemHealth:          1.0, // 100% healthy
		PowerLevel:            1.0, // 100% power
		TotalEngagements:      0,
		SuccessfulEngagements: 0,
		CurrentTargets:        make([]uuid.UUID, 0),
		EngagedTarget:         nil,

		// C2 Integration
		DataLinkStatus: "ONLINE",
		LastC2Update:   time.Now(),
		IFFCode:        fmt.Sprintf("BLUE-%04d", rand.Intn(9999)),

		LastUpdateTime: time.Now(),
	}
}

// NewUASThreat creates a new RED FORCE threat (with limited observable data)
func NewUASThreat(trackNumber string, position *models.GeomPoint, waveNumber int) *UASThreat {
	// Hidden true characteristics (for simulation)
	trueSpeed := 50.0 + rand.Float64()*150.0  // 50-200 kph
	autonomyLevel := rand.Float64()           // 0.0-1.0
	evasionCapability := rand.Float64() > 0.3 // 70% have evasion

	// Determine size class based on random distribution
	sizeRoll := rand.Float64()
	var sizeClass string
	var radarCrossSection float64

	if sizeRoll < 0.4 {
		sizeClass = UASSizeGroup1
		radarCrossSection = 0.01 + rand.Float64()*0.04 // 0.01-0.05 m²
	} else if sizeRoll < 0.7 {
		sizeClass = UASSizeGroup2
		radarCrossSection = 0.05 + rand.Float64()*0.15 // 0.05-0.2 m²
	} else if sizeRoll < 0.9 {
		sizeClass = UASSizeGroup3
		radarCrossSection = 0.2 + rand.Float64()*0.3 // 0.2-0.5 m²
	} else {
		sizeClass = UASSizeGroup4
		radarCrossSection = 0.5 + rand.Float64()*0.5 // 0.5-1.0 m²
	}

	// Initial velocity (hidden from C2)
	heading := rand.Float64() * 360.0
	velocityMagnitude := trueSpeed / 3.6 // Convert to m/s
	headingRad := heading * math.Pi / 180.0

	velocity := &models.GeomPoint{
		Type: "Point",
		Coordinates: []float64{
			velocityMagnitude * math.Cos(headingRad),
			velocityMagnitude * math.Sin(headingRad),
			0,
		},
	}

	// RF emissions (60% of drones emit RF)
	var rfFreq *float64
	rfEmitting := rand.Float64() < 0.6
	if rfEmitting {
		freq := 2400.0 + rand.Float64()*100.0 // 2.4-2.5 GHz typical
		rfFreq = &freq
	}

	return &UASThreat{
		ID:             uuid.New(),
		TrackNumber:    trackNumber,
		Classification: TrackStatusPending,

		// Initial observations
		Position:     position,
		LastSeenTime: time.Now(),
		TrackQuality: 1.0, // Perfect initial track

		// Initial estimates (will be refined)
		EstimatedSpeed:    0, // Unknown until movement observed
		EstimatedHeading:  0, // Unknown until movement observed
		EstimatedAltitude: position.Coordinates[2],

		// Size characteristics
		SizeClass:         sizeClass,
		RadarCrossSection: radarCrossSection,

		// Initial behavior unknown
		ObservedBehavior: BehaviorUnknown,
		ThreatLevel:      3, // Medium until proven otherwise
		IsPartOfSwarm:    false,

		// Sensor data
		RFEmitting:        rfEmitting,
		RFFrequency:       rfFreq,
		ThermalSignature:  true,                       // All drones have heat signature
		AcousticSignature: sizeClass != UASSizeGroup1, // Larger = louder

		// No engagement history yet
		TimesTargeted:      0,
		JammingAttempts:    0,
		KineticAttempts:    0,
		ShowsJamResistance: false,

		// Hidden simulation data
		ActualVelocity: velocity,
		ActualCapabilities: SimulatedCapabilities{
			SpeedKph:          trueSpeed,
			AutonomyLevel:     autonomyLevel,
			EvasionCapability: evasionCapability,
			PayloadType:       "surveillance", // Could expand this
			WaveNumber:        waveNumber,
		},

		LastUpdateTime: time.Now(),
	}
}

// GetMetadata returns the metadata map for a BLUE FORCE Counter-UAS system
func (c *CounterUASSystem) GetMetadata() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metadata := map[string]interface{}{
		// Identity
		"callsign": c.Callsign,
		"iff_code": c.IFFCode,

		// Sensors
		"radar_range_km":        c.RadarRange,
		"eoir_range_km":         c.EOIRRange,
		"rf_detection_range_km": c.RFDetectionRange,
		"sensor_mode":           c.CurrentSensorMode,

		// Weapons
		"engagement_type":    c.EngagementType,
		"effective_range_km": c.EffectiveRange,
		"success_rate":       c.SuccessRate,
		"cooldown_remaining": c.CooldownRemaining,

		// System Status
		"system_health":   c.SystemHealth,
		"power_level":     c.PowerLevel,
		"datalink_status": c.DataLinkStatus,

		// Combat Stats
		"total_engagements":      c.TotalEngagements,
		"successful_engagements": c.SuccessfulEngagements,
		"tracking_targets":       len(c.CurrentTargets),

		// UI Proximity Circles (for rendering detection/engagement ranges)
		"detection_radius_km":  c.RadarRange,                                                      // Primary detection range for UI
		"engagement_radius_km": c.EffectiveRange,                                                  // Engagement range for UI
		"max_sensor_range_km":  math.Max(math.Max(c.RadarRange, c.EOIRRange), c.RFDetectionRange), // Maximum sensor range
	}

	if c.EngagementType == EngagementTypeKinetic {
		metadata["ammo_capacity"] = c.AmmoCapacity
		metadata["ammo_remaining"] = c.AmmoRemaining
		metadata["reload_time_sec"] = c.ReloadTimeSeconds
	}

	if c.EngagedTarget != nil {
		metadata["engaged_target"] = c.EngagedTarget.String()
	}

	return metadata
}

// GetMetadata returns observable metadata for a RED FORCE threat
func (u *UASThreat) GetMetadata() map[string]interface{} {
	u.mu.RLock()
	defer u.mu.RUnlock()

	metadata := map[string]interface{}{
		// Track Info
		"track_number":   u.TrackNumber,
		"classification": u.Classification,
		"track_quality":  u.TrackQuality,
		"last_seen":      u.LastSeenTime.Format(time.RFC3339),

		// Observable Characteristics
		"size_class":          u.SizeClass,
		"radar_cross_section": u.RadarCrossSection,
		"observed_behavior":   u.ObservedBehavior,
		"threat_level":        u.ThreatLevel,

		// Estimated Kinematics
		"estimated_speed_kph": u.EstimatedSpeed,
		"estimated_heading":   u.EstimatedHeading,
		"estimated_altitude":  u.EstimatedAltitude,

		// Sensor Detections
		"rf_emitting":        u.RFEmitting,
		"thermal_signature":  u.ThermalSignature,
		"acoustic_signature": u.AcousticSignature,

		// Engagement History
		"times_targeted":       u.TimesTargeted,
		"jamming_attempts":     u.JammingAttempts,
		"kinetic_attempts":     u.KineticAttempts,
		"shows_jam_resistance": u.ShowsJamResistance,
	}

	if u.RFFrequency != nil {
		metadata["rf_frequency_mhz"] = *u.RFFrequency
	}

	if u.IsPartOfSwarm && u.SwarmID != nil {
		metadata["swarm_id"] = *u.SwarmID
	}

	return metadata
}

// UpdateStatus safely updates the status of a BLUE FORCE system
func (c *CounterUASSystem) UpdateStatus(newStatus string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Status = newStatus
	c.LastUpdateTime = time.Now()
	c.LastC2Update = time.Now()
}

// UpdateClassification safely updates the classification of a RED FORCE track
func (u *UASThreat) UpdateClassification(newClass string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Classification = newClass
	u.LastUpdateTime = time.Now()

	// Update threat level based on classification
	switch newClass {
	case TrackStatusHostile:
		u.ThreatLevel = 5
	case TrackStatusSuspected:
		u.ThreatLevel = 4
	case TrackStatusUnknown:
		u.ThreatLevel = 3
	case TrackStatusNeutral:
		u.ThreatLevel = 1
	}
}

// UpdateObservedKinematics updates estimated movement data from observations
func (u *UASThreat) UpdateObservedKinematics(newPos *models.GeomPoint) {
	u.mu.Lock()
	defer u.mu.Unlock()

	timeDelta := time.Since(u.LastSeenTime).Seconds()
	if timeDelta > 0 && u.Position != nil {
		// Calculate observed speed and heading
		dx := newPos.Coordinates[0] - u.Position.Coordinates[0]
		dy := newPos.Coordinates[1] - u.Position.Coordinates[1]
		dz := newPos.Coordinates[2] - u.Position.Coordinates[2]

		distance := math.Sqrt(dx*dx + dy*dy + dz*dz)
		u.EstimatedSpeed = (distance / timeDelta) * 3.6 // Convert m/s to kph

		// Calculate heading
		u.EstimatedHeading = math.Atan2(dy, dx) * 180 / math.Pi
		if u.EstimatedHeading < 0 {
			u.EstimatedHeading += 360
		}

		// Update behavior based on movement pattern
		if u.EstimatedSpeed > 150 {
			u.ObservedBehavior = BehaviorAggressive
		} else if u.EstimatedSpeed < 20 {
			u.ObservedBehavior = BehaviorSurveillance
		}
	}

	u.Position = newPos
	u.EstimatedAltitude = newPos.Coordinates[2]
	u.LastSeenTime = time.Now()
	u.LastUpdateTime = time.Now()
}

// Location represents a geographic location
type Location struct {
	Lat float64
	Lon float64
	Alt float64
}

// Helper functions for coordinate conversion and distance calculations

// latLonAltToECEF converts latitude, longitude, and altitude to ECEF coordinates
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

// calculateDistance3D calculates the 3D Euclidean distance between two ECEF points
func calculateDistance3D(p1, p2 *models.GeomPoint) float64 {
	dx := p2.Coordinates[0] - p1.Coordinates[0]
	dy := p2.Coordinates[1] - p1.Coordinates[1]
	dz := p2.Coordinates[2] - p1.Coordinates[2]
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// calculateDistanceKm calculates the distance in kilometers between two ECEF points
func calculateDistanceKm(p1, p2 *models.GeomPoint) float64 {
	return calculateDistance3D(p1, p2) / 1000.0
}

// normalizeVector normalizes a 3D vector to unit length
func normalizeVector(v *models.GeomPoint) *models.GeomPoint {
	magnitude := math.Sqrt(v.Coordinates[0]*v.Coordinates[0] + v.Coordinates[1]*v.Coordinates[1] + v.Coordinates[2]*v.Coordinates[2])
	if magnitude == 0 {
		return &models.GeomPoint{
			Type:        "Point",
			Coordinates: []float64{0, 0, 0},
		}
	}
	return &models.GeomPoint{
		Type: "Point",
		Coordinates: []float64{
			v.Coordinates[0] / magnitude,
			v.Coordinates[1] / magnitude,
			v.Coordinates[2] / magnitude,
		},
	}
}
