package dronetornado

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"

	"github.com/picogrid/legion-simulations/pkg/client"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
	"github.com/picogrid/legion-simulations/pkg/simulation"
)

// DroneTornadoSimulation creates drones and moves them in a circle at constant speed
type DroneTornadoSimulation struct {
	config         *Config
	entityIDs      []string
	perDroneRadius []float64
	mu             sync.Mutex
	stopChan       chan struct{}
	startTime      time.Time
}

// NewDroneTornadoSimulation creates a new instance
func NewDroneTornadoSimulation() simulation.Simulation {
	return &DroneTornadoSimulation{
		entityIDs: make([]string, 0),
		stopChan:  make(chan struct{}),
	}
}

// Name returns the simulation name
func (s *DroneTornadoSimulation) Name() string {
	return "Drone Tornado"
}

// Description returns the simulation description
func (s *DroneTornadoSimulation) Description() string {
	return "Creates N drones moving in a circle around a center at a given speed"
}

// Configure sets up the simulation with provided parameters
func (s *DroneTornadoSimulation) Configure(params map[string]interface{}) error {
	cfg, err := ValidateAndParse(params)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	s.config = cfg
	return nil
}

// Run executes the simulation
func (s *DroneTornadoSimulation) Run(ctx context.Context, legionClient *client.Legion) error {
	if s.config == nil {
		return fmt.Errorf("simulation not configured")
	}

	logger.Infof("Starting %s with %d drones, radius=%.1fm, speed=%.1fm/s", s.Name(), s.config.NumDrones, s.config.RadiusMeters, s.config.SpeedMetersPerS)

	// Ensure organization header is set for all API calls
	ctx = client.WithOrgID(ctx, s.config.OrganizationID)

	// Optional cleanup of existing entities that match criteria
	if s.config.CleanupExisting {
		if err := s.cleanupExistingEntities(ctx, legionClient); err != nil {
			logger.Warnf("Cleanup existing entities failed: %v", err)
		}
	}

	// Create entities
	for i := 0; i < s.config.NumDrones; i++ {
		entityID, err := s.createDroneEntity(ctx, legionClient, i)
		if err != nil {
			return fmt.Errorf("failed to create entity %d: %w", i+1, err)
		}
		s.mu.Lock()
		s.entityIDs = append(s.entityIDs, entityID)
		s.mu.Unlock()
	}

	s.startTime = time.Now()

	// Initialize per-drone radii with ±offset randomness, fixed for the run
	s.mu.Lock()
	s.perDroneRadius = make([]float64, len(s.entityIDs))
	r := rand.New(rand.NewSource(s.startTime.UnixNano()))
	for i := range s.entityIDs {
		if s.config.RadiusOffsetM > 0 {
			offset := (r.Float64()*2 - 1) * s.config.RadiusOffsetM
			s.perDroneRadius[i] = s.config.RadiusMeters + offset
		} else {
			s.perDroneRadius[i] = s.config.RadiusMeters
		}
	}
	s.mu.Unlock()

	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()

	timeout := time.After(s.config.Duration)

	// Initial immediate update so they appear placed
	if err := s.updateLocations(ctx, legionClient); err != nil {
		logger.Warnf("Initial location update failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopChan:
			logger.Info("Simulation stopped by user")
			if s.config.DeleteOnExit {
				s.deleteCreatedEntities(ctx, legionClient)
			}
			return nil
		case <-timeout:
			logger.Infof("Simulation completed after %s", s.config.Duration)
			if s.config.DeleteOnExit {
				s.deleteCreatedEntities(ctx, legionClient)
			}
			return nil
		case <-ticker.C:
			if err := s.updateLocations(ctx, legionClient); err != nil {
				logger.Errorf("Error updating locations: %v", err)
			}
		}
	}
}

// Stop gracefully stops the simulation
func (s *DroneTornadoSimulation) Stop() error {
	close(s.stopChan)
	return nil
}

// createDroneEntity creates a single drone entity in Legion
func (s *DroneTornadoSimulation) createDroneEntity(ctx context.Context, legionClient *client.Legion, index int) (string, error) {
	number := index + 1
	name := fmt.Sprintf("Drone %d", number)
	category := models.CategoryDEVICE
	entityType := "Drone"
	status := "ACTIVE"

	// Ensure organization ID valid
	orgUUID, err := uuid.Parse(s.config.OrganizationID)
	if err != nil {
		return "", fmt.Errorf("invalid organization ID: %w", err)
	}
	orgID := strfmt.UUID(orgUUID.String())

	// Try to find existing matching entity by partial name and category/type to avoid duplicate creations
	searchFilters := &models.SearchFilters{
		Name:     name, // partial match supported
		Category: []models.Category{category},
		Type:     entityType,
	}
	searchReq := &models.SearchEntitiesRequest{OrganizationID: &orgID, Filters: searchFilters}
	if resp, err := legionClient.SearchEntities(ctx, searchReq); err == nil && resp != nil && len(resp.Results) > 0 {
		existing := resp.Results[0]
		logger.Infof("Using existing entity: %s (%s)", existing.Name, existing.ID)
		return existing.ID.String(), nil
	}

	// Metadata tag to identify simulation-owned entities
	metadata := map[string]interface{}{
		"sim":          "drone-tornado",
		"drone_number": number,
	}
	metadataJSON, _ := json.Marshal(metadata)
	metadataRaw := json.RawMessage(metadataJSON)

	req := &models.CreateEntityRequest{
		Name:           &name,
		OrganizationID: &orgID,
		Type:           &entityType,
		Category:       &category,
		Status:         &status,
		Metadata:       &metadataRaw,
	}

	logger.Infof("Creating entity: %s", name)
	created, err := legionClient.CreateEntity(ctx, req)
	if err != nil {
		return "", err
	}
	return created.ID.String(), nil
}

// updateLocations updates locations for all drones along the circular path
func (s *DroneTornadoSimulation) updateLocations(ctx context.Context, legionClient *client.Legion) error {
	s.mu.Lock()
	ids := make([]string, len(s.entityIDs))
	copy(ids, s.entityIDs)
	s.mu.Unlock()

	if len(ids) == 0 {
		return nil
	}

	t := time.Since(s.startTime).Seconds()
	omega := s.config.SpeedMetersPerS / s.config.RadiusMeters // rad/s

	// Pre-compute degree offsets per meter
	metersPerDegLat := 111111.0
	metersPerDegLon := 111111.0 * math.Cos(s.config.CenterLat*math.Pi/180.0)

	for i, entityID := range ids {
		phase := float64(i) * 2 * math.Pi / float64(len(ids))
		angle := omega*t + phase

		// Use fixed per-drone radius assigned at start
		radius := s.config.RadiusMeters
		s.mu.Lock()
		if len(s.perDroneRadius) == len(ids) {
			radius = s.perDroneRadius[i]
		}
		s.mu.Unlock()

		dNorth := radius * math.Cos(angle)
		dEast := radius * math.Sin(angle)

		lat := s.config.CenterLat + dNorth/metersPerDegLat
		lon := s.config.CenterLon + dEast/metersPerDegLon
		alt := s.config.CenterAltMeters

		x, y, z := latLonAltToECEF(lat, lon, alt)
		pointType := "Point"
		position := &models.GeomPoint{Type: &pointType, Coordinates: []float64{x, y, z}}

		req := &models.CreateEntityLocationRequest{Position: position, Source: "Drone-Tornado"}
		if _, err := legionClient.CreateEntityLocation(ctx, entityID, req); err != nil {
			logger.Errorf("Failed to update location for entity %s: %v", entityID, err)
			continue
		}
	}
	return nil
}

// cleanupExistingEntities removes pre-existing Drone Tornado-like entities
func (s *DroneTornadoSimulation) cleanupExistingEntities(ctx context.Context, legionClient *client.Legion) error {
	category := models.CategoryDEVICE
	entityType := "Drone"

	orgUUID, err := uuid.Parse(s.config.OrganizationID)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %w", err)
	}
	orgID := strfmt.UUID(orgUUID.String())

	// Search for any entities named with prefix "Drone " and type/category matching
	filters := &models.SearchFilters{
		Name:     "Drone ", // partial match
		Category: []models.Category{category},
		Type:     entityType,
	}
	req := &models.SearchEntitiesRequest{OrganizationID: &orgID, Filters: filters}
	resp, err := legionClient.SearchEntities(ctx, req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.Results) == 0 {
		return nil
	}
	for _, e := range resp.Results {
		if e == nil || e.ID.String() == "" {
			continue
		}
		if err := legionClient.DeleteEntity(ctx, e.ID.String()); err != nil {
			logger.Warnf("Failed to delete entity %s: %v", e.ID, err)
		} else {
			logger.Infof("Deleted old entity: %s (%s)", e.Name, e.ID)
		}
	}
	return nil
}

// deleteCreatedEntities removes entities created during this run
func (s *DroneTornadoSimulation) deleteCreatedEntities(ctx context.Context, legionClient *client.Legion) {
	s.mu.Lock()
	ids := make([]string, len(s.entityIDs))
	copy(ids, s.entityIDs)
	s.mu.Unlock()

	for _, id := range ids {
		if err := legionClient.DeleteEntity(ctx, id); err != nil {
			logger.Warnf("Failed to delete entity %s on exit: %v", id, err)
		}
	}
}

// latLonAltToECEF converts latitude, longitude, altitude to ECEF coordinates
func latLonAltToECEF(lat, lon, alt float64) (x, y, z float64) {
	a := 6378137.0           // WGS84 semi-major axis
	f := 1.0 / 298.257223563 // flattening
	e2 := 2*f - f*f          // first eccentricity squared

	latRad := lat * math.Pi / 180.0
	lonRad := lon * math.Pi / 180.0

	sinLat := math.Sin(latRad)
	N := a / math.Sqrt(1-e2*sinLat*sinLat)

	x = (N + alt) * math.Cos(latRad) * math.Cos(lonRad)
	y = (N + alt) * math.Cos(latRad) * math.Sin(lonRad)
	z = (N*(1-e2) + alt) * math.Sin(latRad)
	return x, y, z
}

// init registers the simulation
func init() {
	if err := simulation.DefaultRegistry.Register("Drone Tornado", NewDroneTornadoSimulation); err != nil {
		logger.Errorf("Failed to register simulation: %v", err)
	}
}
