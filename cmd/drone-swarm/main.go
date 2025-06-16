package main

import (
	"fmt"
	"os"

	// Import to register the simulation
	_ "github.com/picogrid/legion-simulations/cmd/drone-swarm/simulation"
)

func main() {
	fmt.Println("Drone Swarm simulation registered. Use 'legion-sim run' to execute.")
	os.Exit(0)
}
