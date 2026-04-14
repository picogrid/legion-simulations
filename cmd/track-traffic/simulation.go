package tracktraffic

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/picogrid/legion-simulations/pkg/client"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
	"github.com/picogrid/legion-simulations/pkg/simulation"
)

type routePattern string

const (
	routeCircle    routePattern = "circle"
	routeEllipse   routePattern = "ellipse"
	routeFigure8   routePattern = "figure8"
	routeLissajous routePattern = "lissajous"
	routeWalker    routePattern = "walker"
	routeJitter    routePattern = "jitter"
)

type trafficTrackSpec struct {
	Name            string
	Type            string
	Status          string
	Affiliation     models.Affiliation
	Source          string
	Pattern         routePattern
	SpeedMetersPerS float64
	AnchorNorthM    float64
	AnchorEastM     float64
	RadiusNorthM    float64
	RadiusEastM     float64
	BaseAltitudeM   float64
	Metadata        map[string]interface{}
}

type trackTemplate struct {
	NamePrefix      string
	Type            string
	Status          string
	Affiliation     models.Affiliation
	Source          string
	Pattern         routePattern
	SpeedMetersPerS float64
	RadiusNorthM    float64
	RadiusEastM     float64
	BaseAltitudeM   float64
	Metadata        map[string]interface{}
}

type gridSlot struct {
	NorthM float64
	EastM  float64
	Row    int
	Col    int
}

type createdTrack struct {
	ID   string
	Spec trafficTrackSpec
}

// TrackTrafficSimulation creates a spread-out set of tracks with different movement patterns.
type TrackTrafficSimulation struct {
	config      *Config
	tracks      []createdTrack
	startTime   time.Time
	stopChan    chan struct{}
	stopOnce    sync.Once
	cleanupOnce sync.Once
	mu          sync.Mutex
}

// NewTrackTrafficSimulation creates a new instance of the track traffic simulation.
func NewTrackTrafficSimulation() simulation.Simulation {
	return &TrackTrafficSimulation{
		tracks:   make([]createdTrack, 0),
		stopChan: make(chan struct{}),
	}
}

func (s *TrackTrafficSimulation) Name() string {
	return "Track Traffic Demo"
}

func (s *TrackTrafficSimulation) Description() string {
	return "Creates smooth moving tracks with pre-seeded history for map playback testing"
}

func (s *TrackTrafficSimulation) Configure(params map[string]interface{}) error {
	cfg, err := ValidateAndParse(params)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	s.config = cfg
	return nil
}

func (s *TrackTrafficSimulation) Run(ctx context.Context, legionClient *client.Legion) error {
	if s.config == nil {
		return fmt.Errorf("simulation not configured")
	}

	ctx = client.WithOrgID(ctx, s.config.OrganizationID)
	s.startTime = time.Now().UTC()
	defer s.cleanupTracks(legionClient)

	specs := s.defaultTrackSpecs()
	logger.Infof("Starting %s with %d tracks (max concurrency %d)", s.Name(), len(specs), s.config.MaxConcurrency)

	if err := s.createTracksConcurrently(ctx, legionClient, specs); err != nil {
		return fmt.Errorf("failed to create tracks: %w", err)
	}

	if err := s.seedHistory(ctx, legionClient, s.startTime); err != nil {
		return fmt.Errorf("failed to seed historical track locations: %w", err)
	}

	if err := s.appendCurrentLocations(ctx, legionClient, s.startTime); err != nil {
		return fmt.Errorf("failed to write initial track locations: %w", err)
	}

	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()

	timeout := time.After(s.config.Duration)

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				return nil
			}
			return ctx.Err()
		case <-s.stopChan:
			logger.Info("Simulation stopped by user")
			return nil
		case <-timeout:
			logger.Infof("Simulation completed after %s", s.config.Duration)
			return nil
		case tickTime := <-ticker.C:
			if err := s.appendCurrentLocations(ctx, legionClient, tickTime.UTC()); err != nil {
				logger.Errorf("Failed to append track locations: %v", err)
			}
		}
	}
}

func (s *TrackTrafficSimulation) Stop() error {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
	return nil
}

func (s *TrackTrafficSimulation) createTrack(ctx context.Context, legionClient *client.Legion, spec trafficTrackSpec) (string, error) {
	orgUUID, err := uuid.Parse(s.config.OrganizationID)
	if err != nil {
		return "", fmt.Errorf("invalid organization ID: %w", err)
	}

	metadataJSON, err := json.Marshal(spec.Metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}
	metadataRaw := json.RawMessage(metadataJSON)
	name := spec.Name
	entityType := spec.Type
	status := spec.Status
	category := models.CategoryTRACK

	req := &models.CreateEntityRequest{
		Affiliation:    spec.Affiliation,
		Category:       &category,
		Metadata:       &metadataRaw,
		Name:           &name,
		OrganizationID: &orgUUID,
		Status:         &status,
		Type:           &entityType,
	}

	created, err := legionClient.CreateEntity(ctx, req)
	if err != nil {
		return "", err
	}

	logger.Infof("Created track entity %s (%s)", spec.Name, created.ID)
	return created.ID.String(), nil
}

func (s *TrackTrafficSimulation) createTracksConcurrently(ctx context.Context, legionClient *client.Legion, specs []trafficTrackSpec) error {
	results := make([]createdTrack, len(specs))

	err := s.runBounded(ctx, len(specs), func(index int) error {
		trackID, createErr := s.createTrack(ctx, legionClient, specs[index])
		if createErr != nil {
			return fmt.Errorf("%s: %w", specs[index].Name, createErr)
		}
		results[index] = createdTrack{ID: trackID, Spec: specs[index]}
		return nil
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.tracks = append(s.tracks[:0], results...)
	s.mu.Unlock()

	return nil
}

func (s *TrackTrafficSimulation) seedHistory(ctx context.Context, legionClient *client.Legion, now time.Time) error {
	tracks := s.snapshotTracks()

	return s.runBounded(ctx, len(tracks), func(index int) error {
		track := tracks[index]
		for i := s.config.HistoryPoints; i >= 1; i-- {
			recordedAt := now.Add(-time.Duration(i) * s.config.HistoryStep)
			location := s.buildTrackLocation(track.Spec, recordedAt)
			if _, err := legionClient.CreateEntityLocation(ctx, track.ID, location); err != nil {
				return fmt.Errorf("track %s: %w", track.Spec.Name, err)
			}
		}
		return nil
	})
}

func (s *TrackTrafficSimulation) appendCurrentLocations(ctx context.Context, legionClient *client.Legion, recordedAt time.Time) error {
	tracks := s.snapshotTracks()
	return s.runBounded(ctx, len(tracks), func(index int) error {
		track := tracks[index]
		req := s.buildTrackLocation(track.Spec, recordedAt)
		if _, err := legionClient.CreateEntityLocation(ctx, track.ID, req); err != nil {
			return fmt.Errorf("track %s: %w", track.Spec.Name, err)
		}
		return nil
	})
}

func (s *TrackTrafficSimulation) cleanupTracks(legionClient *client.Legion) {
	if !s.config.DeleteOnExit {
		return
	}

	s.cleanupOnce.Do(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cleanupCtx = client.WithOrgID(cleanupCtx, s.config.OrganizationID)

		tracks := s.snapshotTracks()
		if err := s.runBounded(cleanupCtx, len(tracks), func(index int) error {
			track := tracks[index]
			if err := legionClient.DeleteEntity(cleanupCtx, track.ID); err != nil {
				logger.Warnf("Failed to delete track %s (%s): %v", track.Spec.Name, track.ID, err)
				return nil
			}
			logger.Infof("Deleted track %s (%s)", track.Spec.Name, track.ID)
			return nil
		}); err != nil {
			logger.Warnf("Cleanup encountered an error: %v", err)
		}
	})
}

func (s *TrackTrafficSimulation) snapshotTracks() []createdTrack {
	s.mu.Lock()
	defer s.mu.Unlock()

	tracks := make([]createdTrack, len(s.tracks))
	copy(tracks, s.tracks)
	return tracks
}

func (s *TrackTrafficSimulation) buildTrackLocation(spec trafficTrackSpec, recordedAt time.Time) *models.CreateEntityLocationRequest {
	elapsed := recordedAt.Sub(s.startTime).Seconds()
	offsetNorth, offsetEast, altOffset := routeOffsets(spec, elapsed)
	lat, lon := offsetLatLon(
		s.config.CenterLat,
		s.config.CenterLon,
		spec.AnchorNorthM+offsetNorth,
		spec.AnchorEastM+offsetEast,
	)
	alt := s.config.CenterAltMeters + spec.BaseAltitudeM + altOffset

	x, y, z := latLonAltToECEF(lat, lon, alt)
	pointType := "Point"

	return &models.CreateEntityLocationRequest{
		Position: &models.GeomPoint{
			Type:        &pointType,
			Coordinates: []float64{x, y, z},
		},
		Source:     spec.Source,
		RecordedAt: &recordedAt,
	}
}

func (s *TrackTrafficSimulation) defaultTrackSpecs() []trafficTrackSpec {
	templates := s.baseTrackTemplates()
	slots := jitteredGridSlots(s.config.TotalTracks, s.config.GridSpacingM, s.config.GridJitterM)
	specs := make([]trafficTrackSpec, 0, s.config.TotalTracks)

	for i := 0; i < s.config.TotalTracks; i++ {
		template := templates[i%len(templates)]
		slot := slots[i]
		spec := trafficTrackSpec{
			Name:            fmt.Sprintf("%s %03d", template.NamePrefix, i+1),
			Type:            template.Type,
			Status:          template.Status,
			Affiliation:     template.Affiliation,
			Source:          template.Source,
			Pattern:         template.Pattern,
			SpeedMetersPerS: template.SpeedMetersPerS,
			AnchorNorthM:    slot.NorthM,
			AnchorEastM:     slot.EastM,
			RadiusNorthM:    template.RadiusNorthM,
			RadiusEastM:     template.RadiusEastM,
			BaseAltitudeM:   template.BaseAltitudeM,
			Metadata:        cloneMetadata(template.Metadata),
		}
		spec.Metadata["instance_index"] = i + 1
		spec.Metadata["grid_row"] = slot.Row
		spec.Metadata["grid_col"] = slot.Col
		specs = append(specs, spec)
	}

	return specs
}

func (s *TrackTrafficSimulation) baseTrackTemplates() []trackTemplate {
	return []trackTemplate{
		{
			NamePrefix:      "Track Demo Quadcopter",
			Type:            "UAS",
			Status:          "ACTIVE",
			Affiliation:     models.AffiliationFRIEND,
			Source:          "Track-Traffic-Sim",
			Pattern:         routeCircle,
			SpeedMetersPerS: 15,
			RadiusNorthM:    90,
			RadiusEastM:     90,
			BaseAltitudeM:   80,
			Metadata: map[string]interface{}{
				"sim":                "track-traffic",
				"profile":            "small-uas",
				"display_speed_mps":  15,
				"movement_pattern":   "circle",
				"historical_enabled": true,
			},
		},
		{
			NamePrefix:      "Track Demo VTOL",
			Type:            "UAS",
			Status:          "ACTIVE",
			Affiliation:     models.AffiliationFRIEND,
			Source:          "Track-Traffic-Sim",
			Pattern:         routeFigure8,
			SpeedMetersPerS: 28,
			RadiusNorthM:    120,
			RadiusEastM:     75,
			BaseAltitudeM:   120,
			Metadata: map[string]interface{}{
				"sim":                "track-traffic",
				"profile":            "vtol",
				"display_speed_mps":  28,
				"movement_pattern":   "figure8",
				"historical_enabled": true,
			},
		},
		{
			NamePrefix:      "Track Demo Fixed Wing",
			Type:            "UAS",
			Status:          "ACTIVE",
			Affiliation:     models.AffiliationFRIEND,
			Source:          "Track-Traffic-Sim",
			Pattern:         routeEllipse,
			SpeedMetersPerS: 55,
			RadiusNorthM:    180,
			RadiusEastM:     110,
			BaseAltitudeM:   450,
			Metadata: map[string]interface{}{
				"sim":                "track-traffic",
				"profile":            "fixed-wing",
				"display_speed_mps":  55,
				"movement_pattern":   "ellipse",
				"historical_enabled": true,
			},
		},
		{
			NamePrefix:      "Track Demo Fast Mover",
			Type:            "UAS",
			Status:          "ACTIVE",
			Affiliation:     models.AffiliationFRIEND,
			Source:          "Track-Traffic-Sim",
			Pattern:         routeLissajous,
			SpeedMetersPerS: 120,
			RadiusNorthM:    220,
			RadiusEastM:     140,
			BaseAltitudeM:   900,
			Metadata: map[string]interface{}{
				"sim":                "track-traffic",
				"profile":            "high-speed-fixed-wing",
				"display_speed_mps":  120,
				"movement_pattern":   "lissajous",
				"historical_enabled": true,
			},
		},
		{
			NamePrefix:      "Track Demo Walker",
			Type:            "Human",
			Status:          "ACTIVE",
			Affiliation:     models.AffiliationFRIEND,
			Source:          "Track-Traffic-Sim",
			Pattern:         routeWalker,
			SpeedMetersPerS: 5.0 / 3.6,
			RadiusNorthM:    30,
			RadiusEastM:     18,
			BaseAltitudeM:   2,
			Metadata: map[string]interface{}{
				"sim":                "track-traffic",
				"profile":            "walking-human",
				"display_speed_kph":  5,
				"movement_pattern":   "walking-loop",
				"historical_enabled": true,
			},
		},
		{
			NamePrefix:      "Track Demo Pedestrian",
			Type:            "Human",
			Status:          "ACTIVE",
			Affiliation:     models.AffiliationFRIEND,
			Source:          "Track-Traffic-Sim",
			Pattern:         routeFigure8,
			SpeedMetersPerS: 4.6 / 3.6,
			RadiusNorthM:    24,
			RadiusEastM:     16,
			BaseAltitudeM:   2,
			Metadata: map[string]interface{}{
				"sim":                "track-traffic",
				"profile":            "walking-human",
				"display_speed_kph":  4.6,
				"movement_pattern":   "figure8-walk",
				"historical_enabled": true,
			},
		},
		{
			NamePrefix:      "Track Demo Camera Jitter",
			Type:            "Camera",
			Status:          "ACTIVE",
			Affiliation:     models.AffiliationFRIEND,
			Source:          "Track-Traffic-Sim",
			Pattern:         routeJitter,
			SpeedMetersPerS: 0,
			RadiusNorthM:    10,
			RadiusEastM:     10,
			BaseAltitudeM:   8,
			Metadata: map[string]interface{}{
				"sim":                "track-traffic",
				"profile":            "stationary-camera",
				"actual_motion":      "stationary",
				"movement_pattern":   "gps-jitter",
				"jitter_radius_m":    10,
				"historical_enabled": true,
			},
		},
	}
}

func routeOffsets(spec trafficTrackSpec, elapsedSeconds float64) (northM, eastM, altM float64) {
	switch spec.Pattern {
	case routeCircle:
		period := loopPeriod(spec.SpeedMetersPerS, spec.RadiusNorthM, spec.RadiusEastM)
		theta := 2 * math.Pi * elapsedSeconds / period
		return spec.RadiusNorthM * math.Cos(theta), spec.RadiusEastM * math.Sin(theta), 8 * math.Sin(theta*0.5)
	case routeEllipse:
		period := loopPeriod(spec.SpeedMetersPerS, spec.RadiusNorthM, spec.RadiusEastM)
		theta := 2 * math.Pi * elapsedSeconds / period
		return spec.RadiusNorthM * math.Cos(theta), spec.RadiusEastM * math.Sin(theta), 25 * math.Sin(theta*0.7)
	case routeFigure8:
		period := loopPeriod(maxFloat(spec.SpeedMetersPerS, 0.5), maxFloat(spec.RadiusNorthM, 10), maxFloat(spec.RadiusEastM, 10))
		theta := 2 * math.Pi * elapsedSeconds / period
		return spec.RadiusNorthM * math.Sin(theta), spec.RadiusEastM * math.Sin(theta) * math.Cos(theta), 5 * math.Sin(theta*2)
	case routeLissajous:
		period := loopPeriod(maxFloat(spec.SpeedMetersPerS, 0.5), maxFloat(spec.RadiusNorthM, 10), maxFloat(spec.RadiusEastM, 10))
		theta := 2 * math.Pi * elapsedSeconds / period
		return spec.RadiusNorthM * math.Sin(theta), spec.RadiusEastM * math.Sin(2*theta+math.Pi/3), 40 * math.Sin(theta*1.5)
	case routeWalker:
		period := loopPeriod(maxFloat(spec.SpeedMetersPerS, 0.1), maxFloat(spec.RadiusNorthM, 5), maxFloat(spec.RadiusEastM, 5))
		theta := 2 * math.Pi * elapsedSeconds / period
		return spec.RadiusNorthM * math.Sin(theta), spec.RadiusEastM * math.Sin(theta*0.5), 0
	case routeJitter:
		north := spec.RadiusNorthM * (0.55*math.Sin(elapsedSeconds/5.0) + 0.25*math.Sin(elapsedSeconds/1.7+1.2) + 0.2*math.Sin(elapsedSeconds/0.8+2.4))
		east := spec.RadiusEastM * (0.5*math.Sin(elapsedSeconds/4.2+0.3) + 0.3*math.Sin(elapsedSeconds/1.3+2.0) + 0.2*math.Sin(elapsedSeconds/0.7+0.9))
		return north, east, 0
	default:
		return 0, 0, 0
	}
}

func loopPeriod(speedMetersPerS, radiusNorthM, radiusEastM float64) float64 {
	circumference := 2 * math.Pi * math.Sqrt((radiusNorthM*radiusNorthM+radiusEastM*radiusEastM)/2)
	return maxFloat(circumference/maxFloat(speedMetersPerS, 0.5), 5)
}

func offsetLatLon(baseLat, baseLon, northM, eastM float64) (float64, float64) {
	metersPerDegLat := 111111.0
	metersPerDegLon := 111111.0 * math.Cos(baseLat*math.Pi/180.0)
	if math.Abs(metersPerDegLon) < 1 {
		metersPerDegLon = 1
	}

	lat := baseLat + northM/metersPerDegLat
	lon := baseLon + eastM/metersPerDegLon
	return lat, lon
}

func latLonAltToECEF(lat, lon, alt float64) (x, y, z float64) {
	a := 6378137.0
	f := 1.0 / 298.257223563
	e2 := 2*f - f*f

	latRad := lat * math.Pi / 180.0
	lonRad := lon * math.Pi / 180.0

	sinLat := math.Sin(latRad)
	n := a / math.Sqrt(1-e2*sinLat*sinLat)

	x = (n + alt) * math.Cos(latRad) * math.Cos(lonRad)
	y = (n + alt) * math.Cos(latRad) * math.Sin(lonRad)
	z = (n*(1-e2) + alt) * math.Sin(latRad)
	return x, y, z
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func jitteredGridSlots(totalTracks int, gridSpacingM, gridJitterM float64) []gridSlot {
	cols := int(math.Ceil(math.Sqrt(float64(totalTracks))))
	rows := int(math.Ceil(float64(totalTracks) / float64(cols)))
	slots := make([]gridSlot, 0, totalTracks)

	for idx := 0; idx < totalTracks; idx++ {
		row := idx / cols
		col := idx % cols

		north := (float64(row) - float64(rows-1)/2.0) * gridSpacingM
		east := (float64(col) - float64(cols-1)/2.0) * gridSpacingM
		if row%2 == 1 {
			east += gridSpacingM * 0.18
		}

		north += gridJitterM * deterministicSpread(idx, 1)
		east += gridJitterM * deterministicSpread(idx, 2)

		slots = append(slots, gridSlot{
			NorthM: north,
			EastM:  east,
			Row:    row,
			Col:    col,
		})
	}

	return slots
}

func deterministicSpread(index, salt int) float64 {
	value := math.Sin(float64((index+1)*(salt*97+37))) * 43758.5453123
	fraction := value - math.Floor(value)
	return fraction*2 - 1
}

func cloneMetadata(metadata map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

func (s *TrackTrafficSimulation) runBounded(ctx context.Context, itemCount int, fn func(index int) error) error {
	if itemCount == 0 {
		return nil
	}

	workerCount := minInt(itemCount, maxInt(s.config.MaxConcurrency, 1))
	jobs := make(chan int)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for index := range jobs {
			if err := fn(index); err != nil {
				select {
				case errCh <- err:
				default:
				}
			}
		}
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

sendLoop:
	for index := 0; index < itemCount; index++ {
		select {
		case <-ctx.Done():
			break sendLoop
		case err := <-errCh:
			close(jobs)
			wg.Wait()
			return err
		case jobs <- index:
		}
	}
	close(jobs)
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
	}

	return ctx.Err()
}

func init() {
	if err := simulation.DefaultRegistry.Register("Track Traffic Demo", NewTrackTrafficSimulation); err != nil {
		logger.Errorf("Failed to register simulation: %v", err)
	}
}
