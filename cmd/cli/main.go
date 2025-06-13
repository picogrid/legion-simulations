package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/picogrid/legion-simulations/cmd/cli/cmd"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	if err := cmd.Execute(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
