package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	CertificateSources []CertificateSource `mapstructure:"certificate_sources"`
	TrustStores        []TrustStore        `mapstructure:"trust_stores"`
	Settings           Settings            `mapstructure:"settings"`
}

// CertificateSource defines where to fetch new certificates from
type CertificateSource struct {
	Name        string            `mapstructure:"name"`
	Type        string            `mapstructure:"type"` // "url", "file", "directory"
	Source      string            `mapstructure:"source"`
	Enabled     bool              `mapstructure:"enabled"`
	Headers     map[string]string `mapstructure:"headers,omitempty"`
	VerifyTLS   bool              `mapstructure:"verify_tls"`
	Filters     []string          `mapstructure:"filters,omitempty"`
}

// TrustStore defines a target trust store to update
type TrustStore struct {
	Name        string            `mapstructure:"name"`
	Type        string            `mapstructure:"type"` // "system", "application"
	Platform    []string          `mapstructure:"platform"` // ["linux", "darwin", "windows"]
	Target      string            `mapstructure:"target"` // specific store identifier
	Enabled     bool              `mapstructure:"enabled"`
	Options     map[string]string `mapstructure:"options,omitempty"`
	RequireRoot bool              `mapstructure:"require_root"`
}

// Settings contains global application settings
type Settings struct {
	BackupEnabled    bool   `mapstructure:"backup_enabled"`
	BackupDirectory  string `mapstructure:"backup_directory"`
	LogLevel         string `mapstructure:"log_level"`
	MaxRetries       int    `mapstructure:"max_retries"`
	TimeoutSeconds   int    `mapstructure:"timeout_seconds"`
	ValidateAfter    bool   `mapstructure:"validate_after"`
}

var globalConfig *Config

// InitConfig initializes the configuration with the given config file path
func InitConfig(cfgFile string) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName("trust-store-config")
		viper.SetConfigType("yaml")
	}

	// Set defaults
	setDefaults()

	// Read environment variables with TSU_ prefix
	viper.SetEnvPrefix("TSU")
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create a default one
			createDefaultConfig()
		}
	}
}

// LoadConfig loads and returns the configuration
func LoadConfig() (*Config, error) {
	if globalConfig == nil {
		var cfg Config
		if err := viper.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		globalConfig = &cfg
	}
	return globalConfig, nil
}

func setDefaults() {
	viper.SetDefault("settings.backup_enabled", true)
	viper.SetDefault("settings.backup_directory", "./backups")
	viper.SetDefault("settings.log_level", "info")
	viper.SetDefault("settings.max_retries", 3)
	viper.SetDefault("settings.timeout_seconds", 30)
	viper.SetDefault("settings.validate_after", true)
}

func createDefaultConfig() {
	configPath := "./trust-store-config.yaml"
	
	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return
	}

	defaultConfig := `# Trust Store Updater Configuration
# This file defines certificate sources and target trust stores to update

# Certificate sources - where to fetch new root certificates from
certificate_sources:
  - name: "mozilla-ca-bundle"
    type: "url"
    source: "https://curl.se/ca/cacert.pem"
    enabled: true
    verify_tls: true
    filters: []

  - name: "local-certificates"
    type: "directory"
    source: "./certificates"
    enabled: false
    filters:
      - "*.crt"
      - "*.pem"

# Trust stores - target stores to update with new certificates
trust_stores:
  # System trust stores
  - name: "system-ca-certificates"
    type: "system"
    platform: ["linux"]
    target: "ca-certificates"
    enabled: true
    require_root: true

  - name: "system-keychain"
    type: "system"
    platform: ["darwin"]
    target: "system-keychain"
    enabled: true
    require_root: true

  - name: "system-cert-store"
    type: "system"
    platform: ["windows"]
    target: "root"
    enabled: true
    require_root: true

  # Application trust stores
  - name: "docker-ca-certificates"
    type: "application"
    platform: ["linux", "darwin", "windows"]
    target: "docker"
    enabled: false
    require_root: false

  - name: "java-cacerts"
    type: "application"
    platform: ["linux", "darwin", "windows"]
    target: "java-cacerts"
    enabled: false
    require_root: false

# Global settings
settings:
  backup_enabled: true
  backup_directory: "./backups"
  log_level: "info"
  max_retries: 3
  timeout_seconds: 30
  validate_after: true
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err == nil {
		fmt.Printf("Created default configuration file: %s\n", configPath)
		fmt.Println("Please review and customize the configuration before running the updater.")
	}
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	return viper.ConfigFileUsed()
}

// ValidateConfig validates the loaded configuration
func ValidateConfig(cfg *Config) error {
	if len(cfg.CertificateSources) == 0 {
		return fmt.Errorf("no certificate sources configured")
	}

	if len(cfg.TrustStores) == 0 {
		return fmt.Errorf("no trust stores configured")
	}

	// Validate backup directory
	if cfg.Settings.BackupEnabled {
		if cfg.Settings.BackupDirectory == "" {
			return fmt.Errorf("backup directory must be specified when backup is enabled")
		}
		
		// Create backup directory if it doesn't exist
		if err := os.MkdirAll(cfg.Settings.BackupDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create backup directory: %w", err)
		}
	}

	return nil
}
