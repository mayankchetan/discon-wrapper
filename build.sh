#!/bin/bash

# Build script for discon-manager and controller images

set -e

# Define the path to the Go executable
GO_PATH="/usr/bin/go/bin/go"

echo "Building discon-manager system..."

# Step 0: Create build directory if it doesn't exist
mkdir -p build

# Step 1: Build the client DLL
echo "==> Step 1/4: Building discon-client DLL"
$GO_PATH build -buildmode=c-shared -o build/discon-client.dll discon-wrapper/discon-client
echo "✓ Done"

# Step 2: Build the base discon-server image
echo "==> Step 2/4: Building discon-server base image"
docker build -f docker/Dockerfile.server -t discon-server:latest .
echo "✓ Done"

# Step 3: Build the ROSCO controller image
echo "==> Step 3/4: Building ROSCO controller image"
docker build -f docker/Dockerfile.rosco -t discon-server-rosco:latest .
echo "✓ Done"

# Step 4: Build the discon-manager image
echo "==> Step 4/4: Building discon-manager image"
docker-compose build
echo "✓ Done"

echo "All builds completed successfully!"
echo "Client DLL is available at: $(pwd)/build/discon-client.dll"
echo ""
echo "To start the discon-manager system, run:"
echo "  docker-compose up -d"
echo ""
echo "To view logs:"
echo "  docker-compose logs -f"
echo ""
echo "To check status:"
echo "  docker-compose ps"
echo ""
echo "To stop the system:"
echo "  docker-compose down"
echo ""
echo "Access the server at http://localhost:8080"
echo "Management endpoints:"
echo "  /health - Health check"
echo "  /metrics - Basic metrics"
echo "  /containers - List running containers"
echo "  /controllers - List available controllers"
echo ""
echo "Client connections should connect to: ws://localhost:8080/ws with query parameters:"
echo "  controller=ID or version=VERSION - Controller to use"
echo "  path (optional) - Custom controller library path"
echo "  proc (optional) - Custom proc name"
echo ""