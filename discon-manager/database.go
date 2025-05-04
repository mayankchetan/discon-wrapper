package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ControllerDatabase represents the database of controller configurations
type ControllerDatabase struct {
	path        string
	controllers map[string]*Controller
	mutex       sync.RWMutex
}

// Controller represents a controller configuration
type Controller struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	Image        string    `json:"image"`
	Description  string    `json:"description"`
	LibraryPath  string    `json:"library_path"`
	ProcName     string    `json:"proc_name"`
	Ports        PortPair  `json:"ports"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsValid      bool      `json:"is_valid"`
	ValidateInfo string    `json:"validate_info,omitempty"`
}

// PortPair represents a pair of internal and external ports
type PortPair struct {
	Internal int `json:"internal"`
	External int `json:"external"`
}

// DatabaseFile represents the structure of the controllers.json file
type DatabaseFile struct {
	Controllers []*Controller `json:"controllers"`
}

// DiscoveryStats represents the statistics of the discovery process
type DiscoveryStats struct {
	Added     int
	Updated   int
	Removed   int
	Unchanged int
	Failed    int
	FailedIDs []string
}

// NewControllerDatabase creates a new controller database
func NewControllerDatabase(path string) (*ControllerDatabase, error) {
	db := &ControllerDatabase{
		path:        path,
		controllers: make(map[string]*Controller),
	}

	// Load controllers from the database file
	if err := db.Load(); err != nil {
		return nil, err
	}

	return db, nil
}

// Load loads the controllers from the database file
func (db *ControllerDatabase) Load() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Check if the file exists
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		// Create an empty database file
		if err := db.save(); err != nil {
			return err
		}
		return nil
	}

	// Read the database file
	data, err := os.ReadFile(db.path)
	if err != nil {
		return fmt.Errorf("error reading database file: %w", err)
	}

	// Unmarshal the JSON data
	var dbFile DatabaseFile
	if err := json.Unmarshal(data, &dbFile); err != nil {
		return fmt.Errorf("error unmarshaling database file: %w", err)
	}

	// Add controllers to the map
	for _, controller := range dbFile.Controllers {
		db.controllers[controller.ID] = controller
	}

	return nil
}

// GetController returns a controller by ID
func (db *ControllerDatabase) GetController(id string) (*Controller, bool) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	controller, ok := db.controllers[id]
	return controller, ok
}

// GetControllerByVersion returns a controller by version
func (db *ControllerDatabase) GetControllerByVersion(version string) (*Controller, bool) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	for _, controller := range db.controllers {
		if controller.Version == version {
			return controller, true
		}
	}

	return nil, false
}

// ListControllers returns a list of all controllers
func (db *ControllerDatabase) ListControllers() []*Controller {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	controllers := make([]*Controller, 0, len(db.controllers))
	for _, controller := range db.controllers {
		controllers = append(controllers, controller)
	}

	return controllers
}

// AddController adds a controller to the database
func (db *ControllerDatabase) AddController(controller *Controller) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Check for periods in controller ID - they're not allowed because they cause issues in URLs
	if strings.Contains(controller.ID, ".") {
		return fmt.Errorf("controller ID %q contains periods (.), which are not allowed; use hyphens (-) or underscores (_) instead", controller.ID)
	}

	// Set timestamps if not set
	if controller.CreatedAt.IsZero() {
		controller.CreatedAt = time.Now()
	}
	if controller.UpdatedAt.IsZero() {
		controller.UpdatedAt = time.Now()
	}

	db.controllers[controller.ID] = controller

	// Save the database
	return db.save()
}

// UpdateController updates an existing controller in the database
func (db *ControllerDatabase) UpdateController(controller *Controller) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Check for periods in controller ID - they're not allowed because they cause issues in URLs
	if strings.Contains(controller.ID, ".") {
		return fmt.Errorf("controller ID %q contains periods (.), which are not allowed; use hyphens (-) or underscores (_) instead", controller.ID)
	}

	// Check if controller exists
	existing, ok := db.controllers[controller.ID]
	if !ok {
		return fmt.Errorf("controller with ID %q not found", controller.ID)
	}

	// Update timestamp
	controller.CreatedAt = existing.CreatedAt
	controller.UpdatedAt = time.Now()

	// Replace controller
	db.controllers[controller.ID] = controller

	// Save the database
	return db.save()
}

// RemoveController removes a controller from the database
func (db *ControllerDatabase) RemoveController(id string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, ok := db.controllers[id]; !ok {
		return fmt.Errorf("controller with ID %q not found", id)
	}

	delete(db.controllers, id)

	// Save the database
	return db.save()
}

// save saves the controllers to the database file
func (db *ControllerDatabase) save() error {
	// Create the database file structure
	dbFile := DatabaseFile{
		Controllers: make([]*Controller, 0, len(db.controllers)),
	}

	// Add controllers to the list
	for _, controller := range db.controllers {
		dbFile.Controllers = append(dbFile.Controllers, controller)
	}

	// Marshal the JSON data
	data, err := json.MarshalIndent(dbFile, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling database file: %w", err)
	}

	// Write the database file
	if err := os.WriteFile(db.path, data, 0644); err != nil {
		return fmt.Errorf("error writing database file: %w", err)
	}

	return nil
}

// RegisterDiscoveredControllers registers controllers that were discovered from Docker images
func (db *ControllerDatabase) RegisterDiscoveredControllers(discovered []DiscoveredController, removeMissing bool) (*DiscoveryStats, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	stats := &DiscoveryStats{
		FailedIDs: make([]string, 0),
	}

	// Create map of discovered controller IDs
	discoveredIds := make(map[string]bool)

	// Process each discovered controller
	for _, disc := range discovered {
		// Check for periods in controller ID - they're not allowed because they cause issues in URLs
		if strings.Contains(disc.ID, ".") {
			// Don't fail, but log this in stats and skip the controller
			stats.Failed++
			stats.FailedIDs = append(stats.FailedIDs, disc.ID)
			continue
		}

		// Generate a controller from discovered info
		controller := &Controller{
			ID:          disc.ID,
			Name:        disc.Name,
			Version:     disc.Version,
			Image:       disc.Image,
			Description: disc.Description,
			LibraryPath: disc.LibraryPath,
			ProcName:    disc.ProcName,
			Ports: PortPair{
				Internal: disc.Ports.Internal,
				External: disc.Ports.External,
			},
			UpdatedAt:    time.Now(),
			IsValid:      disc.IsValid,
			ValidateInfo: disc.ValidateInfo,
		}

		// Parse creation time if available, otherwise use now
		if disc.CreatedAt != "" {
			if created, err := time.Parse(time.RFC3339, disc.CreatedAt); err == nil {
				controller.CreatedAt = created
			} else {
				controller.CreatedAt = time.Now()
			}
		} else {
			controller.CreatedAt = time.Now()
		}

		// Track failing controllers for stats, but still register them if
		// they're valid but have warnings (like missing proc_name symbol)
		if !disc.IsValid {
			stats.Failed++
			stats.FailedIDs = append(stats.FailedIDs, disc.ID)

			// Skip controllers that critically failed validation (completely invalid)
			// Controllers with just warnings about symbol verification will have IsValid=true
			// but will have ValidateInfo containing warning messages
			continue
		}

		// Track this ID as discovered
		discoveredIds[disc.ID] = true

		// Check if controller already exists
		existing, exists := db.controllers[disc.ID]
		if exists {
			// Check if any fields have changed
			if controllerNeedsUpdate(existing, controller) {
				// Keep creation time from existing record
				controller.CreatedAt = existing.CreatedAt

				// Update controller
				db.controllers[disc.ID] = controller
				stats.Updated++
			} else {
				stats.Unchanged++
			}
		} else {
			// Add new controller
			db.controllers[disc.ID] = controller
			stats.Added++
		}
	}

	// Remove controllers that no longer exist in Docker
	if removeMissing {
		for id := range db.controllers {
			if !discoveredIds[id] {
				delete(db.controllers, id)
				stats.Removed++
			}
		}
	}

	// Save changes to database file
	if stats.Added > 0 || stats.Updated > 0 || stats.Removed > 0 {
		if err := db.save(); err != nil {
			return stats, fmt.Errorf("error saving database after registration: %w", err)
		}
	}

	return stats, nil
}

// controllerNeedsUpdate checks if a controller needs to be updated
func controllerNeedsUpdate(existing, new *Controller) bool {
	if existing.Name != new.Name ||
		existing.Version != new.Version ||
		existing.Image != new.Image ||
		existing.Description != new.Description ||
		existing.LibraryPath != new.LibraryPath ||
		existing.ProcName != new.ProcName ||
		existing.Ports.Internal != new.Ports.Internal ||
		existing.Ports.External != new.Ports.External ||
		existing.IsValid != new.IsValid ||
		existing.ValidateInfo != new.ValidateInfo {
		return true
	}
	return false
}
