package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LogFolder represents a folder to monitor for log files
type LogFolder struct {
	Path       string   `yaml:"path" json:"path"`
	Extensions []string `yaml:"extensions" json:"extensions"` // e.g., ["*.log", "*.txt"]
	Filters    []string `yaml:"filters" json:"filters"`       // e.g., ["error", "warning", "info"]
	Enabled    bool     `yaml:"enabled" json:"enabled"`
	Format     string   `yaml:"format,omitempty" json:"format,omitempty"` // "text" or "json"
}

// Profile represents a configuration profile
type Profile struct {
	Name       string      `yaml:"name" json:"name"`
	Server     string      `yaml:"server" json:"server"`
	Port       int         `yaml:"port" json:"port"`
	Theme      string      `yaml:"theme,omitempty" json:"theme,omitempty"`
	Lang       string      `yaml:"language,omitempty" json:"language,omitempty"`
	ShowTypes  bool        `yaml:"show_types,omitempty" json:"show_types,omitempty"`
	LogFolders []LogFolder `yaml:"log_folders,omitempty" json:"log_folders,omitempty"`
}

// WindowPosition stores window position and size
type WindowPosition struct {
	X      int `yaml:"x" json:"x"`
	Y      int `yaml:"y" json:"y"`
	Width  int `yaml:"width" json:"width"`
	Height int `yaml:"height" json:"height"`
}

// Config holds the application configuration with multiple profiles
type Config struct {
	ActiveProfile  string          `yaml:"active_profile" json:"active_profile"`
	Profiles       []Profile       `yaml:"profiles" json:"profiles"`
	WindowPosition *WindowPosition `yaml:"window_position,omitempty" json:"window_position,omitempty"`
}

// GetActiveProfile returns the currently active profile
func (c *Config) GetActiveProfile() *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Name == c.ActiveProfile {
			return &c.Profiles[i]
		}
	}
	// Return first profile if active not found
	if len(c.Profiles) > 0 {
		return &c.Profiles[0]
	}
	return nil
}

// ConfigDirFunc is a variable that can be overridden in tests
var ConfigDirFunc = os.UserConfigDir

// getConfigPath returns the path to the config file in the user's AppData directory
func getConfigPath() (string, error) {
	// Get user's config directory (AppData\Roaming on Windows)
	configDir, err := ConfigDirFunc()
	if err != nil {
		return "", err
	}

	// Create VersaDumps subdirectory
	appConfigDir := filepath.Join(configDir, "VersaDumps")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appConfigDir, "config.yml"), nil
}

// Legacy Config struct for backward compatibility
type LegacyConfig struct {
	Server    string `yaml:"server"`
	Port      int    `yaml:"port"`
	Theme     string `yaml:"theme,omitempty"`
	Lang      string `yaml:"language,omitempty"`
	ShowTypes bool   `yaml:"show_types,omitempty"`
}

// LoadConfig reads the configuration from config.yml
// If the file doesn't exist, it creates it with default values
// Supports backward compatibility with old config format
func LoadConfig() (*Config, error) {
	// Get config file path
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config.yml exists in new location
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try to migrate from old location (current directory / Program Files)
		oldConfigPath := "config.yml"
		if data, err := os.ReadFile(oldConfigPath); err == nil {
			// Old config exists, try to migrate it
			if err := os.WriteFile(configPath, data, 0644); err == nil {
				// Migration successful, optionally delete old file
				// We don't delete it to be safe, let user do it manually if needed
			}
		}

		// Check again after migration attempt
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Create default configuration with a default profile
			defaultConfig := &Config{
				ActiveProfile: "Default",
				Profiles: []Profile{
					{
						Name:       "Default",
						Server:     "localhost",
						Port:       9191,
						Theme:      "dark",
						Lang:       "en",
						ShowTypes:  true,
						LogFolders: []LogFolder{},
					},
				},
			}

			// Save the default configuration
			if err := SaveConfig(defaultConfig); err != nil {
				return nil, err
			}

			return defaultConfig, nil
		}
	}

	// File exists, read it
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Try to parse as new format first
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err == nil && len(cfg.Profiles) > 0 {
		// New format loaded successfully
		return &cfg, nil
	}

	// Try to parse as legacy format
	var legacyCfg LegacyConfig
	if err := yaml.Unmarshal(data, &legacyCfg); err == nil {
		// Convert legacy config to new format
		cfg = Config{
			ActiveProfile: "Default",
			Profiles: []Profile{
				{
					Name:       "Default",
					Server:     legacyCfg.Server,
					Port:       legacyCfg.Port,
					Theme:      legacyCfg.Theme,
					Lang:       legacyCfg.Lang,
					ShowTypes:  legacyCfg.ShowTypes,
					LogFolders: []LogFolder{},
				},
			},
		}

		// Save in new format
		if err := SaveConfig(&cfg); err != nil {
			return nil, err
		}

		return &cfg, nil
	}

	return nil, err
}

// SaveConfig writes the configuration to config.yml
func SaveConfig(cfg *Config) error {
	// Get config file path
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Open file for write (truncate/create)
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	defer encoder.Close()

	if err := encoder.Encode(cfg); err != nil {
		return err
	}
	return nil
}
