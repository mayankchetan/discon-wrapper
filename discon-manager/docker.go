package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// DockerController manages Docker containers
type DockerController struct {
	client         *client.Client
	config         DockerConfig
	containers     map[string]*ContainerInfo
	networkID      string
	mutex          sync.RWMutex
	containerCount int
	ctx            context.Context
	discovery      *ControllerDiscoveryConfig
}

// ContainerInfo represents information about a running container
type ContainerInfo struct {
	ID          string
	Name        string
	Image       string
	CreatedAt   time.Time
	ContainerIP string
	Hostname    string
	Host        string
	Port        int
	ClientID    string
	Controller  *Controller
}

// DiscoveredController represents a controller discovered from Docker images
type DiscoveredController struct {
	ID           string
	Name         string
	Version      string
	Image        string
	Description  string
	LibraryPath  string
	ProcName     string
	Ports        ControllerPorts
	CreatedAt    string
	IsValid      bool
	ValidateInfo string
}

// ControllerPorts represents port configuration for a controller
type ControllerPorts struct {
	Internal int `json:"internal"`
	External int `json:"external"`
}

// NewDockerController creates a new Docker controller
func NewDockerController(ctx context.Context, config DockerConfig, discovery *ControllerDiscoveryConfig) (*DockerController, error) {
	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}

	controller := &DockerController{
		client:     cli,
		config:     config,
		containers: make(map[string]*ContainerInfo),
		ctx:        ctx,
		discovery:  discovery,
	}

	// Ensure the Docker network exists
	networkID, err := controller.ensureNetwork(ctx)
	if err != nil {
		return nil, err
	}
	controller.networkID = networkID

	return controller, nil
}

// ensureNetwork ensures that the Docker network exists
func (dc *DockerController) ensureNetwork(ctx context.Context) (string, error) {
	// Check if the network already exists
	filters := filters.NewArgs()
	filters.Add("name", dc.config.NetworkName)
	networks, err := dc.client.NetworkList(ctx, types.NetworkListOptions{
		Filters: filters,
	})
	if err != nil {
		return "", fmt.Errorf("error listing networks: %w", err)
	}

	// If the network exists, return its ID
	for _, n := range networks {
		if n.Name == dc.config.NetworkName {
			log.Printf("Found existing network: %s (%s)", n.Name, n.ID)
			return n.ID, nil
		}
	}

	// Create the network
	resp, err := dc.client.NetworkCreate(ctx, dc.config.NetworkName, types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
		Attachable:     true,
	})
	if err != nil {
		return "", fmt.Errorf("error creating network: %w", err)
	}

	log.Printf("Created network: %s (%s)", dc.config.NetworkName, resp.ID)
	return resp.ID, nil
}

// ensureContainerDoesNotExist checks if a container with the given name already exists and removes it
func (dc *DockerController) ensureContainerDoesNotExist(ctx context.Context, containerName string) error {
	// Check if the container already exists by name
	containers, err := dc.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return fmt.Errorf("error listing containers: %w", err)
	}

	// Check for existing container with the same name
	for _, c := range containers {
		for _, name := range c.Names {
			// Container names from API have a leading slash that needs to be removed
			if strings.TrimPrefix(name, "/") == containerName {
				log.Printf("Found existing container with name %s (ID: %s), removing it", containerName, c.ID[:12])

				// Stop the container if it's running
				if c.State == "running" {
					timeoutSeconds := 10
					if err := dc.client.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
						log.Printf("Warning: error stopping container %s: %v", c.ID[:12], err)
						// Continue trying to remove anyway
					}
				}

				// Remove the container
				if err := dc.client.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
					Force:         true,
					RemoveVolumes: true,
				}); err != nil {
					return fmt.Errorf("error removing existing container %s: %w", c.ID[:12], err)
				}

				log.Printf("Successfully removed existing container %s", containerName)
				break
			}
		}
	}

	return nil
}

// StartContainer starts a container for the given controller configuration
func (dc *DockerController) StartContainer(ctx context.Context, controller *Controller, clientID string) (*ContainerInfo, error) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	dc.containerCount++

	// Generate a unique name for the container
	name := fmt.Sprintf("%s%s-%d", dc.config.ContainerPrefix, controller.ID, dc.containerCount)

	// Ensure no container with the same name already exists
	if err := dc.ensureContainerDoesNotExist(ctx, name); err != nil {
		return nil, err
	}

	// Create container configuration
	config := &container.Config{
		Image:    controller.Image,
		Hostname: name,
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", controller.Ports.Internal)): struct{}{},
		},
		Env: []string{
			"DEBUG_LEVEL=1",
		},
	}

	// Create host configuration with resource limits
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   ParseMemoryLimit(dc.config.MemoryLimit),
			NanoCPUs: int64(dc.config.CPULimit * 1e9), // Convert CPUs to nano-CPUs
		},
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", controller.Ports.Internal)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprintf("%d", controller.Ports.External),
				},
			},
		},
	}

	// Set up network configuration
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			dc.config.NetworkName: {
				NetworkID: dc.networkID,
			},
		},
	}

	// Create the container
	resp, err := dc.client.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, name)
	if err != nil {
		return nil, fmt.Errorf("error creating container: %w", err)
	}

	// Start the container
	if err := dc.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("error starting container: %w", err)
	}

	// Get container information
	containerInfo, err := dc.client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("error inspecting container: %w", err)
	}

	// Get the container IP address
	containerIP := containerInfo.NetworkSettings.Networks[dc.config.NetworkName].IPAddress

	// Create container info
	info := &ContainerInfo{
		ID:          resp.ID,
		Name:        name,
		Image:       controller.Image,
		CreatedAt:   time.Now(),
		ContainerIP: containerIP,
		Hostname:    name,
		Host:        name,
		Port:        controller.Ports.Internal,
		ClientID:    clientID,
		Controller:  controller,
	}

	// Add to map
	dc.containers[resp.ID] = info

	log.Printf("Started container: %s (%s) for client %s", name, resp.ID[:12], clientID)
	return info, nil
}

// StopContainer stops the container with the given ID
func (dc *DockerController) StopContainer(ctx context.Context, containerID string) error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	// Check if the container exists in our tracking map
	info, ok := dc.containers[containerID]
	if !ok {
		return fmt.Errorf("container with ID %q not found in internal tracking", containerID)
	}

	// Check if the container actually exists in Docker
	_, err := dc.client.ContainerInspect(ctx, containerID)
	if err != nil {
		// Container doesn't exist in Docker, just remove it from our map
		log.Printf("Container %s doesn't exist in Docker, removing from tracking: %v", containerID[:12], err)
		delete(dc.containers, containerID)
		return nil
	}

	// Convert timeout from Duration to seconds as int
	timeoutSeconds := 10

	// Stop the container with a timeout
	if err := dc.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
		log.Printf("Error stopping container %s: %v", containerID[:12], err)
		// Even if we can't stop it, try to remove it and clean up our map
	}

	// Remove the container
	if err := dc.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}); err != nil {
		log.Printf("Error removing container %s: %v", containerID[:12], err)
		// Even if removal fails, we should still clean up our map
	}

	// Remove from map
	delete(dc.containers, containerID)

	log.Printf("Stopped and removed container: %s (%s) for client %s", info.Name, containerID[:12], info.ClientID)
	return nil
}

// GetContainer returns the container with the given ID
func (dc *DockerController) GetContainer(containerID string) (*ContainerInfo, bool) {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	info, exists := dc.containers[containerID]
	return info, exists
}

// GetContainerByClientID returns the container for the given client ID
func (dc *DockerController) GetContainerByClientID(clientID string) (*ContainerInfo, bool) {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	for _, info := range dc.containers {
		if info.ClientID == clientID {
			return info, true
		}
	}

	return nil, false
}

// ListContainers returns a list of all active containers managed by this controller
func (dc *DockerController) ListContainers() []*ContainerInfo {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	containers := make([]*ContainerInfo, 0, len(dc.containers))
	for _, container := range dc.containers {
		containers = append(containers, container)
	}

	return containers
}

// CleanupContainers stops and removes all containers
func (dc *DockerController) CleanupContainers(ctx context.Context) error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	var lastErr error
	for id, info := range dc.containers {
		// Check if the container actually exists in Docker before trying to stop it
		_, err := dc.client.ContainerInspect(ctx, id)
		if err != nil {
			log.Printf("Container %s doesn't exist in Docker, skipping cleanup: %v", id[:12], err)
			delete(dc.containers, id)
			continue
		}

		// Stop and remove the container
		timeoutSeconds := 10
		if err := dc.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
			log.Printf("Error stopping container %s: %v", id[:12], err)
			lastErr = err
			// Continue to removal attempt even if stop failed
		}

		if err := dc.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		}); err != nil {
			log.Printf("Error removing container %s: %v", id[:12], err)
			lastErr = err
			// Continue to cleanup our tracking map even if removal failed
		}

		log.Printf("Cleaned up container: %s (%s)", info.Name, id[:12])
		delete(dc.containers, id)
	}

	return lastErr
}

// Close closes the Docker controller
func (dc *DockerController) Close() error {
	return dc.client.Close()
}

// ParseMemoryLimit parses a memory limit string (e.g., "512m") into bytes
func ParseMemoryLimit(limit string) int64 {
	limit = strings.ToLower(limit)

	var multiplier int64 = 1
	if strings.HasSuffix(limit, "k") {
		multiplier = 1024
		limit = strings.TrimSuffix(limit, "k")
	} else if strings.HasSuffix(limit, "m") {
		multiplier = 1024 * 1024
		limit = strings.TrimSuffix(limit, "m")
	} else if strings.HasSuffix(limit, "g") {
		multiplier = 1024 * 1024 * 1024
		limit = strings.TrimSuffix(limit, "g")
	}

	value := 0
	fmt.Sscanf(limit, "%d", &value)

	return int64(value) * multiplier
}

// GetContainerLogs retrieves logs from a container
func (dc *DockerController) GetContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	return dc.client.ContainerLogs(ctx, containerID, options)
}

// DiscoverControllerImages discovers controller Docker images by their labels
func (dc *DockerController) DiscoverControllerImages(ctx context.Context) ([]DiscoveredController, error) {
	log.Println("Discovering controller Docker images...")

	// Create filter for Docker images with the controller type label
	filters := filters.NewArgs()
	filters.Add("label", "org.discon.type=controller")

	// List images with controller labels
	images, err := dc.client.ImageList(ctx, types.ImageListOptions{
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing controller images: %w", err)
	}

	log.Printf("Found %d potential controller images", len(images))

	var controllers []DiscoveredController

	// Process each image to extract controller metadata
	for _, img := range images {
		// Skip images without repository tags
		if len(img.RepoTags) == 0 {
			log.Printf("Skipping image %s with no tags", img.ID[:12])
			continue
		}

		// Get detailed image information to access all labels
		inspect, _, err := dc.client.ImageInspectWithRaw(ctx, img.ID)
		if err != nil {
			log.Printf("Error inspecting image %s: %v", img.ID[:12], err)
			continue
		}

		// Extract controller info from labels
		id := getImageLabel(inspect, "org.discon.controller.id")
		name := getImageLabel(inspect, "org.discon.controller.name")
		version := getImageLabel(inspect, "org.discon.controller.version")
		description := getImageLabel(inspect, "org.discon.controller.description")
		libraryPath := getImageLabel(inspect, "org.discon.controller.library_path")
		procName := getImageLabel(inspect, "org.discon.controller.proc_name")
		created := getImageLabel(inspect, "org.discon.controller.created")
		portsStr := getImageLabel(inspect, "org.discon.controller.ports")

		// Skip if missing required fields
		if id == "" {
			id = strings.ReplaceAll(strings.ReplaceAll(img.RepoTags[0], "/", "-"), ":", "-")
			log.Printf("Image %s missing id label, using derived id: %s", img.ID[:12], id)
		}

		if name == "" || version == "" {
			log.Printf("Image %s missing required labels (name or version), skipping", img.ID[:12])
			continue
		}

		// Use image tag if created time not provided
		if created == "" {
			created = time.Now().UTC().Format(time.RFC3339)
		}

		// Parse ports if provided
		ports := ControllerPorts{
			Internal: 8080, // Default internal port
			External: 0,    // Will be dynamically assigned
		}
		if portsStr != "" {
			if err := json.Unmarshal([]byte(portsStr), &ports); err != nil {
				log.Printf("Error parsing ports for image %s: %v", img.ID[:12], err)
			}
		}

		// Create the controller entry
		controller := DiscoveredController{
			ID:           id,
			Name:         name,
			Version:      version,
			Image:        img.RepoTags[0],
			Description:  description,
			LibraryPath:  libraryPath,
			ProcName:     procName,
			Ports:        ports,
			CreatedAt:    created,
			IsValid:      true, // Will be updated by validation if enabled
			ValidateInfo: "",
		}

		// Validate controller if validation is enabled
		if dc.discovery != nil && dc.discovery.Validation.Enabled {
			controller.IsValid, controller.ValidateInfo = dc.ValidateController(ctx, controller)
		}

		controllers = append(controllers, controller)
		log.Printf("Discovered controller: %s (Image: %s)", controller.Name, controller.Image)
	}

	return controllers, nil
}

// getImageLabel is a helper function to get a label value with a default fallback
func getImageLabel(image types.ImageInspect, label string) string {
	if value, ok := image.Config.Labels[label]; ok {
		return value
	}
	return ""
}

// ValidateController validates that a controller image is valid
func (dc *DockerController) ValidateController(ctx context.Context, controller DiscoveredController) (bool, string) {
	log.Printf("Validating controller: %s (Image: %s)", controller.Name, controller.Image)

	// Skip validation if library path is not provided
	if controller.LibraryPath == "" {
		return false, "Library path not specified, validation failed"
	}

	// Create a temporary container to validate the controller
	containerName := fmt.Sprintf("%svalidate-%s", dc.config.ContainerPrefix, controller.ID)

	// Ensure no validation container with the same name exists
	if err := dc.ensureContainerDoesNotExist(ctx, containerName); err != nil {
		return false, fmt.Sprintf("Error cleaning up existing validation container: %v", err)
	}

	// Create container config
	config := &container.Config{
		Image:    controller.Image,
		Hostname: containerName,
		Cmd:      []string{"sleep", "10"}, // Just keep the container alive briefly
	}

	// Create the container
	resp, err := dc.client.ContainerCreate(ctx, config, nil, nil, nil, containerName)
	if err != nil {
		return false, fmt.Sprintf("Error creating validation container: %v", err)
	}

	// Make sure to clean up the container when done
	defer func() {
		if err := dc.client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			log.Printf("Error removing validation container %s: %v", resp.ID[:12], err)
		}
	}()

	// Start the container
	if err := dc.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return false, fmt.Sprintf("Error starting validation container: %v", err)
	}

	// Check if the library file exists
	checkCmd := []string{"ls", controller.LibraryPath}
	execConfig := types.ExecConfig{
		Cmd:          checkCmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create the exec instance
	execID, err := dc.client.ContainerExecCreate(ctx, resp.ID, execConfig)
	if err != nil {
		return false, fmt.Sprintf("Error creating exec for validation: %v", err)
	}

	// Start the exec instance
	resp2, err := dc.client.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return false, fmt.Sprintf("Error attaching to exec for validation: %v", err)
	}
	defer resp2.Close()

	// Read the output
	output, err := io.ReadAll(resp2.Reader)
	if err != nil {
		return false, fmt.Sprintf("Error reading validation output: %v", err)
	}

	// Check if the library file exists
	if strings.Contains(string(output), "No such file or directory") {
		return false, fmt.Sprintf("Library file not found at path: %s", controller.LibraryPath)
	}

	validationInfo := fmt.Sprintf("Controller is valid, library file exists at %s", controller.LibraryPath)

	// If symbol verification is enabled, check for the symbol
	if dc.discovery.Validation.VerifySymbols && controller.ProcName != "" {
		// Use nm to check for the symbol
		symbolCheckCmd := []string{"sh", "-c", fmt.Sprintf("nm -D %s | grep -w %s || echo 'Symbol not found'", controller.LibraryPath, controller.ProcName)}

		execConfig = types.ExecConfig{
			Cmd:          symbolCheckCmd,
			AttachStdout: true,
			AttachStderr: true,
		}

		// Create the exec instance for symbol check
		execID, err = dc.client.ContainerExecCreate(ctx, resp.ID, execConfig)
		if err != nil {
			log.Printf("Warning: Error creating exec for symbol verification: %v", err)
			return true, fmt.Sprintf("%s; Warning: Symbol verification failed: %v", validationInfo, err)
		}

		// Start the exec instance
		resp2, err = dc.client.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
		if err != nil {
			log.Printf("Warning: Error attaching to exec for symbol verification: %v", err)
			return true, fmt.Sprintf("%s; Warning: Symbol verification failed: %v", validationInfo, err)
		}
		defer resp2.Close()

		// Read the output
		output, err = io.ReadAll(resp2.Reader)
		if err != nil {
			log.Printf("Warning: Error reading symbol verification output: %v", err)
			return true, fmt.Sprintf("%s; Warning: Symbol verification failed: %v", validationInfo, err)
		}

		// Check if the symbol exists - only a warning if not found
		if strings.Contains(string(output), "Symbol not found") {
			log.Printf("Warning: Symbol %s not found in library file %s", controller.ProcName, controller.LibraryPath)
			return true, fmt.Sprintf("%s; Warning: Symbol %s not found", validationInfo, controller.ProcName)
		}
	}

	// Run test call if configured
	if dc.discovery.Validation.TestCall {
		// This would execute a test call to check if the controller responds correctly
		// Implementation depends on how controllers should be tested
		// For now, just log that test calls would happen here
		log.Printf("Test call validation would occur here for %s", controller.Name)
	}

	return true, validationInfo
}
