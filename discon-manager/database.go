package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// ControllerDatabase represents the database of controller configurations
type ControllerDatabase struct {
	path       string
	controllers map[string]*Controller
	mutex      sync.RWMutex
}

// Controller represents a controller configuration
type Controller struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Image       string    `json:"image"`
	Description string    `json:"description"`
	LibraryPath string    `json:"library_path"`
	ProcName    string    `json:"proc_name"`
	Ports       PortPair  `json:"ports"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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

// NewControllerDatabase creates a new controller database
func NewControllerDatabase(path string) (*ControllerDatabase, error) {
	db := &ControllerDatabase{
		path:       path,
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