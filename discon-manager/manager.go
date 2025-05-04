package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	// Using imports from the root module
	dw "discon-wrapper"
	"discon-wrapper/shared/utils"

	"github.com/docker/docker/api/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Manager coordinates WebSocket connections and container management
type Manager struct {
	config           *Config
	dockerController *DockerController
	database         *ControllerDatabase
	connections      map[string]*ClientConnection
	connMutex        sync.RWMutex
	upgrader         websocket.Upgrader
	ctx              context.Context
	server           *http.Server
	logger           *utils.DebugLogger
	connCounter      int32         // Counter for connection IDs
	counterMutex     sync.Mutex    // Mutex for the counter
	adminHandler     *AdminHandler // Handler for admin UI and API
}

// ClientConnection represents a connection from a client
type ClientConnection struct {
	ID             string
	RemoteAddr     string
	ConnectedAt    time.Time
	LastActivityAt time.Time
	ContainerID    string
	ContainerInfo  *ContainerInfo
	ControllerID   string
	ControllerPath string
	ProcName       string
	WS             *websocket.Conn
	ProxyCloseCh   chan struct{}
	manager        *Manager
	logger         *utils.DebugLogger
}

// NewManager creates a new manager
func NewManager(ctx context.Context, config *Config) (*Manager, error) {
	// Create main logger
	logger := utils.NewDebugLogger(config.Server.DebugLevel, "discon-manager")

	// Create controller discovery config if needed
	var discoveryConfig *ControllerDiscoveryConfig
	if config.ControllerDiscovery.Mode != "manual" {
		discoveryConfig = &config.ControllerDiscovery
	}

	// Create Docker controller
	dockerController, err := NewDockerController(ctx, config.Docker, discoveryConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating Docker controller: %w", err)
	}

	// Create controller database
	database, err := NewControllerDatabase(config.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("error creating controller database: %w", err)
	}

	// Create manager
	manager := &Manager{
		config:           config,
		dockerController: dockerController,
		database:         database,
		connections:      make(map[string]*ClientConnection),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all connections
			},
		},
		ctx:         ctx,
		logger:      logger,
		connCounter: 0, // Initialize connection counter
	}

	return manager, nil
}

// Start starts the manager
func (m *Manager) Start() error {
	// Create router
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", m.HandleWebSocket)
	mux.HandleFunc("/health", m.HandleHealth)
	mux.HandleFunc("/metrics", m.HandleMetrics)
	mux.HandleFunc("/containers", m.HandleContainers)
	mux.HandleFunc("/controllers", m.HandleControllers)

	// Initialize and set up admin handler
	m.adminHandler = NewAdminHandler(m)
	m.adminHandler.SetupRoutes(mux)

	// Create server
	m.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", m.config.Server.Host, m.config.Server.Port),
		Handler: mux,
	}

	// If controller discovery is enabled, run discovery
	if m.config.ControllerDiscovery.Mode == "startup" || m.config.ControllerDiscovery.Mode == "periodic" {
		if err := m.RunControllerDiscovery(); err != nil {
			m.logger.Error("Error running controller discovery: %v", err)
		}
	}

	// If periodic discovery is enabled, start discovery goroutine
	if m.config.ControllerDiscovery.Mode == "periodic" && m.config.ControllerDiscovery.IntervalMinutes > 0 {
		go m.periodicDiscoveryRoutine()
	}

	// Start server in a goroutine
	go func() {
		m.logger.Debug("Starting server on %s", m.server.Addr)
		if err := m.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Error starting server: %v", err)
		}
	}()

	// Start cleanup goroutine
	go m.cleanupRoutine()

	return nil
}

// Stop stops the manager
func (m *Manager) Stop() error {
	// Stop server
	if m.server != nil {
		// Create a deadline to wait for current connections to finish
		ctx, cancel := context.WithTimeout(m.ctx, 15*time.Second)
		defer cancel()

		m.logger.Debug("Shutting down server...")
		if err := m.server.Shutdown(ctx); err != nil {
			m.logger.Error("Error shutting down server: %v", err)
		}
	}

	// Close all client connections
	m.connMutex.Lock()
	for _, conn := range m.connections {
		conn.Close()
	}
	m.connMutex.Unlock()

	// Stop all containers
	if err := m.dockerController.CleanupContainers(m.ctx); err != nil {
		m.logger.Error("Error cleaning up containers: %v", err)
	}

	// Close Docker controller
	if err := m.dockerController.Close(); err != nil {
		m.logger.Error("Error closing Docker controller: %v", err)
	}

	return nil
}

// cleanupRoutine periodically checks for inactive containers and removes them
func (m *Manager) cleanupRoutine() {
	ticker := time.NewTicker(time.Duration(m.config.Docker.CleanupTimeout) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupInactiveContainers()
		}
	}
}

// cleanupInactiveContainers removes containers that haven't been used for a while
func (m *Manager) cleanupInactiveContainers() {
	m.connMutex.RLock()
	defer m.connMutex.RUnlock()

	// Get a list of connections to clean up to avoid modification during iteration
	var containersToCleanup []struct {
		containerID string
		clientID    string
	}

	// Check if any containers need to be cleaned up
	for _, conn := range m.connections {
		if conn.ContainerID == "" {
			continue
		}

		// If the connection hasn't been active for a while, clean up the container
		if time.Since(conn.LastActivityAt) > time.Duration(m.config.Docker.CleanupTimeout)*time.Second {
			// Add to our cleanup list instead of cleaning immediately
			containersToCleanup = append(containersToCleanup, struct {
				containerID string
				clientID    string
			}{
				containerID: conn.ContainerID,
				clientID:    conn.ID,
			})
		}
	}

	// Clean up the containers outside the lock
	for _, container := range containersToCleanup {
		// Double-check the container is still in our tracking map before trying to clean up
		if _, exists := m.dockerController.GetContainer(container.containerID); exists {
			m.logger.Debug("Cleaning up inactive container %s for client %s", container.containerID, container.clientID)
			go func(containerID string) {
				if err := m.dockerController.StopContainer(m.ctx, containerID); err != nil {
					m.logger.Error("Error stopping container %s: %v", containerID, err)
				}
			}(container.containerID)
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Create unique identifier for this connection
	connID := uuid.New().String()

	// Get a unique numeric connection ID for logging
	m.counterMutex.Lock()
	connNumID := m.connCounter
	m.connCounter++
	m.counterMutex.Unlock()

	// Create a connection-specific logger with the unique connection ID
	logger := utils.NewConnectionLogger(m.config.Server.DebugLevel, "discon-manager", connNumID)

	// Parse query parameters
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Error parsing query parameters: "+err.Error(), http.StatusBadRequest)
		logger.Error("Error parsing query parameters: %v", err)
		return
	}

	// Extract controller parameters
	controllerPath := params.Get("path")
	procName := params.Get("proc")
	controllerID := params.Get("controller")
	controllerVersion := params.Get("version")

	logger.Debug("New connection from %s requesting controller %s (path: %s, proc: %s, version: %s)",
		r.RemoteAddr, controllerID, controllerPath, procName, controllerVersion)

	// Find controller by using the following priority:
	// 1. If controllerID is provided, use that
	// 2. If controllerPath is provided and matches a controller ID, use that
	// 3. If controllerVersion is provided, use that
	// 4. Use the first controller as default
	var controller *Controller

	if controllerID != "" {
		// Method 1: By explicit controller ID
		var ok bool
		controller, ok = m.database.GetController(controllerID)
		if !ok {
			http.Error(w, "Controller not found: "+controllerID, http.StatusBadRequest)
			logger.Error("Controller not found: %s", controllerID)
			return
		}
		logger.Debug("Selected controller by ID: %s", controllerID)
	} else if controllerPath != "" {
		// Method 2: Try to match path to a controller ID
		// Extract potential controller ID from path (e.g., "discon-server-rosco" -> "rosco")
		pathParts := strings.Split(controllerPath, "-")
		if len(pathParts) > 0 {
			potentialID := pathParts[len(pathParts)-1]
			controller, _ = m.database.GetController(potentialID)

			// If found by path-derived ID, use it
			if controller != nil {
				logger.Debug("Selected controller by path-derived ID: %s from path %s", potentialID, controllerPath)
			} else {
				// Try exact match on controller ID using the full path
				controller, _ = m.database.GetController(controllerPath)
				if controller != nil {
					logger.Debug("Selected controller by exact path match: %s", controllerPath)
				}
			}
		}
	}

	// If still no controller, try by version
	if controller == nil && controllerVersion != "" {
		var ok bool
		controller, ok = m.database.GetControllerByVersion(controllerVersion)
		if !ok {
			http.Error(w, "Controller version not found: "+controllerVersion, http.StatusBadRequest)
			logger.Error("Controller version not found: %s", controllerVersion)
			return
		}
		logger.Debug("Selected controller by version: %s", controllerVersion)
	}

	// If still no controller, use first controller as default
	if controller == nil {
		controllers := m.database.ListControllers()
		if len(controllers) == 0 {
			http.Error(w, "No controllers available", http.StatusInternalServerError)
			logger.Error("No controllers available")
			return
		}
		controller = controllers[0]
		logger.Debug("Selected default controller: %s", controller.ID)
	}

	logger.Debug("Using controller: %s (image: %s)", controller.ID, controller.Image)

	// IMPORTANT: Do NOT override the library path with the client's path parameter
	// That parameter is only used for controller selection
	// Only override the proc name if provided
	if procName != "" {
		controller.ProcName = procName
	}

	logger.Debug("Using controller library path: %s and proc: %s", controller.LibraryPath, controller.ProcName)

	// Start container for this controller
	containerInfo, err := m.dockerController.StartContainer(m.ctx, controller, connID)
	if err != nil {
		http.Error(w, "Error starting controller container: "+err.Error(), http.StatusInternalServerError)
		logger.Error("Error starting controller container: %v", err)
		return
	}

	logger.Debug("Started container: %s with IP %s", containerInfo.Name, containerInfo.ContainerIP)

	// Upgrade connection to WebSocket
	clientConn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Error upgrading to WebSocket: %v", err)
		go func() {
			// Clean up the container if we couldn't establish the WebSocket connection
			if err := m.dockerController.StopContainer(m.ctx, containerInfo.ID); err != nil {
				logger.Error("Error stopping container %s: %v", containerInfo.ID[:12], err)
			}
		}()
		return
	}

	// Create client connection
	conn := &ClientConnection{
		ID:             connID,
		RemoteAddr:     r.RemoteAddr,
		ConnectedAt:    time.Now(),
		LastActivityAt: time.Now(),
		ContainerID:    containerInfo.ID,
		ContainerInfo:  containerInfo,
		ControllerID:   controller.ID,
		ControllerPath: controller.LibraryPath,
		ProcName:       controller.ProcName,
		WS:             clientConn,
		ProxyCloseCh:   make(chan struct{}),
		manager:        m,
		logger:         logger,
	}

	// Register connection
	m.connMutex.Lock()
	m.connections[connID] = conn
	m.connMutex.Unlock()

	// Start proxying WebSocket connections
	go conn.proxyConnectionToContainer()
}

// HandleHealth handles health check requests
func (m *Manager) HandleHealth(w http.ResponseWriter, r *http.Request) {
	// Basic health check implementation
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// HandleMetrics handles metrics requests
func (m *Manager) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	// Collect container metrics
	containers := m.dockerController.ListContainers()

	// Basic metrics implementation
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"active_containers":%d}`, len(containers))
}

// HandleContainers handles container listing requests
func (m *Manager) HandleContainers(w http.ResponseWriter, r *http.Request) {
	// List containers
	containers := m.dockerController.ListContainers()

	// Basic container listing implementation
	w.Header().Set("Content-Type", "text/plain")
	for _, container := range containers {
		fmt.Fprintf(w, "Container: %s (ID: %s, Client: %s)\n",
			container.Name, container.ID[:12], container.ClientID)
	}
}

// HandleControllers handles controller listing requests
func (m *Manager) HandleControllers(w http.ResponseWriter, r *http.Request) {
	// List controllers
	controllers := m.database.ListControllers()

	// Basic controller listing implementation
	w.Header().Set("Content-Type", "text/plain")
	for _, controller := range controllers {
		fmt.Fprintf(w, "Controller: %s (Version: %s, Image: %s)\n",
			controller.Name, controller.Version, controller.Image)
	}
}

// getActiveContainers returns a list of active containers for the admin UI
func (m *Manager) getActiveContainers() []*ContainerInfo {
	return m.dockerController.ListContainers()
}

// proxyConnectionToContainer proxies the client WebSocket connection to the container
func (cc *ClientConnection) proxyConnectionToContainer() {
	// Create URL for the server WebSocket
	targetURL := fmt.Sprintf("ws://%s:%d/ws", cc.ContainerInfo.Host, cc.ContainerInfo.Port)

	// Add query parameters for shared library path and proc
	u, err := url.Parse(targetURL)
	if err != nil {
		cc.logger.Error("Error parsing target URL: %v", err)
		cc.Close()
		return
	}

	q := u.Query()

	// IMPORTANT: Use the controller's library_path value from the database
	// rather than the path parameter from the client
	cc.logger.Debug("Using controller library path: %s and proc: %s", cc.ControllerPath, cc.ProcName)
	q.Add("path", cc.ControllerPath)
	q.Add("proc", cc.ProcName)
	u.RawQuery = q.Encode()

	cc.logger.Debug("Connecting to container WebSocket at %s", u.String())

	// Check container logs to see if the server is ready
	// This helps us know when the WebSocket server inside the container is actually ready
	serverReady := make(chan bool, 1)
	go func() {
		// Poll container logs for signs the server is ready
		startTime := time.Now()
		maxWaitTime := 30 * time.Second

		for time.Since(startTime) < maxWaitTime {
			// Get container logs
			logReader, err := cc.manager.dockerController.GetContainerLogs(
				cc.manager.ctx,
				cc.ContainerID,
				types.ContainerLogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Follow:     false,
					Tail:       "20",
				},
			)

			if err == nil {
				// Read logs and check for server ready message
				logs := make([]byte, 8192) // Increased buffer size for more log context
				n, _ := logReader.Read(logs)
				logReader.Close()

				// Even if we can't read logs, check if the container is ready via TCP directly
				if n > 0 {
					logsStr := string(logs[:n])
					cc.logger.Debug("Container logs excerpt: %s", logsStr)

					// Look for signs the server is ready
					if strings.Contains(logsStr, "WebSocket server started") ||
						strings.Contains(logsStr, "Starting WebSocket server") ||
						strings.Contains(logsStr, "Server initialized") {
						cc.logger.Debug("Container WebSocket server is starting up, detected in logs")
						// Continue to connection attempts, but with better odds
						break
					}
				}
			}

			// Check if we can make a TCP connection to the port
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", cc.ContainerInfo.ContainerIP, cc.ContainerInfo.Port), 500*time.Millisecond)
			if err == nil {
				conn.Close()
				// TCP connection successful, server is likely up
				cc.logger.Debug("TCP connection to container port successful, likely ready")
				break
			}

			time.Sleep(500 * time.Millisecond)
		}

		// Signal that the container initialization phase is done
		serverReady <- true
	}()

	// Wait for initialization check to complete or timeout
	select {
	case <-serverReady:
		cc.logger.Debug("Proceeding with WebSocket connection attempts after container initialization check")
	case <-time.After(10 * time.Second):
		cc.logger.Debug("Container initialization check timed out, proceeding anyway")
	}

	// Connect to the WebSocket server in the container with retry logic
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 3 * time.Second

	// Add retry logic
	maxRetries := 12 // Allow up to 30 seconds for container startup
	var serverConn *websocket.Conn
	var dialErr error
	var backoffDuration time.Duration = 500 * time.Millisecond

	// In host network mode, skip hostname attempts and only use IP
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			cc.logger.Debug("Retrying connection to container (attempt %d/%d)...", retry+1, maxRetries)
			// Wait with increasing backoff
			time.Sleep(backoffDuration)
			// Increase backoff for next time, but cap at 2.5 seconds
			backoffDuration += 250 * time.Millisecond
			if backoffDuration > 2500*time.Millisecond {
				backoffDuration = 2500 * time.Millisecond
			}
		}

		// Since we're using host networking, only try IP address - don't try hostname
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		ipURL := *u // Make a copy of the URL
		ipURL.Host = fmt.Sprintf("%s:%d", cc.ContainerInfo.ContainerIP, cc.ContainerInfo.Port)

		serverConn, _, dialErr = dialer.DialContext(ctx, ipURL.String(), nil)
		cancel() // Cancel the timeout context

		if dialErr == nil {
			break // Connection successful
		}

		cc.logger.Debug("IP connection attempt %d failed: %v", retry+1, dialErr)

		// Check if container is still running
		if retry > 0 && retry%3 == 0 {
			containerInfo, err := cc.manager.dockerController.client.ContainerInspect(cc.manager.ctx, cc.ContainerID)
			if err != nil || !containerInfo.State.Running {
				cc.logger.Error("Container is no longer running, aborting connection attempts")
				break
			}

			// Check if the library file exists in the container
			if retry == 3 {
				// Diagnostic: Check if library file exists
				execConfig := types.ExecConfig{
					Cmd:          []string{"ls", "-l", cc.ControllerPath},
					AttachStdout: true,
					AttachStderr: true,
				}

				execID, err := cc.manager.dockerController.client.ContainerExecCreate(cc.manager.ctx, cc.ContainerID, execConfig)
				if err != nil {
					cc.logger.Error("Failed to create exec: %v", err)
				} else {
					resp, err := cc.manager.dockerController.client.ContainerExecAttach(cc.manager.ctx, execID.ID, types.ExecStartCheck{})
					if err != nil {
						cc.logger.Error("Failed to attach to exec: %v", err)
					} else {
						defer resp.Close()
						output := make([]byte, 1024)
						n, _ := resp.Reader.Read(output)
						if n > 0 {
							cc.logger.Debug("File check result: %s", string(output[:n]))
						}
					}

					inspectResp, err := cc.manager.dockerController.client.ContainerExecInspect(cc.manager.ctx, execID.ID)
					if err != nil {
						cc.logger.Error("Failed to inspect exec: %v", err)
					} else {
						if inspectResp.ExitCode != 0 {
							cc.logger.Error("Library file %s does not exist in container or is not accessible", cc.ControllerPath)
						}
					}
				}
			}
		}
	}

	if dialErr != nil {
		cc.logger.Error("Error connecting to container WebSocket: %v", dialErr)
		cc.Close()
		return
	}

	defer serverConn.Close()

	cc.logger.Debug("Connected to container WebSocket, starting proxy")

	// Create done channel for goroutine synchronization
	done := make(chan struct{})

	// Proxy client messages to server
	go func() {
		defer close(done)

		for {
			// Read message from client with timeout
			cc.WS.SetReadDeadline(time.Now().Add(30 * time.Second))
			messageType, message, err := cc.WS.ReadMessage()
			cc.WS.SetReadDeadline(time.Time{}) // Clear deadline

			if err != nil {
				cc.logger.Debug("Error reading from client: %v", err)
				return
			}

			cc.LastActivityAt = time.Now()

			// If it's a binary message, try to unmarshal it for logging
			if messageType == websocket.BinaryMessage && cc.logger.DebugLevel >= 2 {
				var payload dw.Payload
				if err := payload.UnmarshalBinary(message); err == nil {
					cc.logger.Verbose("Client -> Container: %v", payload)
				}
			}

			// Write the message to the server with timeout
			serverConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := serverConn.WriteMessage(messageType, message); err != nil {
				cc.logger.Debug("Error writing to server: %v", err)
				serverConn.SetWriteDeadline(time.Time{}) // Clear deadline
				return
			}
			serverConn.SetWriteDeadline(time.Time{}) // Clear deadline
		}
	}()

	// Proxy server messages to client
	for {
		select {
		case <-done:
			return
		case <-cc.ProxyCloseCh:
			return
		default:
			// Read message from server with timeout
			serverConn.SetReadDeadline(time.Now().Add(30 * time.Second))
			messageType, message, err := serverConn.ReadMessage()
			serverConn.SetReadDeadline(time.Time{}) // Clear deadline

			if err != nil {
				cc.logger.Debug("Error reading from server: %v", err)
				return
			}

			cc.LastActivityAt = time.Now()

			// If it's a binary message, try to unmarshal it for logging
			if messageType == websocket.BinaryMessage && cc.logger.DebugLevel >= 2 {
				var payload dw.Payload
				if err := payload.UnmarshalBinary(message); err == nil {
					cc.logger.Verbose("Container -> Client: %v", payload)
				}
			}

			// Write the message to the client with timeout
			cc.WS.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := cc.WS.WriteMessage(messageType, message); err != nil {
				cc.logger.Debug("Error writing to client: %v", err)
				cc.WS.SetWriteDeadline(time.Time{}) // Clear deadline
				return
			}
			cc.WS.SetWriteDeadline(time.Time{}) // Clear deadline
		}
	}
}

// Close closes the client connection and cleans up resources
func (cc *ClientConnection) Close() {
	// Signal the proxy goroutine to stop
	close(cc.ProxyCloseCh)

	// Close the WebSocket connection
	if cc.WS != nil {
		cc.WS.Close()
	}

	// Stop the container
	if cc.ContainerID != "" {
		cc.logger.Debug("Stopping container %s", cc.ContainerID[:12])
		if err := cc.manager.dockerController.StopContainer(cc.manager.ctx, cc.ContainerID); err != nil {
			cc.logger.Error("Error stopping container %s: %v", cc.ContainerID[:12], err)
		}
	}

	// Remove the connection from the manager
	cc.manager.connMutex.Lock()
	delete(cc.manager.connections, cc.ID)
	cc.manager.connMutex.Unlock()

	cc.logger.Debug("Connection closed")
}

// RunControllerDiscovery discovers controller images and registers them
func (m *Manager) RunControllerDiscovery() error {
	m.logger.Debug("Running controller discovery...")

	// Discover controller images
	controllers, err := m.dockerController.DiscoverControllerImages(m.ctx)
	if err != nil {
		return fmt.Errorf("error discovering controller images: %w", err)
	}

	m.logger.Debug("Found %d controller images", len(controllers))

	// Register controllers in the database
	stats, err := m.database.RegisterDiscoveredControllers(
		controllers,
		m.config.ControllerDiscovery.RemoveMissing,
	)
	if err != nil {
		return fmt.Errorf("error registering controllers: %w", err)
	}

	m.logger.Debug("Controller registration stats: Added=%d, Updated=%d, Removed=%d, Failed=%d",
		stats.Added, stats.Updated, stats.Removed, stats.Failed)

	if stats.Failed > 0 {
		m.logger.Error("Some controllers failed validation but will be used with warnings: %v", stats.FailedIDs)
	}

	return nil
}

// periodicDiscoveryRoutine periodically runs controller discovery
func (m *Manager) periodicDiscoveryRoutine() {
	interval := time.Duration(m.config.ControllerDiscovery.IntervalMinutes) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.logger.Debug("Starting periodic controller discovery (interval: %v)", interval)

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := m.RunControllerDiscovery(); err != nil {
				m.logger.Error("Error running periodic controller discovery: %v", err)
			}
		}
	}
}
