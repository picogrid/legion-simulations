package simulation

// SimulationConfig represents the configuration structure for a simulation
// loaded from simulation.yaml
type SimulationConfig struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Version     string      `yaml:"version"`
	Category    string      `yaml:"category"`
	Parameters  []Parameter `yaml:"parameters"`
}

// Parameter defines a configurable parameter for a simulation
type Parameter struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"` // integer, float, string, duration, boolean
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
	Required    bool        `yaml:"required"`
	Min         interface{} `yaml:"min,omitempty"`
	Max         interface{} `yaml:"max,omitempty"`
	Options     []string    `yaml:"options,omitempty"` // For string enums
}
