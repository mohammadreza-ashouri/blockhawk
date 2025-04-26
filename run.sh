#!/bin/bash

# Clear terminal
clear

# Print header
echo "======================================================"
echo "  BlockHawk Security Monitor - SQLite Edition"
echo "======================================================"
echo ""

# Create necessary directories
mkdir -p data
mkdir -p logs

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Build the application
echo "Building BlockHawk..."
go build -o blockhawk cmd/server/main.go

# Check if build was successful
if [ $? -ne 0 ]; then
    echo "Build failed. Please check for errors."
    exit 1
fi

# Run the application
echo "Starting BlockHawk (press Ctrl+C to stop)..."
echo "Access the dashboard at http://localhost:8080"
echo ""
./blockhawk