package simulation

import (
	"context"

	"github.com/picogrid/legion-simulations/pkg/client"
)

// Simulation defines the interface that all simulations must implement
type Simulation interface {
	// Name returns the name of the simulation
	Name() string

	// Description returns a brief description of what the simulation does
	Description() string

	// Configure sets up the simulation with the provided parameters
	Configure(params map[string]interface{}) error

	// Run executes the simulation using the provided Legion client
	Run(ctx context.Context, client *client.Legion) error

	// Stop gracefully shuts down the simulation
	Stop() error
}
