package simple

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/pkg/client"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
	"github.com/picogrid/legion-simulations/pkg/simulation"
)

// Location represents a geographic location
type Location struct {
	City  string
	State string
	Lat   float64
	Lon   float64
}

// SimpleSimulation implements a basic simulation with a few entities
type SimpleSimulation struct {
	config    *Config
	entities  []string   // Entity IDs
	locations []Location // Locations for entities
	mu        sync.Mutex
	stopChan  chan struct{}
}

// NewSimpleSimulation creates a new instance of the simple simulation
func NewSimpleSimulation() simulation.Simulation {
	return &SimpleSimulation{
		entities:  make([]string, 0),
		locations: getDefaultLocations(),
		stopChan:  make(chan struct{}),
	}
}

// getDefaultLocations returns a set of US city locations for entity placement
func getDefaultLocations() []Location {
	return []Location{
		{City: "New York", State: "NY", Lat: 40.7128, Lon: -74.0060},
		{City: "Los Angeles", State: "CA", Lat: 34.0522, Lon: -118.2437},
		{City: "Chicago", State: "IL", Lat: 41.8781, Lon: -87.6298},
		{City: "Houston", State: "TX", Lat: 29.7604, Lon: -95.3698},
		{City: "Phoenix", State: "AZ", Lat: 33.4484, Lon: -112.0740},
		{City: "Philadelphia", State: "PA", Lat: 39.9526, Lon: -75.1652},
		{City: "San Antonio", State: "TX", Lat: 29.4241, Lon: -98.4936},
		{City: "San Diego", State: "CA", Lat: 32.7157, Lon: -117.1611},
		{City: "Dallas", State: "TX", Lat: 32.7767, Lon: -96.7970},
		{City: "San Jose", State: "CA", Lat: 37.3382, Lon: -121.8863},
	}
}

// Name returns the simulation name
func (s *SimpleSimulation) Name() string {
	return "Simple Entity Test"
}

// Description returns the simulation description
func (s *SimpleSimulation) Description() string {
	return "Basic simulation with a few entities for testing Legion connectivity"
}

// Configure sets up the simulation with provided parameters
func (s *SimpleSimulation) Configure(params map[string]interface{}) error {
	config, err := ValidateAndParse(params)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	s.config = config
	return nil
}

// Run executes the simulation
func (s *SimpleSimulation) Run(ctx context.Context, legionClient *client.Legion) error {
	logger.Infof("Starting %s simulation with %d drones", s.Name(), s.config.NumEntities)

	// Add organization ID to context for all API calls
	ctx = client.WithOrgID(ctx, s.config.OrganizationID)

	for i := 0; i < s.config.NumEntities; i++ {
		location := s.locations[i%len(s.locations)]

		entityID, err := s.createEntity(ctx, legionClient, i, location)
		if err != nil {
			return fmt.Errorf("failed to create drone %d: %w", i+1, err)
		}
		s.mu.Lock()
		s.entities = append(s.entities, entityID)
		s.mu.Unlock()
	}

	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()

	timeout := time.After(s.config.Duration)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopChan:
			logger.Info("Simulation stopped by user")
			return nil
		case <-timeout:
			logger.Infof("Simulation completed after %s", s.config.Duration)
			return nil
		case <-ticker.C:
			if err := s.updateLocations(ctx, legionClient); err != nil {
				logger.Errorf("Error updating locations: %v", err)
			}
		}
	}
}

// Stop gracefully shuts down the simulation
func (s *SimpleSimulation) Stop() error {
	close(s.stopChan)
	return nil
}

// getEntityCategory returns the appropriate category for the entity type
func (s *SimpleSimulation) getEntityCategory() string {
	return "UXV"
}

// createEntity creates a single entity in Legion
func (s *SimpleSimulation) createEntity(ctx context.Context, legionClient *client.Legion, index int, location Location) (string, error) {
	droneNumber := index + 1
	droneName := fmt.Sprintf("Simulator Drone %d - %s", droneNumber, location.City)
	category := models.CategoryUXV
	entityType := "UAV"
	status := "ACTIVE"

	orgID, err := uuid.Parse(s.config.OrganizationID)
	if err != nil {
		return "", fmt.Errorf("invalid organization ID: %w", err)
	}
	orgIDStrfmt := strfmt.UUID(orgID.String())

	searchFilters := &models.SearchFilters{
		Name:     droneName,
		Category: []models.Category{category},
	}
	searchReq := &models.SearchEntitiesRequest{
		OrganizationID: &orgIDStrfmt,
		Filters:        searchFilters,
	}

	searchResp, err := legionClient.SearchEntities(ctx, searchReq)
	if err != nil {
		logger.Warnf("Failed to search for existing drone: %v", err)
	} else if searchResp != nil && len(searchResp.Results) > 0 {
		existingDrone := searchResp.Results[0]
		logger.Infof("🚁 Drone %d (%s) Found ✅: %s", droneNumber, location.City, existingDrone.ID)
		return existingDrone.ID.String(), nil
	}

	metadata := map[string]interface{}{
		"manufacturer": "Legion Simulator",
		"model":        "QUAD-X1",
		"serial":       fmt.Sprintf("SIM-%d-%d", droneNumber, time.Now().Unix()),
		"drone_number": droneNumber,
		"home_city":    location.City,
		"home_state":   location.State,
		"home_lat":     location.Lat,
		"home_lon":     location.Lon,
		"capabilities": []string{"telemetry", "gps", "camera", "autonomous_flight"},
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}
	metadataRaw := json.RawMessage(metadataJSON)

	req := &models.CreateEntityRequest{
		Name:           &droneName,
		OrganizationID: &orgIDStrfmt,
		Type:           &entityType,
		Category:       &category,
		Status:         &status,
		Metadata:       &metadataRaw,
	}

	logger.Infof("Creating entity: name=%s, type=%s, category=%s, status=%s, orgId=%s",
		droneName, entityType, category, status, s.config.OrganizationID)

	resp, err := legionClient.CreateEntity(ctx, req)
	if err != nil {
		return "", err
	}

	logger.Infof("🚁 Drone %d (%s) Created ✅: %s", droneNumber, location.City, resp.ID)
	return resp.ID.String(), nil
}

// updateLocations updates the location of all entities
func (s *SimpleSimulation) updateLocations(ctx context.Context, legionClient *client.Legion) error {
	s.mu.Lock()
	entityIDs := make([]string, len(s.entities))
	copy(entityIDs, s.entities)
	s.mu.Unlock()

	logger.Infof("Updating locations for %d entities...", len(entityIDs))

	for i, entityID := range entityIDs {
		// Get the assigned location for this entity
		location := s.locations[i%len(s.locations)]

		// Move in a small circle around the entity's home location
		// Use time-based movement for smooth circular motion
		angle := float64(time.Now().Unix())*0.1 + float64(i)*(2*math.Pi/float64(len(entityIDs)))
		radius := 0.005 // degrees (approximately 500m)

		lat := location.Lat + radius*math.Cos(angle)
		lon := location.Lon + radius*math.Sin(angle)
		alt := 100.0 + float64(i*10) // Different altitudes for each entity

		// Convert lat/lon/alt to ECEF coordinates
		x, y, z := latLonAltToECEF(lat, lon, alt)

		// Create GeoJSON point for position
		pointType := "Point"
		position := &models.GeomPoint{
			Type:        &pointType,
			Coordinates: []float64{x, y, z},
		}

		req := &models.CreateEntityLocationRequest{
			Position: position,
			Source:   "Simple-Simulation",
		}

		_, err := legionClient.CreateEntityLocation(ctx, entityID, req)
		if err != nil {
			logger.Errorf("Failed to update location for entity %s: %v", entityID, err)
			continue
		}
	}

	logger.Info("Location update complete")
	return nil
}

// latLonAltToECEF converts latitude, longitude, altitude to ECEF coordinates
func latLonAltToECEF(lat, lon, alt float64) (x, y, z float64) {
	// WGS84 ellipsoid constants
	a := 6378137.0           // semi-major axis
	f := 1.0 / 298.257223563 // flattening
	e2 := 2*f - f*f          // first eccentricity squared

	// Convert degrees to radians
	latRad := lat * math.Pi / 180.0
	lonRad := lon * math.Pi / 180.0

	// Calculate N - radius of curvature in the prime vertical
	sinLat := math.Sin(latRad)
	N := a / math.Sqrt(1-e2*sinLat*sinLat)

	// Calculate ECEF coordinates
	x = (N + alt) * math.Cos(latRad) * math.Cos(lonRad)
	y = (N + alt) * math.Cos(latRad) * math.Sin(lonRad)
	z = (N*(1-e2) + alt) * math.Sin(latRad)

	return x, y, z
}

// init registers the simulation
func init() {
	err := simulation.DefaultRegistry.Register("Simple Entity Test", NewSimpleSimulation)
	if err != nil {
		logger.Errorf("Failed to register simulation: %v", err)
		return
	}
}
