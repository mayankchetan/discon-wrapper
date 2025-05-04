package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
)

const (
	program = "discon-manager"
	version = "v0.1.0"
)

func main() {
	// Configure logging with timestamp, file and line number
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Display program and version
	log.Printf("Started %s %s", program, version)

	// Parse command-line arguments
	configPath := flag.String("config", "config/config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Create a new context that will be canceled on interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create signal channel for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a new controller manager
	manager, err := NewManager(ctx, config)
	if err != nil {
		log.Fatalf("Error creating manager: %v", err)
	}

	// Start the manager
	if err := manager.Start(); err != nil {
		log.Fatalf("Error starting manager: %v", err)
	}

	// Wait for interrupt signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)
	
	// Shutdown gracefully
	if err := manager.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	
	log.Printf("Shutdown complete")
}

// loadConfig loads the configuration from the specified file
func loadConfig(configPath string) (*Config, error) {
	// Create new viper instance
	v := viper.New()
	v.SetConfigFile(configPath)

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal config into struct
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return config, nil
}