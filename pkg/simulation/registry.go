package simulation

import (
	"fmt"
	"sync"
)

// Registry manages available simulations
type Registry struct {
	mu          sync.RWMutex
	simulations map[string]func() Simulation
}

// NewRegistry creates a new simulation registry
func NewRegistry() *Registry {
	return &Registry{
		simulations: make(map[string]func() Simulation),
	}
}

// Register adds a simulation to the registry
func (r *Registry) Register(name string, factory func() Simulation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.simulations[name]; exists {
		return fmt.Errorf("simulation %s already registered", name)
	}

	r.simulations[name] = factory
	return nil
}

// Get returns a new instance of the requested simulation
func (r *Registry) Get(name string) (Simulation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.simulations[name]
	if !exists {
		return nil, fmt.Errorf("simulation %s not found", name)
	}

	return factory(), nil
}

// List returns all registered simulation names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.simulations))
	for name := range r.simulations {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry is the global simulation registry
var DefaultRegistry = NewRegistry()
