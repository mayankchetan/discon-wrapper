package main

import (
	"context"
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

// NewDockerController creates a new Docker controller
func NewDockerController(ctx context.Context, config DockerConfig) (*DockerController, error) {
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

	// Check if the container exists
	info, ok := dc.containers[containerID]
	if !ok {
		return fmt.Errorf("container with ID %q not found", containerID)
	}

	// Convert timeout from Duration to seconds as int
	timeoutSeconds := 10

	// Stop the container with a timeout
	if err := dc.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
		return fmt.Errorf("error stopping container: %w", err)
	}

	// Remove the container
	if err := dc.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}); err != nil {
		return fmt.Errorf("error removing container: %w", err)
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

	info, ok := dc.containers[containerID]
	return info, ok
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

// ListContainers returns a list of all containers
func (dc *DockerController) ListContainers() []*ContainerInfo {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	containers := make([]*ContainerInfo, 0, len(dc.containers))
	for _, info := range dc.containers {
		containers = append(containers, info)
	}

	return containers
}

// CleanupContainers stops and removes all containers
func (dc *DockerController) CleanupContainers(ctx context.Context) error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	var lastErr error
	for id, info := range dc.containers {
		// Stop and remove the container
		timeoutSeconds := 10
		if err := dc.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
			log.Printf("Error stopping container %s: %v", id[:12], err)
			lastErr = err
			continue
		}

		if err := dc.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		}); err != nil {
			log.Printf("Error removing container %s: %v", id[:12], err)
			lastErr = err
			continue
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
