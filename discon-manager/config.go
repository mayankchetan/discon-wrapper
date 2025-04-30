package main

// Config represents the configuration for the discon-manager
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Docker   DockerConfig   `mapstructure:"docker"`
	Database DatabaseConfig `mapstructure:"database"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Health   HealthConfig   `mapstructure:"health"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port       int    `mapstructure:"port"`
	Host       string `mapstructure:"host"`
	DebugLevel int    `mapstructure:"debug_level"`
}

// DockerConfig represents the Docker configuration
type DockerConfig struct {
	NetworkName     string  `mapstructure:"network_name"`
	ContainerPrefix string  `mapstructure:"container_prefix"`
	MemoryLimit     string  `mapstructure:"memory_limit"`
	CPULimit        float64 `mapstructure:"cpu_limit"`
	CleanupTimeout  int     `mapstructure:"cleanup_timeout"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// MetricsConfig represents the metrics configuration
type MetricsConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	Path              string `mapstructure:"path"`
	CollectionInterval int    `mapstructure:"collection_interval"`
}

// HealthConfig represents the health check configuration
type HealthConfig struct {
	Interval int `mapstructure:"interval"`
	Timeout  int `mapstructure:"timeout"`
}