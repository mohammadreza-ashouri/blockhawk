#!/bin/bash

# Clear terminal
clear

# Print header
echo "======================================================"
echo "  BlockHawk Security Monitor - Setup"
echo "======================================================"
echo ""

# Create necessary directories
echo "Creating project directories..."
mkdir -p data
mkdir -p logs

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}')
echo "Using $GO_VERSION"

# Install dependencies
echo "Installing dependencies..."
go get github.com/mattn/go-sqlite3
go mod tidy

# Check for errors
if [ $? -ne 0 ]; then
    echo "Error installing dependencies. Please check for errors."
    exit 1
fi

# Make scripts executable
chmod +x run.sh

echo ""
echo "Setup complete! You can now run the application with ./run.sh"
echo ""