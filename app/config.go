package main

import (
	"os"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Server string `yaml:"server"`
	Port   int    `yaml:"port"`
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
