#!/bin/bash
# Run different drone swarm simulation scenarios

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR/../../.."

echo "======================================"
echo "Drone Swarm Simulation Examples"
echo "======================================"
echo ""
echo "Available scenarios:"
echo "1. Default (Interactive) - 50 threats vs 10 defenders"
echo "2. Large Scale Battle - 100 threats vs 20 defenders"
echo "3. Defensive Test - 15 threats vs 10 defenders"
echo "4. Custom Parameters"
echo ""
read -p "Select scenario (1-4): " choice

case $choice in
  1)
    echo "Running default interactive simulation..."
    ./bin/legion-sim run -s "Drone Swarm Combat"
    ;;
  2)
    echo "Running large scale battle..."
    ./bin/legion-sim run -s "Drone Swarm Combat" -p cmd/drone-swarm/examples/large-scale-battle.yaml
    ;;
  3)
    echo "Running defensive test..."
    ./bin/legion-sim run -s "Drone Swarm Combat" -p cmd/drone-swarm/examples/defensive-test.yaml
    ;;
  4)
    echo "Enter custom parameters..."
    read -p "Number of Counter-UAS systems: " counter_uas
    read -p "Number of UAS threats: " threats
    read -p "Number of waves: " waves
    read -p "Duration (e.g., 2m, 5m): " duration
    
    export LEGION_NUM_COUNTER_UAS_SYSTEMS=$counter_uas
    export LEGION_NUM_UAS_THREATS=$threats
    export LEGION_WAVES=$waves
    export LEGION_DURATION=$duration
    
    ./bin/legion-sim run -s "Drone Swarm Combat"
    ;;
  *)
    echo "Invalid choice"
    exit 1
    ;;
esac