package reporting

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/pkg/logger"
)

// SimulationLogger handles simulation-specific logging
type SimulationLogger struct {
	simulationID string
	startTime    time.Time
	events       []SimulationEvent
	metrics      map[string]Metric
	mu           sync.RWMutex
}

// SimulationEvent represents a logged simulation event
type SimulationEvent struct {
	Timestamp time.Time
	Type      string
	Severity  string
	TeamName  string
	EntityID  *uuid.UUID
	Message   string
	Details   map[string]interface{}
}

// Metric represents a tracked metric
type Metric struct {
	Name        string
	Value       float64
	Unit        string
	LastUpdated time.Time
	History     []MetricPoint
}

// MetricPoint represents a metric value at a point in time
type MetricPoint struct {
	Timestamp time.Time
	Value     float64
}

// EventType constants
const (
	EventTypeEngagement   = "engagement"
	EventTypeDestruction  = "destruction"
	EventTypeSpawn        = "spawn"
	EventTypeObjective    = "objective"
	EventTypeFormation    = "formation"
	EventTypeSystem       = "system"
	EventTypeTeamStatus   = "team_status"
	EventTypeWaveLaunch   = "wave_launch"
	EventTypeDetection    = "detection"
	EventTypeEvasion      = "evasion"
	EventTypeInterception = "interception"
	EventTypeThreat       = "threat"
	EventTypeCommand      = "command"
)

// Severity constants
const (
	SeverityDebug    = "debug"
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// Color definitions
var (
	colorDebug    = color.New(color.FgHiBlack)
	colorInfo     = color.New(color.FgCyan)
	colorWarning  = color.New(color.FgYellow)
	colorError    = color.New(color.FgRed)
	colorCritical = color.New(color.FgRed, color.Bold)
	colorTeamRed  = color.New(color.FgRed, color.Bold)
	colorTeamBlue = color.New(color.FgBlue, color.Bold)
	colorSuccess  = color.New(color.FgGreen)
)

// NewSimulationLogger creates a new simulation logger
func NewSimulationLogger(simulationID string) *SimulationLogger {
	sl := &SimulationLogger{
		simulationID: simulationID,
		startTime:    time.Now(),
		events:       make([]SimulationEvent, 0),
		metrics:      make(map[string]Metric),
	}

	// Log simulation start
	sl.logColoredMessage(SeverityInfo, "Simulation Started",
		fmt.Sprintf("ID: %s | Time: %s", simulationID, sl.startTime.Format("15:04:05")))

	return sl
}

// LogEngagement logs an engagement event
func (sl *SimulationLogger) LogEngagement(attacker, target uuid.UUID, result string, details map[string]interface{}) {
	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeEngagement,
		Severity:  SeverityInfo,
		EntityID:  &attacker,
		Message:   fmt.Sprintf("Engagement: %s -> %s: %s", attacker, target, result),
		Details:   details,
	})
}

// LogDestruction logs a drone destruction
func (sl *SimulationLogger) LogDestruction(entityID uuid.UUID, teamName string, cause string) {
	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeDestruction,
		Severity:  SeverityWarning,
		TeamName:  teamName,
		EntityID:  &entityID,
		Message:   fmt.Sprintf("Drone destroyed: %s (cause: %s)", entityID, cause),
		Details: map[string]interface{}{
			"cause": cause,
		},
	})

	teamColor := sl.getTeamColor(teamName)
	sl.logColoredMessage(SeverityWarning, "💥 Drone Destroyed",
		fmt.Sprintf("Team: %s | ID: %s | Cause: %s",
			teamColor.Sprint(teamName), entityID.String()[:8], cause))
}

// LogSpawn logs a drone spawn
func (sl *SimulationLogger) LogSpawn(entityID uuid.UUID, teamName string, droneType string) {
	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeSpawn,
		Severity:  SeverityInfo,
		TeamName:  teamName,
		EntityID:  &entityID,
		Message:   fmt.Sprintf("Drone spawned: %s (%s)", entityID, droneType),
		Details: map[string]interface{}{
			"drone_type": droneType,
		},
	})
}

// LogObjective logs an objective event
func (sl *SimulationLogger) LogObjective(teamName string, objectiveType string, status string, details map[string]interface{}) {
	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeObjective,
		Severity:  SeverityInfo,
		TeamName:  teamName,
		Message:   fmt.Sprintf("Objective %s: %s", objectiveType, status),
		Details:   details,
	})
}

// LogTeamStatus logs team status update
func (sl *SimulationLogger) LogTeamStatus(teamName string, activeDrones, totalDrones, losses int) {
	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeTeamStatus,
		Severity:  SeverityInfo,
		TeamName:  teamName,
		Message:   fmt.Sprintf("Team %s: %d/%d active, %d losses", teamName, activeDrones, totalDrones, losses),
		Details: map[string]interface{}{
			"active_drones": activeDrones,
			"total_drones":  totalDrones,
			"losses":        losses,
		},
	})
}

// LogError logs an error event
func (sl *SimulationLogger) LogError(message string, err error, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["error"] = err.Error()

	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeSystem,
		Severity:  SeverityError,
		Message:   message,
		Details:   details,
	})

	logger.Errorf("%s: %v", message, err)
}

// UpdateMetric updates a metric value
func (sl *SimulationLogger) UpdateMetric(name string, value float64, unit string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	metric, exists := sl.metrics[name]
	if !exists {
		metric = Metric{
			Name:    name,
			Unit:    unit,
			History: make([]MetricPoint, 0),
		}
	}

	metric.Value = value
	metric.LastUpdated = time.Now()
	metric.History = append(metric.History, MetricPoint{
		Timestamp: time.Now(),
		Value:     value,
	})

	// Keep only last 1000 points
	if len(metric.History) > 1000 {
		metric.History = metric.History[len(metric.History)-1000:]
	}

	sl.metrics[name] = metric
}

// GetEvents returns all logged events
func (sl *SimulationLogger) GetEvents() []SimulationEvent {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	events := make([]SimulationEvent, len(sl.events))
	copy(events, sl.events)
	return events
}

// GetMetrics returns current metrics
func (sl *SimulationLogger) GetMetrics() map[string]Metric {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	metrics := make(map[string]Metric)
	for k, v := range sl.metrics {
		metrics[k] = v
	}
	return metrics
}

// GetSummary returns a simulation summary
func (sl *SimulationLogger) GetSummary() SimulationSummary {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	duration := time.Since(sl.startTime)

	// Count events by type
	eventCounts := make(map[string]int)
	teamEvents := make(map[string]map[string]int)

	for _, event := range sl.events {
		eventCounts[event.Type]++

		if event.TeamName != "" {
			if teamEvents[event.TeamName] == nil {
				teamEvents[event.TeamName] = make(map[string]int)
			}
			teamEvents[event.TeamName][event.Type]++
		}
	}

	return SimulationSummary{
		SimulationID: sl.simulationID,
		StartTime:    sl.startTime,
		Duration:     duration,
		TotalEvents:  len(sl.events),
		EventCounts:  eventCounts,
		TeamEvents:   teamEvents,
		Metrics:      sl.metrics,
	}
}

// SimulationSummary represents a summary of the simulation
type SimulationSummary struct {
	SimulationID string
	StartTime    time.Time
	Duration     time.Duration
	TotalEvents  int
	EventCounts  map[string]int
	TeamEvents   map[string]map[string]int
	Metrics      map[string]Metric
}

// LogWaveLaunch logs a wave launch event
func (sl *SimulationLogger) LogWaveLaunch(teamName string, waveNumber int, droneCount int, details map[string]interface{}) {
	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeWaveLaunch,
		Severity:  SeverityInfo,
		TeamName:  teamName,
		Message:   fmt.Sprintf("Wave %d launched: %d drones", waveNumber, droneCount),
		Details:   details,
	})

	teamColor := sl.getTeamColor(teamName)
	sl.logColoredMessage(SeverityInfo, "🚀 Wave Launch",
		fmt.Sprintf("Team: %s | Wave: %d | Drones: %d",
			teamColor.Sprint(teamName), waveNumber, droneCount))
}

// LogDetection logs a detection event
func (sl *SimulationLogger) LogDetection(detector, target uuid.UUID, teamName, targetTeam string, distance float64) {
	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeDetection,
		Severity:  SeverityInfo,
		TeamName:  teamName,
		EntityID:  &detector,
		Message:   fmt.Sprintf("Target detected: %s at %.0fm", target.String()[:8], distance),
		Details: map[string]interface{}{
			"target_id":   target,
			"target_team": targetTeam,
			"distance":    distance,
		},
	})

	teamColor := sl.getTeamColor(teamName)
	targetColor := sl.getTeamColor(targetTeam)
	sl.logColoredMessage(SeverityDebug, "👁️ Detection",
		fmt.Sprintf("%s detected %s at %.0fm",
			teamColor.Sprint(teamName), targetColor.Sprint(targetTeam), distance))
}

// LogInterception logs an interception event
func (sl *SimulationLogger) LogInterception(interceptor, target uuid.UUID, teamName string, success bool) {
	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeInterception,
		Severity:  SeverityInfo,
		TeamName:  teamName,
		EntityID:  &interceptor,
		Message: fmt.Sprintf("Interception %s: %s -> %s",
			map[bool]string{true: "successful", false: "failed"}[success],
			interceptor.String()[:8], target.String()[:8]),
		Details: map[string]interface{}{
			"target_id": target,
			"success":   success,
		},
	})

	if success {
		sl.logColoredMessage(SeverityInfo, "🎯 Interception Success",
			fmt.Sprintf("Team: %s | Interceptor: %s",
				sl.getTeamColor(teamName).Sprint(teamName), interceptor.String()[:8]))
	}
}

// LogThreatAssessment logs threat assessment updates
func (sl *SimulationLogger) LogThreatAssessment(teamName string, threatLevel string, threats int) {
	sl.logEvent(SimulationEvent{
		Timestamp: time.Now(),
		Type:      EventTypeThreat,
		Severity:  SeverityInfo,
		TeamName:  teamName,
		Message:   fmt.Sprintf("Threat level: %s (%d active threats)", threatLevel, threats),
		Details: map[string]interface{}{
			"threat_level": threatLevel,
			"threat_count": threats,
		},
	})

	var threatColor *color.Color
	switch threatLevel {
	case "CRITICAL":
		threatColor = colorCritical
	case "HIGH":
		threatColor = colorError
	case "MEDIUM":
		threatColor = colorWarning
	default:
		threatColor = colorInfo
	}

	sl.logColoredMessage(SeverityInfo, "⚠️ Threat Assessment",
		fmt.Sprintf("Team: %s | Level: %s | Active: %d",
			sl.getTeamColor(teamName).Sprint(teamName),
			threatColor.Sprint(threatLevel), threats))
}

// logEvent adds an event to the log
func (sl *SimulationLogger) logEvent(event SimulationEvent) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.events = append(sl.events, event)

	// Keep only last 10000 events to prevent memory issues
	if len(sl.events) > 10000 {
		sl.events = sl.events[len(sl.events)-10000:]
	}
}

// logColoredMessage logs a message with color based on severity
func (sl *SimulationLogger) logColoredMessage(severity, eventType, message string) {
	timestamp := time.Now().Format("15:04:05.000")

	var severityColor *color.Color
	switch severity {
	case SeverityDebug:
		severityColor = colorDebug
	case SeverityInfo:
		severityColor = colorInfo
	case SeverityWarning:
		severityColor = colorWarning
	case SeverityError:
		severityColor = colorError
	case SeverityCritical:
		severityColor = colorCritical
	default:
		severityColor = colorInfo
	}

	fmt.Printf("[%s] %s %s | %s\n",
		timestamp,
		severityColor.Sprint(fmt.Sprintf("%-8s", severity)),
		eventType,
		message)
}

// getTeamColor returns the color for a team
func (sl *SimulationLogger) getTeamColor(teamName string) *color.Color {
	switch teamName {
	case "Red":
		return colorTeamRed
	case "Blue":
		return colorTeamBlue
	default:
		return colorInfo
	}
}

// PrintSummary prints a formatted summary
func (sl *SimulationLogger) PrintSummary() {
	summary := sl.GetSummary()

	colorSuccess.Println("\n╔═══════════════════════════════════════════════════════════╗")
	colorSuccess.Printf("║             SIMULATION SUMMARY - %s             ║\n", summary.SimulationID[:8])
	colorSuccess.Println("╚═══════════════════════════════════════════════════════════╝")

	fmt.Printf("\n📊 Duration: %v | Total Events: %d\n", summary.Duration, summary.TotalEvents)

	fmt.Println("\n📈 Event Distribution:")
	for eventType, count := range summary.EventCounts {
		fmt.Printf("   %-20s: %d\n", eventType, count)
	}

	fmt.Println("\n🏆 Team Performance:")
	for team, events := range summary.TeamEvents {
		teamColor := sl.getTeamColor(team)
		fmt.Printf("\n   %s:\n", teamColor.Sprint(team))
		for eventType, count := range events {
			fmt.Printf("      %-18s: %d\n", eventType, count)
		}
	}

	if len(summary.Metrics) > 0 {
		fmt.Println("\n📊 Performance Metrics:")
		for name, metric := range summary.Metrics {
			fmt.Printf("   %-20s: %.2f %s\n", name, metric.Value, metric.Unit)
		}
	}

	colorSuccess.Println("\n═══════════════════════════════════════════════════════════")
}
