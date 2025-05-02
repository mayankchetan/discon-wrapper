package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/google/uuid"
)

// AdminHandler handles the admin web interface and API endpoints
type AdminHandler struct {
	manager     *Manager
	templates   *template.Template
	templateDir string
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(manager *Manager) *AdminHandler {
	templateDir := "templates"

	// Create template instance with custom functions
	funcMap := template.FuncMap{
		"truncateID": func(id string) string {
			if len(id) > 12 {
				return id[:12]
			}
			return id
		},
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
	}

	// Parse templates
	templates := template.New("").Funcs(funcMap)
	templateFiles, _ := filepath.Glob(filepath.Join(templateDir, "*.html"))
	templates, _ = templates.ParseFiles(templateFiles...)

	return &AdminHandler{
		manager:     manager,
		templates:   templates,
		templateDir: templateDir,
	}
}

// SetupRoutes sets up the admin routes
func (ah *AdminHandler) SetupRoutes(mux *http.ServeMux) {
	// Main admin UI
	mux.HandleFunc("/admin", ah.authMiddleware(ah.handleAdminUI))
	mux.HandleFunc("/admin/login", ah.handleLogin)

	// Admin API endpoints
	mux.HandleFunc("/admin/controllers", ah.authMiddleware(ah.handleControllers))
	mux.HandleFunc("/admin/controllers/", ah.authMiddleware(ah.handleControllerOperations))
	mux.HandleFunc("/admin/containers", ah.authMiddleware(ah.handleContainers))
	mux.HandleFunc("/admin/containers/", ah.authMiddleware(ah.handleContainerOperations))
}

// authMiddleware wraps an HTTP handler with authentication checks
func (ah *AdminHandler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only perform authentication if enabled
		if ah.manager.config.Auth.Enabled {
			session, err := r.Cookie("session")

			// No valid session found, redirect to login
			if err != nil || !ah.validateSessionCookie(session.Value) {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}
		}

		// Authentication passed or disabled, proceed to the handler
		next(w, r)
	}
}

// validateSessionCookie validates a session cookie
func (ah *AdminHandler) validateSessionCookie(sessionValue string) bool {
	// For now, just check if session value is non-empty
	// In a real-world application, you might want to store sessions in a database
	// or validate against a cryptographically secure token
	return sessionValue != ""
}

// handleLogin handles the login page
func (ah *AdminHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// If auth is disabled, redirect to admin page
	if !ah.manager.config.Auth.Enabled {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Handle POST request (form submission)
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Constant time comparison to prevent timing attacks
		validUsername := subtle.ConstantTimeCompare([]byte(username), []byte(ah.manager.config.Auth.Username)) == 1
		validPassword := subtle.ConstantTimeCompare([]byte(password), []byte(ah.manager.config.Auth.Password)) == 1

		if validUsername && validPassword {
			// Create session
			sessionID := uuid.New().String()

			// Set cookie with session ID
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   3600, // 1 hour
				SameSite: http.SameSiteLaxMode,
			})

			// Redirect to admin page
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		// Invalid credentials, show login page with error
		ah.renderLoginPage(w, "Invalid username or password")
		return
	}

	// Handle GET request (show login form)
	ah.renderLoginPage(w, "")
}

// renderLoginPage renders the login page with optional error message
func (ah *AdminHandler) renderLoginPage(w http.ResponseWriter, errorMessage string) {
	data := struct {
		Error string
	}{
		Error: errorMessage,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := ah.templates.ExecuteTemplate(w, "login.html", data); err != nil {
		http.Error(w, "Error rendering login page: "+err.Error(), http.StatusInternalServerError)
	}
}

// handleAdminUI handles the main admin UI
func (ah *AdminHandler) handleAdminUI(w http.ResponseWriter, r *http.Request) {
	controllers := ah.manager.database.ListControllers()

	// Get active containers
	containers := ah.manager.getActiveContainers()

	// Render the admin template
	ah.renderAdminUI(w, controllers, containers)
}

// renderAdminUI renders the admin UI with controllers and containers data
func (ah *AdminHandler) renderAdminUI(w http.ResponseWriter, controllers []*Controller, containers []*ContainerInfo) {
	data := struct {
		Controllers []*Controller
		Containers  []*ContainerInfo
	}{
		Controllers: controllers,
		Containers:  containers,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := ah.templates.ExecuteTemplate(w, "admin.html", data); err != nil {
		http.Error(w, "Error rendering admin page: "+err.Error(), http.StatusInternalServerError)
	}
}

// handleControllers handles controller listing and creation
func (ah *AdminHandler) handleControllers(w http.ResponseWriter, r *http.Request) {
	// Handle GET request (list controllers)
	if r.Method == http.MethodGet {
		controllers := ah.manager.database.ListControllers()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(controllers)
		return
	}

	// Handle POST request (create controller)
	if r.Method == http.MethodPost {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Get form values
		id := r.FormValue("id")
		name := r.FormValue("name")
		version := r.FormValue("version")
		image := r.FormValue("image")
		description := r.FormValue("description")
		libraryPath := r.FormValue("library_path")
		procName := r.FormValue("proc_name")

		// Parse port values
		internalPort, err := strconv.Atoi(r.FormValue("internal_port"))
		if err != nil {
			http.Error(w, "Invalid internal port", http.StatusBadRequest)
			return
		}

		externalPort, err := strconv.Atoi(r.FormValue("external_port"))
		if err != nil {
			http.Error(w, "Invalid external port", http.StatusBadRequest)
			return
		}

		// Create new controller
		controller := &Controller{
			ID:          id,
			Name:        name,
			Version:     version,
			Image:       image,
			Description: description,
			LibraryPath: libraryPath,
			ProcName:    procName,
			Ports: PortPair{
				Internal: internalPort,
				External: externalPort,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Add controller to database
		if err := ah.manager.database.AddController(controller); err != nil {
			http.Error(w, fmt.Sprintf("Error adding controller: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to admin page
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Method not allowed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleControllerOperations handles operations on a specific controller
func (ah *AdminHandler) handleControllerOperations(w http.ResponseWriter, r *http.Request) {
	// Extract controller ID from URL
	path := r.URL.Path
	parts := splitPath(path)

	// Check if we have enough path parts
	if len(parts) < 3 {
		http.Error(w, "Invalid controller ID", http.StatusBadRequest)
		return
	}

	controllerID := parts[2]

	// Handle GET request (get controller details)
	if r.Method == http.MethodGet && len(parts) == 3 {
		controller, ok := ah.manager.database.GetController(controllerID)
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(controller)
		return
	}

	// Handle POST request for controller update
	if r.Method == http.MethodPost && len(parts) == 4 && parts[3] == "update" {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Get controller
		controller, ok := ah.manager.database.GetController(controllerID)
		if !ok {
			http.NotFound(w, r)
			return
		}

		// Update controller fields
		controller.Name = r.FormValue("name")
		controller.Version = r.FormValue("version")
		controller.Image = r.FormValue("image")
		controller.Description = r.FormValue("description")
		controller.LibraryPath = r.FormValue("library_path")
		controller.ProcName = r.FormValue("proc_name")
		controller.UpdatedAt = time.Now()

		// Parse port values
		internalPort, err := strconv.Atoi(r.FormValue("internal_port"))
		if err == nil {
			controller.Ports.Internal = internalPort
		}

		externalPort, err := strconv.Atoi(r.FormValue("external_port"))
		if err == nil {
			controller.Ports.External = externalPort
		}

		// Update controller in database
		if err := ah.manager.database.AddController(controller); err != nil {
			http.Error(w, fmt.Sprintf("Error updating controller: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to admin page
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Handle DELETE request
	if r.Method == http.MethodDelete && len(parts) == 3 {
		// Remove controller from database
		if err := ah.manager.database.RemoveController(controllerID); err != nil {
			http.Error(w, fmt.Sprintf("Error deleting controller: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle POST request for controller testing
	if r.Method == http.MethodPost && len(parts) == 4 && parts[3] == "test" {
		// Get controller
		controller, ok := ah.manager.database.GetController(controllerID)
		if !ok {
			http.NotFound(w, r)
			return
		}

		// Run the test
		success, output := ah.testController(controller)

		// Return result
		result := struct {
			Success bool   `json:"success"`
			Output  string `json:"output"`
		}{
			Success: success,
			Output:  output,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// Method not allowed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleContainers handles container listing
func (ah *AdminHandler) handleContainers(w http.ResponseWriter, r *http.Request) {
	// Only handle GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active containers
	containers := ah.manager.getActiveContainers()

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// handleContainerOperations handles operations on a specific container
func (ah *AdminHandler) handleContainerOperations(w http.ResponseWriter, r *http.Request) {
	// Extract container ID from URL
	path := r.URL.Path
	parts := splitPath(path)

	// Check if we have enough path parts
	if len(parts) < 4 || parts[3] != "stop" {
		http.Error(w, "Invalid container operation", http.StatusBadRequest)
		return
	}

	containerID := parts[2]

	// Handle POST request for stopping container
	if r.Method == http.MethodPost {
		// Stop container
		if err := ah.manager.dockerController.StopContainer(ah.manager.ctx, containerID); err != nil {
			http.Error(w, fmt.Sprintf("Error stopping container: %v", err), http.StatusInternalServerError)
			return
		}

		// Return success
		w.WriteHeader(http.StatusOK)
		return
	}

	// Method not allowed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// testController tests a controller by starting a container and checking if it works
func (ah *AdminHandler) testController(controller *Controller) (bool, string) {
	// Generate a test client ID
	testClientID := "test-" + uuid.New().String()

	// Start the container for testing
	containerInfo, err := ah.manager.dockerController.StartContainer(ah.manager.ctx, controller, testClientID)
	if err != nil {
		return false, fmt.Sprintf("Error starting container: %v", err)
	}

	// Make sure to clean up the container when done
	defer ah.manager.dockerController.StopContainer(ah.manager.ctx, containerInfo.ID)

	// Wait for container to start and initialize
	time.Sleep(2 * time.Second)

	// Check if container is running
	inspectData, err := ah.manager.dockerController.client.ContainerInspect(ah.manager.ctx, containerInfo.ID)
	if err != nil {
		return false, fmt.Sprintf("Error inspecting container: %v", err)
	}

	if !inspectData.State.Running {
		return false, "Container failed to start"
	}

	// Check if controller library exists
	execConfig := types.ExecConfig{
		Cmd:          []string{"ls", "-l", controller.LibraryPath},
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := ah.manager.dockerController.client.ContainerExecCreate(ah.manager.ctx, containerInfo.ID, execConfig)
	if err != nil {
		return false, fmt.Sprintf("Error creating exec: %v", err)
	}

	resp, err := ah.manager.dockerController.client.ContainerExecAttach(ah.manager.ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return false, fmt.Sprintf("Error attaching to exec: %v", err)
	}
	defer resp.Close()

	output := make([]byte, 1024)
	n, _ := resp.Reader.Read(output)

	if n == 0 {
		return false, "No output from container exec"
	}

	fileCheckOutput := string(output[:n])

	// Check exec exit code
	inspectResp, err := ah.manager.dockerController.client.ContainerExecInspect(ah.manager.ctx, execID.ID)
	if err != nil {
		return false, fmt.Sprintf("Error inspecting exec: %v", err)
	}

	if inspectResp.ExitCode != 0 {
		return false, fmt.Sprintf("Library check failed with exit code %d: %s", inspectResp.ExitCode, fileCheckOutput)
	}

	// All checks passed
	return true, fmt.Sprintf("Container started successfully\nIP: %s\nPort: %d\nLibrary check:\n%s",
		containerInfo.ContainerIP, containerInfo.Port, fileCheckOutput)
}

// Helper to split URL path
func splitPath(path string) []string {
	// Remove leading slash and split by slash
	path = path[1:]
	return split(path, '/')
}

// Helper to split string by separator
func split(s string, sep byte) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			if start < i {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
