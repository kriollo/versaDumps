package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Server    string `yaml:"server"`
	Port      int    `yaml:"port"`
	Theme     string `yaml:"theme,omitempty"`
	Lang      string `yaml:"language,omitempty"`
	ShowTypes bool   `yaml:"show_types,omitempty"`
}

// LoadConfig reads the configuration from config.yml
func LoadConfig() (*Config, error) {
	// Assumes config.yml is in the same directory as the executable
	// or in the project root during development.
	f, err := os.Open("config.yml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes the configuration to config.yml
func SaveConfig(cfg *Config) error {
	// Open file for write (truncate/create)
	f, err := os.Create("config.yml")
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
