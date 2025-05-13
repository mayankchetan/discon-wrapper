package main

// Config represents the configuration for the discon-manager
type Config struct {
	Server              ServerConfig              `mapstructure:"server"`
	Docker              DockerConfig              `mapstructure:"docker"`
	Database            DatabaseConfig            `mapstructure:"database"`
	ControllerDiscovery ControllerDiscoveryConfig `mapstructure:"controller_discovery"`
	Metrics             MetricsConfig             `mapstructure:"metrics"`
	Health              HealthConfig              `mapstructure:"health"`
	Auth                AuthConfig                `mapstructure:"auth"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port       int    `mapstructure:"port"`
	Host       string `mapstructure:"host"`
	DebugLevel int    `mapstructure:"debug_level"`
}

// DockerConfig represents the Docker configuration
type DockerConfig struct {
	NetworkName     string            `mapstructure:"network_name"`
	ContainerPrefix string            `mapstructure:"container_prefix"`
	MemoryLimit     string            `mapstructure:"memory_limit"`
	CPULimit        float64           `mapstructure:"cpu_limit"`
	CleanupTimeout  int               `mapstructure:"cleanup_timeout"`
	MountTimezone   bool              `mapstructure:"mount_timezone"`
	Environment     map[string]string `mapstructure:"environment"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// ControllerDiscoveryConfig represents controller discovery configuration
type ControllerDiscoveryConfig struct {
	Mode            string                     `mapstructure:"mode"` // "manual", "startup", or "periodic"
	IntervalMinutes int                        `mapstructure:"interval_minutes"`
	AutoRegister    bool                       `mapstructure:"auto_register"`
	RemoveMissing   bool                       `mapstructure:"remove_missing"`
	Validation      ControllerValidationConfig `mapstructure:"validation"`
}

// ControllerValidationConfig represents controller validation configuration
type ControllerValidationConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	VerifySymbols bool `mapstructure:"verify_symbols"`
	TestCall      bool `mapstructure:"test_call"`
}

// MetricsConfig represents the metrics configuration
type MetricsConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	Path               string `mapstructure:"path"`
	CollectionInterval int    `mapstructure:"collection_interval"`
}

// HealthConfig represents the health check configuration
type HealthConfig struct {
	Interval int `mapstructure:"interval"`
	Timeout  int `mapstructure:"timeout"`
}

// AuthConfig represents authentication configuration for admin access
type AuthConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Enabled  bool   `mapstructure:"enabled"`
}
