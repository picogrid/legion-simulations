package reporting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/pkg/logger"
)

// AARGenerator generates After Action Reports
type AARGenerator struct {
	logger *SimulationLogger
	config AARConfig
}

// AARConfig configures AAR generation
type AARConfig struct {
	OutputDir        string
	Format           string // "json", "html", "markdown"
	IncludeGraphs    bool
	DetailLevel      string                 // "summary", "detailed", "full"
	SimulationConfig map[string]interface{} // Configuration used for the simulation
}

// AAR represents an After Action Report
type AAR struct {
	Metadata        AARMetadata             `json:"metadata"`
	Summary         ExecutiveSummary        `json:"summary"`
	Timeline        []TimelineEntry         `json:"timeline"`
	TeamAnalysis    map[string]TeamAnalysis `json:"team_analysis"`
	Engagements     EngagementAnalysis      `json:"engagements"`
	Performance     PerformanceAnalysis     `json:"performance"`
	ThreatAnalysis  ThreatAnalysis          `json:"threat_analysis"`
	SystemAnalysis  SystemAnalysis          `json:"system_analysis"`
	EventLog        []EventLogEntry         `json:"event_log"`
	Statistics      SummaryStatistics       `json:"statistics"`
	Recommendations []Recommendation        `json:"recommendations"`
	Lessons         []LessonLearned         `json:"lessons_learned"`
}

// AARMetadata contains report metadata
type AARMetadata struct {
	SimulationID    string    `json:"simulation_id"`
	GeneratedAt     time.Time `json:"generated_at"`
	SimulationStart time.Time `json:"simulation_start"`
	SimulationEnd   time.Time `json:"simulation_end"`
	Duration        string    `json:"duration"`
	Version         string    `json:"version"`
}

// ExecutiveSummary provides high-level overview
type ExecutiveSummary struct {
	Outcome          string   `json:"outcome"`
	WinningTeam      string   `json:"winning_team"`
	TotalEngagements int      `json:"total_engagements"`
	TotalLosses      int      `json:"total_losses"`
	KeyEvents        []string `json:"key_events"`
}

// TimelineEntry represents an event in the timeline
type TimelineEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	ElapsedTime string                 `json:"elapsed_time"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`
	Details     map[string]interface{} `json:"details"`
}

// TeamAnalysis contains team-specific analysis
type TeamAnalysis struct {
	TeamName            string                `json:"team_name"`
	FinalStatus         string                `json:"final_status"`
	InitialStrength     int                   `json:"initial_strength"`
	FinalStrength       int                   `json:"final_strength"`
	Losses              int                   `json:"losses"`
	Kills               int                   `json:"kills"`
	EffectivenessRating float64               `json:"effectiveness_rating"`
	DronePerformance    map[string]DroneStats `json:"drone_performance"`
	TacticalAnalysis    TacticalAnalysis      `json:"tactical_analysis"`
}

// DroneStats contains statistics for a drone type
type DroneStats struct {
	Type            string  `json:"type"`
	Deployed        int     `json:"deployed"`
	Survived        int     `json:"survived"`
	Kills           int     `json:"kills"`
	EngagementRatio float64 `json:"engagement_ratio"`
	SurvivalRate    float64 `json:"survival_rate"`
}

// TacticalAnalysis contains tactical performance metrics
type TacticalAnalysis struct {
	FormationMaintenance float64 `json:"formation_maintenance"`
	ObjectiveCompletion  float64 `json:"objective_completion"`
	ResponseTime         float64 `json:"avg_response_time_ms"`
	Coordination         float64 `json:"coordination_score"`
}

// EngagementAnalysis contains engagement statistics
type EngagementAnalysis struct {
	TotalEngagements       int            `json:"total_engagements"`
	SuccessfulHits         int            `json:"successful_hits"`
	HitRate                float64        `json:"hit_rate"`
	AverageEngagementRange float64        `json:"avg_engagement_range_m"`
	EngagementsByType      map[string]int `json:"engagements_by_type"`
	EngagementHeatmap      []HeatmapPoint `json:"engagement_heatmap"`
}

// HeatmapPoint represents a location with engagement intensity
type HeatmapPoint struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Intensity   float64 `json:"intensity"`
	Engagements int     `json:"engagements"`
}

// PerformanceAnalysis contains system performance metrics
type PerformanceAnalysis struct {
	AverageUpdateTime   float64            `json:"avg_update_time_ms"`
	PeakEntityCount     int                `json:"peak_entity_count"`
	TotalAPIRequests    int                `json:"total_api_requests"`
	APIErrorRate        float64            `json:"api_error_rate"`
	SimulationStability float64            `json:"simulation_stability"`
	ResourceUtilization map[string]float64 `json:"resource_utilization"`
}

// ThreatAnalysis contains threat assessment data
type ThreatAnalysis struct {
	TotalThreatsIdentified int            `json:"total_threats_identified"`
	ThreatsNeutralized     int            `json:"threats_neutralized"`
	AverageThreatDuration  string         `json:"avg_threat_duration"`
	ThreatsByType          map[string]int `json:"threats_by_type"`
	ThreatTimeline         []ThreatEvent  `json:"threat_timeline"`
	PeakThreatLevel        string         `json:"peak_threat_level"`
}

// ThreatEvent represents a threat detection event
type ThreatEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	ThreatLevel string    `json:"threat_level"`
	Team        string    `json:"team"`
	ThreatCount int       `json:"threat_count"`
}

// SystemAnalysis contains system performance analysis
type SystemAnalysis struct {
	CommunicationReliability float64         `json:"communication_reliability"`
	CommandLatency           float64         `json:"avg_command_latency_ms"`
	SensorAccuracy           float64         `json:"sensor_accuracy"`
	WeaponSystemEfficiency   float64         `json:"weapon_system_efficiency"`
	AutonomyPerformance      float64         `json:"autonomy_performance"`
	SystemFailures           []SystemFailure `json:"system_failures"`
}

// SystemFailure represents a system failure event
type SystemFailure struct {
	Timestamp    time.Time `json:"timestamp"`
	System       string    `json:"system"`
	Description  string    `json:"description"`
	Impact       string    `json:"impact"`
	RecoveryTime string    `json:"recovery_time"`
}

// EventLogEntry represents a detailed event log entry
type EventLogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Entity      string                 `json:"entity,omitempty"`
	Team        string                 `json:"team,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// SummaryStatistics contains overall simulation statistics
type SummaryStatistics struct {
	TotalDronesDeployed     int                `json:"total_drones_deployed"`
	TotalDronesLost         int                `json:"total_drones_lost"`
	TotalEngagements        int                `json:"total_engagements"`
	SuccessfulInterceptions int                `json:"successful_interceptions"`
	AverageMissionDuration  string             `json:"avg_mission_duration"`
	PeakConcurrentDrones    int                `json:"peak_concurrent_drones"`
	ResourceUtilization     map[string]float64 `json:"resource_utilization"`
}

// Recommendation represents an improvement recommendation
type Recommendation struct {
	Priority        string `json:"priority"` // "High", "Medium", "Low"
	Category        string `json:"category"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	ExpectedBenefit string `json:"expected_benefit"`
}

// LessonLearned represents an insight from the simulation
type LessonLearned struct {
	Category       string `json:"category"`
	Observation    string `json:"observation"`
	Impact         string `json:"impact"`
	Recommendation string `json:"recommendation"`
}

// NewAARGenerator creates a new AAR generator
func NewAARGenerator(logger *SimulationLogger, config AARConfig) *AARGenerator {
	return &AARGenerator{
		logger: logger,
		config: config,
	}
}

// GenerateAAR creates an After Action Report
func (g *AARGenerator) GenerateAAR() (*AAR, error) {
	summary := g.logger.GetSummary()
	events := g.logger.GetEvents()

	aar := &AAR{
		Metadata: AARMetadata{
			SimulationID:    summary.SimulationID,
			GeneratedAt:     time.Now(),
			SimulationStart: summary.StartTime,
			SimulationEnd:   summary.StartTime.Add(summary.Duration),
			Duration:        summary.Duration.String(),
			Version:         "2.0",
		},
		TeamAnalysis: make(map[string]TeamAnalysis),
	}

	// Generate executive summary
	aar.Summary = g.generateExecutiveSummary(events, summary)

	// Build timeline
	aar.Timeline = g.buildTimeline(events, summary.StartTime)

	// Analyze teams
	aar.TeamAnalysis = g.analyzeTeams(events, summary)

	// Analyze engagements
	aar.Engagements = g.analyzeEngagements(events)

	// Analyze performance
	aar.Performance = g.analyzePerformance(summary)

	// Analyze threats
	aar.ThreatAnalysis = g.analyzeThreatData(events)

	// Analyze system performance
	aar.SystemAnalysis = g.analyzeSystemPerformance(events, summary)

	// Generate event log
	if g.config.DetailLevel == "full" {
		aar.EventLog = g.generateEventLog(events)
	}

	// Generate summary statistics
	aar.Statistics = g.generateStatistics(events, summary)

	// Generate recommendations
	aar.Recommendations = g.generateRecommendations(aar)

	// Generate lessons learned
	aar.Lessons = g.generateLessonsLearned(aar)

	return aar, nil
}

// SaveAAR saves the AAR to file
func (g *AARGenerator) SaveAAR(aar *AAR) error {
	// Create reports directory if it doesn't exist
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("AAR_%s_%s", aar.Metadata.SimulationID[:8], timestamp)

	var err error
	switch g.config.Format {
	case "json":
		err = g.saveJSON(aar, filename)
	case "html":
		err = g.saveHTML(aar, filename)
	case "markdown":
		err = g.saveMarkdown(aar, filename)
	default:
		return fmt.Errorf("unsupported format: %s", g.config.Format)
	}

	if err == nil {
		logger.Successf("AAR saved to: %s", filepath.Join(g.config.OutputDir, filename+"."+g.config.Format))
	}

	return err
}

// saveJSON saves AAR as JSON
func (g *AARGenerator) saveJSON(aar *AAR, filename string) error {
	data, err := json.MarshalIndent(aar, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal AAR: %w", err)
	}

	path := filepath.Join(g.config.OutputDir, filename+".json")
	return os.WriteFile(path, data, 0644)
}

// saveHTML saves AAR as HTML
func (g *AARGenerator) saveHTML(aar *AAR, filename string) error {
	var sb strings.Builder

	// HTML header
	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
	<title>After Action Report</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
		.container { background-color: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		h1 { color: #333; border-bottom: 3px solid #007bff; padding-bottom: 10px; }
		h2 { color: #007bff; margin-top: 30px; }
		h3 { color: #555; }
		.metric { display: inline-block; margin: 10px 20px 10px 0; }
		.metric-label { font-weight: bold; color: #666; }
		.metric-value { font-size: 1.2em; color: #007bff; }
		.team-red { color: #dc3545; }
		.team-blue { color: #007bff; }
		.status-operational { color: #28a745; }
		.status-degraded { color: #ffc107; }
		.status-critical { color: #dc3545; }
		.status-eliminated { color: #6c757d; }
		table { border-collapse: collapse; width: 100%; margin: 20px 0; }
		th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
		th { background-color: #007bff; color: white; }
		.timeline-entry { margin: 10px 0; padding: 10px; background-color: #f8f9fa; border-left: 3px solid #007bff; }
		.priority-high { background-color: #dc3545; color: white; padding: 2px 8px; border-radius: 3px; }
		.priority-medium { background-color: #ffc107; color: black; padding: 2px 8px; border-radius: 3px; }
		.priority-low { background-color: #28a745; color: white; padding: 2px 8px; border-radius: 3px; }
	</style>
</head>
<body>
<div class="container">
`)

	// Title and metadata
	sb.WriteString("<h1>After Action Report</h1>\n")
	sb.WriteString(fmt.Sprintf("<p><strong>Simulation ID:</strong> %s</p>\n", aar.Metadata.SimulationID))
	sb.WriteString(fmt.Sprintf("<p><strong>Generated:</strong> %s</p>\n", aar.Metadata.GeneratedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("<p><strong>Duration:</strong> %s</p>\n", aar.Metadata.Duration))

	// Executive Summary
	sb.WriteString("<h2>Executive Summary</h2>\n")
	sb.WriteString(fmt.Sprintf("<p><strong>Outcome:</strong> %s</p>\n", aar.Summary.Outcome))
	sb.WriteString(fmt.Sprintf("<p><strong>Winner:</strong> <span class='team-%s'>%s</span></p>\n",
		strings.ToLower(aar.Summary.WinningTeam), aar.Summary.WinningTeam))
	sb.WriteString("<div class='metric'><span class='metric-label'>Total Engagements:</span> <span class='metric-value'>" +
		fmt.Sprintf("%d</span></div>\n", aar.Summary.TotalEngagements))
	sb.WriteString("<div class='metric'><span class='metric-label'>Total Losses:</span> <span class='metric-value'>" +
		fmt.Sprintf("%d</span></div>\n", aar.Summary.TotalLosses))

	// Team Analysis
	sb.WriteString("<h2>Team Analysis</h2>\n")
	sb.WriteString("<table>\n")
	sb.WriteString("<tr><th>Team</th><th>Status</th><th>Strength</th><th>Losses</th><th>Kills</th><th>Effectiveness</th></tr>\n")
	for teamName, analysis := range aar.TeamAnalysis {
		statusClass := fmt.Sprintf("status-%s", strings.ToLower(analysis.FinalStatus))
		teamClass := fmt.Sprintf("team-%s", strings.ToLower(teamName))
		sb.WriteString(fmt.Sprintf("<tr><td class='%s'>%s</td>", teamClass, teamName))
		sb.WriteString(fmt.Sprintf("<td class='%s'>%s</td>", statusClass, analysis.FinalStatus))
		sb.WriteString(fmt.Sprintf("<td>%d/%d</td>", analysis.FinalStrength, analysis.InitialStrength))
		sb.WriteString(fmt.Sprintf("<td>%d</td>", analysis.Losses))
		sb.WriteString(fmt.Sprintf("<td>%d</td>", analysis.Kills))
		sb.WriteString(fmt.Sprintf("<td>%.2f</td></tr>\n", analysis.EffectivenessRating))
	}
	sb.WriteString("</table>\n")

	// Recommendations
	if len(aar.Recommendations) > 0 {
		sb.WriteString("<h2>Recommendations</h2>\n")
		for _, rec := range aar.Recommendations {
			priorityClass := fmt.Sprintf("priority-%s", strings.ToLower(rec.Priority))
			sb.WriteString(fmt.Sprintf("<h3>%s <span class='%s'>%s</span></h3>\n", rec.Title, priorityClass, rec.Priority))
			sb.WriteString(fmt.Sprintf("<p>%s</p>\n", rec.Description))
			sb.WriteString(fmt.Sprintf("<p><em>Expected Benefit: %s</em></p>\n", rec.ExpectedBenefit))
		}
	}

	// Close HTML
	sb.WriteString("</div>\n</body>\n</html>\n")

	path := filepath.Join(g.config.OutputDir, filename+".html")
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// saveMarkdown saves AAR as Markdown
func (g *AARGenerator) saveMarkdown(aar *AAR, filename string) error {
	var sb strings.Builder

	// Header
	sb.WriteString("# After Action Report\n\n")
	sb.WriteString(fmt.Sprintf("**Simulation ID:** %s\n", aar.Metadata.SimulationID))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", aar.Metadata.GeneratedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n\n", aar.Metadata.Duration))

	// Executive Summary
	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Outcome:** %s\n\n", aar.Summary.Outcome))
	sb.WriteString(fmt.Sprintf("**Winner:** %s\n\n", aar.Summary.WinningTeam))
	sb.WriteString(fmt.Sprintf("**Total Engagements:** %d\n\n", aar.Summary.TotalEngagements))
	sb.WriteString(fmt.Sprintf("**Total Losses:** %d\n\n", aar.Summary.TotalLosses))

	if len(aar.Summary.KeyEvents) > 0 {
		sb.WriteString("### Key Events\n")
		for _, event := range aar.Summary.KeyEvents {
			sb.WriteString(fmt.Sprintf("- %s\n", event))
		}
		sb.WriteString("\n")
	}

	// Team Analysis
	sb.WriteString("## Team Analysis\n\n")
	for teamName, analysis := range aar.TeamAnalysis {
		sb.WriteString(fmt.Sprintf("### %s\n\n", teamName))
		sb.WriteString(fmt.Sprintf("- **Final Status:** %s\n", analysis.FinalStatus))
		sb.WriteString(fmt.Sprintf("- **Strength:** %d/%d (%.1f%% survival rate)\n",
			analysis.FinalStrength, analysis.InitialStrength,
			float64(analysis.FinalStrength)/float64(analysis.InitialStrength)*100))
		sb.WriteString(fmt.Sprintf("- **Losses:** %d\n", analysis.Losses))
		sb.WriteString(fmt.Sprintf("- **Kills:** %d\n", analysis.Kills))
		sb.WriteString(fmt.Sprintf("- **Effectiveness:** %.2f\n\n", analysis.EffectivenessRating))
	}

	// Engagement Analysis
	sb.WriteString("## Engagement Analysis\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Engagements:** %d\n", aar.Engagements.TotalEngagements))
	sb.WriteString(fmt.Sprintf("- **Successful Hits:** %d (%.1f%% hit rate)\n",
		aar.Engagements.SuccessfulHits, aar.Engagements.HitRate*100))
	sb.WriteString(fmt.Sprintf("- **Average Range:** %.0fm\n\n", aar.Engagements.AverageEngagementRange))

	// Threat Analysis
	if g.config.DetailLevel != "summary" {
		sb.WriteString("## Threat Analysis\n\n")
		sb.WriteString(fmt.Sprintf("- **Threats Identified:** %d\n", aar.ThreatAnalysis.TotalThreatsIdentified))
		sb.WriteString(fmt.Sprintf("- **Threats Neutralized:** %d\n", aar.ThreatAnalysis.ThreatsNeutralized))
		sb.WriteString(fmt.Sprintf("- **Peak Threat Level:** %s\n\n", aar.ThreatAnalysis.PeakThreatLevel))
	}

	// System Performance
	sb.WriteString("## System Performance\n\n")
	sb.WriteString(fmt.Sprintf("- **Average Update Time:** %.2fms\n", aar.Performance.AverageUpdateTime))
	sb.WriteString(fmt.Sprintf("- **Peak Entity Count:** %d\n", aar.Performance.PeakEntityCount))
	sb.WriteString(fmt.Sprintf("- **Simulation Stability:** %.1f%%\n\n", aar.Performance.SimulationStability*100))

	// Recommendations
	if len(aar.Recommendations) > 0 {
		sb.WriteString("## Recommendations\n\n")
		for _, rec := range aar.Recommendations {
			sb.WriteString(fmt.Sprintf("### %s (%s Priority)\n", rec.Title, rec.Priority))
			sb.WriteString(fmt.Sprintf("%s\n\n", rec.Description))
			sb.WriteString(fmt.Sprintf("**Expected Benefit:** %s\n\n", rec.ExpectedBenefit))
		}
	}

	// Lessons Learned
	if len(aar.Lessons) > 0 {
		sb.WriteString("## Lessons Learned\n\n")
		for _, lesson := range aar.Lessons {
			sb.WriteString(fmt.Sprintf("### %s\n", lesson.Category))
			sb.WriteString(fmt.Sprintf("**Observation:** %s\n\n", lesson.Observation))
			sb.WriteString(fmt.Sprintf("**Impact:** %s\n\n", lesson.Impact))
			sb.WriteString(fmt.Sprintf("**Recommendation:** %s\n\n", lesson.Recommendation))
		}
	}

	path := filepath.Join(g.config.OutputDir, filename+".md")
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// generateExecutiveSummary creates the executive summary
func (g *AARGenerator) generateExecutiveSummary(events []SimulationEvent, summary SimulationSummary) ExecutiveSummary {
	exec := ExecutiveSummary{
		TotalEngagements: summary.EventCounts[EventTypeEngagement],
		TotalLosses:      summary.EventCounts[EventTypeDestruction],
		KeyEvents:        make([]string, 0),
	}

	// Determine outcome and winning team
	teamLosses := make(map[string]int)
	for _, event := range events {
		if event.Type == EventTypeDestruction && event.TeamName != "" {
			teamLosses[event.TeamName]++
		}
	}

	// Find team with least losses (simplified)
	minLosses := int(^uint(0) >> 1) // Max int
	for team, losses := range teamLosses {
		if losses < minLosses {
			minLosses = losses
			exec.WinningTeam = team
		}
	}

	if exec.WinningTeam != "" {
		exec.Outcome = fmt.Sprintf("%s achieved tactical superiority", exec.WinningTeam)
	} else {
		exec.Outcome = "Stalemate - no clear victor"
	}

	// Extract key events
	for _, event := range events {
		if event.Type == EventTypeObjective ||
			(event.Type == EventTypeDestruction && len(exec.KeyEvents) < 5) {
			exec.KeyEvents = append(exec.KeyEvents, event.Message)
		}
	}

	return exec
}

// buildTimeline creates a chronological timeline
func (g *AARGenerator) buildTimeline(events []SimulationEvent, startTime time.Time) []TimelineEntry {
	timeline := make([]TimelineEntry, 0)

	// Sort events by timestamp
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	// Convert significant events to timeline entries
	for _, event := range events {
		if g.isSignificantEvent(event) {
			elapsed := event.Timestamp.Sub(startTime)
			entry := TimelineEntry{
				Timestamp:   event.Timestamp,
				ElapsedTime: formatDuration(elapsed),
				EventType:   event.Type,
				Description: event.Message,
				Impact:      g.assessImpact(event),
				Details:     event.Details,
			}
			timeline = append(timeline, entry)
		}
	}

	return timeline
}

// analyzeTeams performs team-specific analysis
func (g *AARGenerator) analyzeTeams(events []SimulationEvent, summary SimulationSummary) map[string]TeamAnalysis {
	teams := make(map[string]TeamAnalysis)

	// Extract team data from events
	for teamName, teamEvents := range summary.TeamEvents {
		analysis := TeamAnalysis{
			TeamName:         teamName,
			Losses:           teamEvents[EventTypeDestruction],
			DronePerformance: make(map[string]DroneStats),
		}

		// Calculate effectiveness rating (simplified)
		if analysis.InitialStrength > 0 {
			analysis.EffectivenessRating = float64(analysis.Kills) / float64(analysis.InitialStrength)
			analysis.FinalStrength = analysis.InitialStrength - analysis.Losses
		}

		// Determine final status
		switch {
		case analysis.FinalStrength == 0:
			analysis.FinalStatus = "Eliminated"
		case float64(analysis.FinalStrength)/float64(analysis.InitialStrength) < 0.3:
			analysis.FinalStatus = "Critical"
		case float64(analysis.FinalStrength)/float64(analysis.InitialStrength) < 0.6:
			analysis.FinalStatus = "Degraded"
		default:
			analysis.FinalStatus = "Operational"
		}

		teams[teamName] = analysis
	}

	return teams
}

// analyzeEngagements performs engagement analysis
func (g *AARGenerator) analyzeEngagements(events []SimulationEvent) EngagementAnalysis {
	analysis := EngagementAnalysis{
		EngagementsByType: make(map[string]int),
		EngagementHeatmap: make([]HeatmapPoint, 0),
	}

	var totalRange float64
	var rangeCount int

	for _, event := range events {
		if event.Type == EventTypeEngagement {
			analysis.TotalEngagements++

			// Extract engagement details
			if details := event.Details; details != nil {
				if hit, ok := details["hit"].(bool); ok && hit {
					analysis.SuccessfulHits++
				}

				if engType, ok := details["type"].(string); ok {
					analysis.EngagementsByType[engType]++
				}

				if distance, ok := details["distance"].(float64); ok {
					totalRange += distance
					rangeCount++
				}
			}
		}
	}

	// Calculate averages
	if analysis.TotalEngagements > 0 {
		analysis.HitRate = float64(analysis.SuccessfulHits) / float64(analysis.TotalEngagements)
	}

	if rangeCount > 0 {
		analysis.AverageEngagementRange = totalRange / float64(rangeCount)
	}

	return analysis
}

// analyzePerformance analyzes system performance
func (g *AARGenerator) analyzePerformance(summary SimulationSummary) PerformanceAnalysis {
	analysis := PerformanceAnalysis{
		ResourceUtilization: make(map[string]float64),
	}

	// Extract performance metrics from summary
	if metric, ok := summary.Metrics["update_time"]; ok && len(metric.History) > 0 {
		var total float64
		for _, point := range metric.History {
			total += point.Value
		}
		analysis.AverageUpdateTime = total / float64(len(metric.History))
	}

	if metric, ok := summary.Metrics["entity_count"]; ok {
		analysis.PeakEntityCount = int(metric.Value)
	}

	// Calculate stability (simplified - based on error rate)
	errorCount := summary.EventCounts[EventTypeSystem]
	totalEvents := summary.TotalEvents
	if totalEvents > 0 {
		errorRate := float64(errorCount) / float64(totalEvents)
		analysis.SimulationStability = 1.0 - errorRate
	}

	return analysis
}

// Helper functions

func (g *AARGenerator) isSignificantEvent(event SimulationEvent) bool {
	return event.Type == EventTypeEngagement ||
		event.Type == EventTypeDestruction ||
		event.Type == EventTypeObjective ||
		(event.Type == EventTypeTeamStatus && event.Severity != SeverityInfo)
}

func (g *AARGenerator) assessImpact(event SimulationEvent) string {
	switch event.Type {
	case EventTypeDestruction:
		return "High - Force reduction"
	case EventTypeObjective:
		return "High - Mission progress"
	case EventTypeEngagement:
		if hit, ok := event.Details["hit"].(bool); ok && hit {
			return "Medium - Successful engagement"
		}
		return "Low - Missed engagement"
	default:
		return "Low"
	}
}

// analyzeThreatData analyzes threat-related events
func (g *AARGenerator) analyzeThreatData(events []SimulationEvent) ThreatAnalysis {
	analysis := ThreatAnalysis{
		ThreatsByType:  make(map[string]int),
		ThreatTimeline: make([]ThreatEvent, 0),
	}

	var threatDurations []time.Duration
	var lastThreatTime time.Time
	maxThreatLevel := "LOW"

	for _, event := range events {
		if event.Type == EventTypeThreat {
			analysis.TotalThreatsIdentified++

			if details := event.Details; details != nil {
				if level, ok := details["threat_level"].(string); ok {
					if g.compareThreatLevels(level, maxThreatLevel) > 0 {
						maxThreatLevel = level
					}

					threatEvent := ThreatEvent{
						Timestamp:   event.Timestamp,
						ThreatLevel: level,
						Team:        event.TeamName,
					}

					if count, ok := details["threat_count"].(int); ok {
						threatEvent.ThreatCount = count
					}

					analysis.ThreatTimeline = append(analysis.ThreatTimeline, threatEvent)
				}
			}

			if !lastThreatTime.IsZero() {
				threatDurations = append(threatDurations, event.Timestamp.Sub(lastThreatTime))
			}
			lastThreatTime = event.Timestamp
		}

		if event.Type == EventTypeDestruction {
			if details := event.Details; details != nil {
				if cause, ok := details["cause"].(string); ok && strings.Contains(cause, "intercepted") {
					analysis.ThreatsNeutralized++
				}
			}
		}
	}

	analysis.PeakThreatLevel = maxThreatLevel

	// Calculate average threat duration
	if len(threatDurations) > 0 {
		var totalDuration time.Duration
		for _, d := range threatDurations {
			totalDuration += d
		}
		avgDuration := totalDuration / time.Duration(len(threatDurations))
		analysis.AverageThreatDuration = formatDuration(avgDuration)
	}

	return analysis
}

// analyzeSystemPerformance analyzes system-level performance
func (g *AARGenerator) analyzeSystemPerformance(events []SimulationEvent, summary SimulationSummary) SystemAnalysis {
	analysis := SystemAnalysis{
		SystemFailures: make([]SystemFailure, 0),
	}

	// Calculate communication reliability (based on successful commands)
	var totalCommands, successfulCommands int
	for _, event := range events {
		if event.Type == EventTypeCommand {
			totalCommands++
			if details := event.Details; details != nil {
				if success, ok := details["success"].(bool); ok && success {
					successfulCommands++
				}
			}
		}
	}

	if totalCommands > 0 {
		analysis.CommunicationReliability = float64(successfulCommands) / float64(totalCommands)
	} else {
		analysis.CommunicationReliability = 1.0
	}

	// Calculate sensor accuracy (based on detections)
	var totalDetections, accurateDetections int
	for _, event := range events {
		if event.Type == EventTypeDetection {
			totalDetections++
			if details := event.Details; details != nil {
				if accurate, ok := details["accurate"].(bool); !ok || accurate {
					accurateDetections++
				}
			}
		}
	}

	if totalDetections > 0 {
		analysis.SensorAccuracy = float64(accurateDetections) / float64(totalDetections)
	} else {
		analysis.SensorAccuracy = 1.0
	}

	// Calculate weapon system efficiency
	if totalEngagements := summary.EventCounts[EventTypeEngagement]; totalEngagements > 0 {
		successfulHits := 0
		for _, event := range events {
			if event.Type == EventTypeEngagement {
				if details := event.Details; details != nil {
					if hit, ok := details["hit"].(bool); ok && hit {
						successfulHits++
					}
				}
			}
		}
		analysis.WeaponSystemEfficiency = float64(successfulHits) / float64(totalEngagements)
	} else {
		analysis.WeaponSystemEfficiency = 0.0
	}

	// Extract system failures
	for _, event := range events {
		if event.Type == EventTypeSystem && event.Severity == SeverityError {
			failure := SystemFailure{
				Timestamp:   event.Timestamp,
				Description: event.Message,
			}

			if details := event.Details; details != nil {
				if system, ok := details["system"].(string); ok {
					failure.System = system
				}
				if impact, ok := details["impact"].(string); ok {
					failure.Impact = impact
				}
			}

			analysis.SystemFailures = append(analysis.SystemFailures, failure)
		}
	}

	// Calculate command latency from metrics
	if metric, ok := summary.Metrics["command_latency"]; ok {
		analysis.CommandLatency = metric.Value
	}

	// Calculate autonomy performance (simplified)
	analysis.AutonomyPerformance = (analysis.CommunicationReliability + analysis.SensorAccuracy +
		analysis.WeaponSystemEfficiency) / 3.0

	return analysis
}

// generateEventLog creates a detailed event log
func (g *AARGenerator) generateEventLog(events []SimulationEvent) []EventLogEntry {
	log := make([]EventLogEntry, 0, len(events))

	for _, event := range events {
		entry := EventLogEntry{
			Timestamp:   event.Timestamp,
			EventType:   event.Type,
			Severity:    event.Severity,
			Description: event.Message,
			Team:        event.TeamName,
			Details:     event.Details,
		}

		if event.EntityID != nil {
			entry.Entity = event.EntityID.String()[:8]
		}

		log = append(log, entry)
	}

	return log
}

// generateStatistics generates summary statistics
func (g *AARGenerator) generateStatistics(events []SimulationEvent, summary SimulationSummary) SummaryStatistics {
	stats := SummaryStatistics{
		TotalEngagements:    summary.EventCounts[EventTypeEngagement],
		ResourceUtilization: make(map[string]float64),
	}

	// Count drones and losses
	dronesDeployed := make(map[uuid.UUID]bool)
	dronesLost := make(map[uuid.UUID]bool)
	var maxConcurrent int

	for _, event := range events {
		if event.Type == EventTypeSpawn && event.EntityID != nil {
			dronesDeployed[*event.EntityID] = true
		}
		if event.Type == EventTypeDestruction && event.EntityID != nil {
			dronesLost[*event.EntityID] = true
		}
		if event.Type == EventTypeInterception {
			if details := event.Details; details != nil {
				if success, ok := details["success"].(bool); ok && success {
					stats.SuccessfulInterceptions++
				}
			}
		}
	}

	stats.TotalDronesDeployed = len(dronesDeployed)
	stats.TotalDronesLost = len(dronesLost)

	// Extract metrics
	if metric, ok := summary.Metrics["entity_count"]; ok {
		stats.PeakConcurrentDrones = int(metric.Value)
		for _, point := range metric.History {
			if int(point.Value) > maxConcurrent {
				maxConcurrent = int(point.Value)
			}
		}
		if maxConcurrent > stats.PeakConcurrentDrones {
			stats.PeakConcurrentDrones = maxConcurrent
		}
	}

	stats.AverageMissionDuration = summary.Duration.String()

	// Calculate resource utilization
	if metric, ok := summary.Metrics["cpu_usage"]; ok {
		stats.ResourceUtilization["cpu"] = metric.Value
	}
	if metric, ok := summary.Metrics["memory_usage"]; ok {
		stats.ResourceUtilization["memory"] = metric.Value
	}

	return stats
}

// generateRecommendations generates improvement recommendations
func (g *AARGenerator) generateRecommendations(aar *AAR) []Recommendation {
	recs := make([]Recommendation, 0)

	// Check engagement effectiveness
	if aar.Engagements.HitRate < 0.3 {
		recs = append(recs, Recommendation{
			Priority:        "High",
			Category:        "Targeting",
			Title:           "Improve Targeting Algorithms",
			Description:     "Current hit rate is below 30%, indicating significant issues with targeting accuracy.",
			ExpectedBenefit: "Increase hit rate to 50-60%, reducing ammunition expenditure and increasing mission effectiveness.",
		})
	}

	// Check communication reliability
	if aar.SystemAnalysis.CommunicationReliability < 0.95 {
		recs = append(recs, Recommendation{
			Priority:        "High",
			Category:        "Communications",
			Title:           "Enhance Communication Protocols",
			Description:     "Communication reliability is below 95%, risking command and control effectiveness.",
			ExpectedBenefit: "Improve mission coordination and reduce response times.",
		})
	}

	// Check threat response
	if aar.ThreatAnalysis.TotalThreatsIdentified > 0 {
		neutralizationRate := float64(aar.ThreatAnalysis.ThreatsNeutralized) / float64(aar.ThreatAnalysis.TotalThreatsIdentified)
		if neutralizationRate < 0.8 {
			recs = append(recs, Recommendation{
				Priority:        "Medium",
				Category:        "Threat Response",
				Title:           "Improve Threat Neutralization Capability",
				Description:     fmt.Sprintf("Only %.1f%% of identified threats were neutralized.", neutralizationRate*100),
				ExpectedBenefit: "Increase survivability and mission success rate.",
			})
		}
	}

	// Check system stability
	if aar.Performance.SimulationStability < 0.98 {
		recs = append(recs, Recommendation{
			Priority:        "Medium",
			Category:        "System Stability",
			Title:           "Address System Stability Issues",
			Description:     "System stability is below optimal levels, indicating potential reliability issues.",
			ExpectedBenefit: "Reduce system failures and improve overall mission reliability.",
		})
	}

	// Check resource utilization
	if aar.Statistics.ResourceUtilization["cpu"] > 0.8 || aar.Statistics.ResourceUtilization["memory"] > 0.8 {
		recs = append(recs, Recommendation{
			Priority:        "Low",
			Category:        "Performance",
			Title:           "Optimize Resource Usage",
			Description:     "High resource utilization detected, which may limit scalability.",
			ExpectedBenefit: "Enable handling of larger swarm sizes and more complex scenarios.",
		})
	}

	return recs
}

// generateLessonsLearned extracts insights from the simulation
func (g *AARGenerator) generateLessonsLearned(aar *AAR) []LessonLearned {
	lessons := make([]LessonLearned, 0)

	// Analyze engagement effectiveness
	if aar.Engagements.HitRate < 0.3 {
		lessons = append(lessons, LessonLearned{
			Category:       "Engagement",
			Observation:    "Low hit rate observed across all engagements",
			Impact:         "Reduced combat effectiveness and increased ammunition expenditure",
			Recommendation: "Review and optimize targeting algorithms, consider environmental factors",
		})
	} else if aar.Engagements.HitRate > 0.7 {
		lessons = append(lessons, LessonLearned{
			Category:       "Engagement",
			Observation:    "High hit rate demonstrates effective targeting systems",
			Impact:         "Efficient resource utilization and high combat effectiveness",
			Recommendation: "Document current targeting parameters for future reference",
		})
	}

	// Analyze team performance disparities
	var maxEffectiveness, minEffectiveness float64 = 0, 999
	var bestTeam, worstTeam string

	for teamName, analysis := range aar.TeamAnalysis {
		if analysis.EffectivenessRating > maxEffectiveness {
			maxEffectiveness = analysis.EffectivenessRating
			bestTeam = teamName
		}
		if analysis.EffectivenessRating < minEffectiveness {
			minEffectiveness = analysis.EffectivenessRating
			worstTeam = teamName
		}
	}

	if maxEffectiveness-minEffectiveness > 0.5 {
		lessons = append(lessons, LessonLearned{
			Category: "Team Balance",
			Observation: fmt.Sprintf("Significant performance disparity between %s (%.2f) and %s (%.2f)",
				bestTeam, maxEffectiveness, worstTeam, minEffectiveness),
			Impact:         "Unbalanced engagement leading to predictable outcomes",
			Recommendation: "Review team compositions and adjust force ratios or capabilities",
		})
	}

	// Analyze threat response effectiveness
	if aar.ThreatAnalysis.TotalThreatsIdentified > 0 {
		neutralizationRate := float64(aar.ThreatAnalysis.ThreatsNeutralized) / float64(aar.ThreatAnalysis.TotalThreatsIdentified)
		if neutralizationRate < 0.6 {
			lessons = append(lessons, LessonLearned{
				Category:       "Threat Response",
				Observation:    fmt.Sprintf("Low threat neutralization rate (%.1f%%)", neutralizationRate*100),
				Impact:         "Increased vulnerability to enemy attacks",
				Recommendation: "Enhance threat detection and interception capabilities",
			})
		}
	}

	// Analyze system reliability
	if len(aar.SystemAnalysis.SystemFailures) > 0 {
		lessons = append(lessons, LessonLearned{
			Category:       "System Reliability",
			Observation:    fmt.Sprintf("%d system failures occurred during simulation", len(aar.SystemAnalysis.SystemFailures)),
			Impact:         "Potential mission degradation and increased operator workload",
			Recommendation: "Implement redundancy and improve error recovery mechanisms",
		})
	}

	// Analyze communication performance
	if aar.SystemAnalysis.CommunicationReliability < 0.95 {
		lessons = append(lessons, LessonLearned{
			Category: "Communications",
			Observation: fmt.Sprintf("Communication reliability at %.1f%% (below 95%% threshold)",
				aar.SystemAnalysis.CommunicationReliability*100),
			Impact:         "Delayed command execution and potential loss of coordination",
			Recommendation: "Strengthen communication protocols and consider mesh networking",
		})
	}

	// Analyze resource utilization
	if aar.Statistics.TotalDronesLost > 0 {
		lossRate := float64(aar.Statistics.TotalDronesLost) / float64(aar.Statistics.TotalDronesDeployed)
		if lossRate > 0.5 {
			lessons = append(lessons, LessonLearned{
				Category:       "Force Preservation",
				Observation:    fmt.Sprintf("High attrition rate (%.1f%% losses)", lossRate*100),
				Impact:         "Unsustainable resource consumption",
				Recommendation: "Review defensive tactics and consider more conservative engagement rules",
			})
		}
	}

	// Analyze peak threat levels
	if aar.ThreatAnalysis.PeakThreatLevel == "CRITICAL" {
		lessons = append(lessons, LessonLearned{
			Category:       "Threat Management",
			Observation:    "Threat level reached CRITICAL during simulation",
			Impact:         "System overwhelmed by simultaneous threats",
			Recommendation: "Develop surge capacity and prioritization algorithms for high-threat scenarios",
		})
	}

	return lessons
}

// compareThreatLevels compares two threat levels
func (g *AARGenerator) compareThreatLevels(a, b string) int {
	levels := map[string]int{
		"LOW":      1,
		"MEDIUM":   2,
		"HIGH":     3,
		"CRITICAL": 4,
	}

	return levels[a] - levels[b]
}

func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
