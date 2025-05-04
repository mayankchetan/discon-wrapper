#!/bin/bash
# Script to build multiple ROSCO controller versions

# Set default versions to build if none provided
DEFAULT_VERSIONS=("main" "develop" "v2.9.7" "v2.9.6" "v2.9.5" "v2.8.1" "v2.9.4")

# Get versions from command line or use defaults
if [ $# -gt 0 ]; then
    VERSIONS=("$@")
else
    VERSIONS=("${DEFAULT_VERSIONS[@]}")
fi

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}==> Building ROSCO Controller images for versions: ${YELLOW}${VERSIONS[*]}${NC}"

# Path to controllers database
DB_PATH="./discon-manager/db/controllers.json"
TEMP_DB_PATH="/tmp/controllers_temp.json"

# Function to check if Docker is available
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is not installed or not in PATH"
        exit 1
    fi
}

# Function to build a single ROSCO version
build_rosco_version() {
    local version=$1
    local tag_suffix=$2
    
    echo -e "${GREEN}==> Building ROSCO version ${YELLOW}${version}${NC}"
    
    # Build the Docker image with the specific version
    docker build -f docker/Dockerfile.rosco \
        --build-arg ROSCO_VERSION="${version}" \
        -t "discon-server-rosco:${tag_suffix}" .
    
    if [ $? -ne 0 ]; then
        echo "Failed to build ROSCO version ${version}"
        return 1
    fi
    
    echo -e "${GREEN}✓ Successfully built ROSCO version ${YELLOW}${version}${NC}"
    return 0
}

# Function to add a controller to the database
add_controller_to_db() {
    local version=$1
    local tag_suffix=$2
    local id="rosco-${tag_suffix}"
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        echo "Warning: jq not found, skipping database update for version ${version}"
        echo "Please install jq and update the database manually"
        return 1
    fi
    
    echo "Updating controllers database for ROSCO ${version}..."
    
    # Create a new controller entry
    local now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local new_controller=$(cat << EOF
{
  "id": "${id}",
  "name": "ROSCO Controller ${version}",
  "version": "${version}",
  "image": "discon-server-rosco:${tag_suffix}",
  "description": "ROSCO controller implementation version ${version}",
  "library_path": "/app/build/libdiscon.so",
  "proc_name": "DISCON",
  "ports": {
    "internal": 8080,
    "external": 0
  },
  "created_at": "${now}",
  "updated_at": "${now}"
}
EOF
)
    
    # Check if the controller already exists in the database
    local controller_exists=$(jq --arg id "${id}" '.controllers | map(select(.id == $id)) | length' "$DB_PATH")
    
    if [ "$controller_exists" -gt 0 ]; then
        echo "Controller ${id} already exists in database, updating..."
        jq --argjson controller "$new_controller" --arg id "${id}" '
        .controllers = .controllers | map(if .id == $id then $controller else . end)
        ' "$DB_PATH" > "$TEMP_DB_PATH"
    else
        echo "Adding new controller ${id} to database..."
        jq --argjson controller "$new_controller" '
        .controllers += [$controller]
        ' "$DB_PATH" > "$TEMP_DB_PATH"
    fi
    
    # Replace the original file if temp file was created successfully
    if [ -s "$TEMP_DB_PATH" ]; then
        cp "$TEMP_DB_PATH" "$DB_PATH"
        echo -e "${GREEN}✓ Successfully updated database for ROSCO ${YELLOW}${version}${NC}"
    else
        echo "Error: Failed to update database for ROSCO ${version}"
        return 1
    fi
}

# Main execution
check_docker

# Build each version
for version in "${VERSIONS[@]}"; do
    # Create tag-friendly version string (replace dots with underscores)
    tag_suffix=$(echo "${version}" | tr '.' '_')
    
    # Build the Docker image
    if build_rosco_version "$version" "$tag_suffix"; then
        # Add the controller to the database
        add_controller_to_db "$version" "$tag_suffix"
    fi
done

echo -e "${GREEN}==> All ROSCO versions built successfully!${NC}"
echo -e "You can now restart the discon-manager to use the new controllers."
echo -e "Run: ${YELLOW}sudo docker-compose down && sudo docker-compose up${NC}"