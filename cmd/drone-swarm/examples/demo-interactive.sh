#!/bin/bash

echo "======================================"
echo "Legion Simulations - Interactive Demo"
echo "======================================"
echo ""
echo "This demo shows how the CLI prompts work:"
echo "1. Without environment variables - uses defaults from simulation.yaml"
echo "2. With environment variables - uses them as defaults in prompts"
echo ""
echo "Press Enter to continue..."
read

echo ""
echo "=== Demo 1: No environment variables set ==="
echo "Watch how it shows the defaults from simulation.yaml..."
echo ""
echo "Simulating user pressing Enter to accept all defaults:"
echo ""

# Use yes with empty lines to simulate pressing Enter
yes '' | head -9 | ../../../bin/legion-sim run -s "Drone Swarm Combat" 2>&1 | grep -E "(Number of|attack waves|Simulation update|Maximum simulation|Clean up)" || true

echo ""
echo "=== Demo 2: With environment variables ==="
echo "Setting LEGION_NUM_UAS_THREATS=100..."
export LEGION_NUM_UAS_THREATS=100
echo ""
echo "Now the prompt will show 100 as the default for threats:"
echo ""

yes '' | head -9 | ../../../bin/legion-sim run -s "Drone Swarm Combat" 2>&1 | grep -E "(Number of|attack waves|Simulation update|Maximum simulation|Clean up)" || true

echo ""
echo "=== Demo 3: Skip all prompts (automation mode) ==="
echo "Setting LEGION_SKIP_PROMPTS=true..."
export LEGION_SKIP_PROMPTS=true
echo ""
echo "Now it will run without any prompts:"
echo ""

../../../bin/legion-sim run -s "Drone Swarm Combat" 2>&1 | head -20