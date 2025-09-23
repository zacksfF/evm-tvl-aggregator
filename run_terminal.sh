#!/bin/bash

# EVM TVL Aggregator Terminal Demo
# This script demonstrates the terminal interface

set -e

echo "EVM TVL Aggregator Terminal Interface"
echo "====================================="
echo
echo "Status: Starting services..."
echo

# Check if API is already running
if curl -s http://localhost:8080/api/v1/health > /dev/null 2>&1; then
    echo "API server already running at http://localhost:8080"
else
    echo "Starting API server..."
    go run cmd/api/main.go &
    API_PID=$!
    
    # Wait for API to start
    echo "Waiting for API to be ready..."
    for i in {1..15}; do
        if curl -s http://localhost:8080/api/v1/health > /dev/null 2>&1; then
            echo "API server ready"
            break
        fi
        if [ $i -eq 15 ]; then
            echo "Error: API server failed to start"
            kill $API_PID 2>/dev/null || true
            exit 1
        fi
        sleep 1
    done
fi

echo
echo "API Health Check:"
curl -s http://localhost:8080/api/v1/health | head -c 100
echo
echo

echo "Terminal Interface Controls:"
echo "  Up/Down/Left/Right: Navigate"
echo "  Space: Pause/Resume"
echo "  /: Search"
echo "  r: Refresh"
echo "  ?: Help"
echo "  q: Quit"
echo

echo "Starting Terminal Interface..."
echo "(Press 'q' to quit)"
echo

# Start TUI
go run cmd/tui/main.go

# Cleanup if we started the API
if [ ! -z "$API_PID" ]; then
    echo
    echo "Stopping API server..."
    kill $API_PID 2>/dev/null || true
fi

echo "Demo completed"