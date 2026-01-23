package main

import (
	"path/filepath"
	"testing"
)

// TestGetConfigPath tests the configuration path function
func TestGetConfigPath(t *testing.T) {
	path, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath() failed: %v", err)
	}

	if path == "" {
		t.Error("getConfigPath() returned empty path")
	}

	// Verify path contains expected components
	if !filepath.IsAbs(path) {
		t.Error("getConfigPath() should return absolute path")
	}

	// Verify it ends with config.yml
	if filepath.Base(path) != "config.yml" {
		t.Errorf("Expected config.yml, got %s", filepath.Base(path))
	}
}

// TestLoadConfig_Default tests loading default configuration
func TestLoadConfig_Default(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the config directory for testing
	originalConfigDirFunc := ConfigDirFunc
	ConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { ConfigDirFunc = originalConfigDirFunc }()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify default configuration
	if cfg.ActiveProfile != "Default" {
		t.Errorf("Expected ActiveProfile='Default', got '%s'", cfg.ActiveProfile)
	}

	if len(cfg.Profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(cfg.Profiles))
	}

	defaultProfile := cfg.Profiles[0]
	if defaultProfile.Name != "Default" {
		t.Errorf("Expected profile name='Default', got '%s'", defaultProfile.Name)
	}

	if defaultProfile.Server != "localhost" {
		t.Errorf("Expected server='localhost', got '%s'", defaultProfile.Server)
	}

	if defaultProfile.Port != 9191 {
		t.Errorf("Expected port=9191, got %d", defaultProfile.Port)
	}

	if defaultProfile.Theme != "dark" {
		t.Errorf("Expected theme='dark', got '%s'", defaultProfile.Theme)
	}

	if defaultProfile.Lang != "en" {
		t.Errorf("Expected language='en', got '%s'", defaultProfile.Lang)
	}

	if !defaultProfile.ShowTypes {
		t.Error("Expected ShowTypes=true")
	}
}

// TestSaveAndLoadConfig tests saving and loading configuration
func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the config directory for testing
	originalConfigDirFunc := ConfigDirFunc
	ConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { ConfigDirFunc = originalConfigDirFunc }()

	// Create test configuration
	testConfig := &Config{
		ActiveProfile: "TestProfile",
		Profiles: []Profile{
			{
				Name:      "TestProfile",
				Server:    "test.example.com",
				Port:      8080,
				Theme:     "light",
				Lang:      "es",
				ShowTypes: false,
				LogFolders: []LogFolder{
					{
						Path:       "/test/logs",
						Extensions: []string{"*.log", "*.txt"},
						Filters:    []string{"error", "warning"},
						Enabled:    true,
						Format:     "text",
					},
				},
			},
		},
	}

	// Save configuration
	err := SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Load configuration
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify loaded configuration matches saved configuration
	if loadedConfig.ActiveProfile != testConfig.ActiveProfile {
		t.Errorf("ActiveProfile mismatch: expected '%s', got '%s'",
			testConfig.ActiveProfile, loadedConfig.ActiveProfile)
	}

	if len(loadedConfig.Profiles) != len(testConfig.Profiles) {
		t.Fatalf("Profile count mismatch: expected %d, got %d",
			len(testConfig.Profiles), len(loadedConfig.Profiles))
	}

	savedProfile := testConfig.Profiles[0]
	loadedProfile := loadedConfig.Profiles[0]

	if loadedProfile.Name != savedProfile.Name {
		t.Errorf("Profile name mismatch: expected '%s', got '%s'",
			savedProfile.Name, loadedProfile.Name)
	}

	if loadedProfile.Server != savedProfile.Server {
		t.Errorf("Server mismatch: expected '%s', got '%s'",
			savedProfile.Server, loadedProfile.Server)
	}

	if loadedProfile.Port != savedProfile.Port {
		t.Errorf("Port mismatch: expected %d, got %d",
			savedProfile.Port, loadedProfile.Port)
	}

	if loadedProfile.Theme != savedProfile.Theme {
		t.Errorf("Theme mismatch: expected '%s', got '%s'",
			savedProfile.Theme, loadedProfile.Theme)
	}

	if loadedProfile.Lang != savedProfile.Lang {
		t.Errorf("Language mismatch: expected '%s', got '%s'",
			savedProfile.Lang, loadedProfile.Lang)
	}

	if loadedProfile.ShowTypes != savedProfile.ShowTypes {
		t.Errorf("ShowTypes mismatch: expected %v, got %v",
			savedProfile.ShowTypes, loadedProfile.ShowTypes)
	}

	// Verify log folders
	if len(loadedProfile.LogFolders) != len(savedProfile.LogFolders) {
		t.Fatalf("LogFolders count mismatch: expected %d, got %d",
			len(savedProfile.LogFolders), len(loadedProfile.LogFolders))
	}

	savedFolder := savedProfile.LogFolders[0]
	loadedFolder := loadedProfile.LogFolders[0]

	if loadedFolder.Path != savedFolder.Path {
		t.Errorf("LogFolder path mismatch: expected '%s', got '%s'",
			savedFolder.Path, loadedFolder.Path)
	}

	if loadedFolder.Format != savedFolder.Format {
		t.Errorf("LogFolder format mismatch: expected '%s', got '%s'",
			savedFolder.Format, loadedFolder.Format)
	}

	if loadedFolder.Enabled != savedFolder.Enabled {
		t.Errorf("LogFolder enabled mismatch: expected %v, got %v",
			savedFolder.Enabled, loadedFolder.Enabled)
	}
}

// TestGetActiveProfile tests the GetActiveProfile method
func TestGetActiveProfile(t *testing.T) {
	config := &Config{
		ActiveProfile: "Production",
		Profiles: []Profile{
			{Name: "Development", Server: "localhost", Port: 9191},
			{Name: "Production", Server: "prod.example.com", Port: 8080},
			{Name: "Testing", Server: "test.example.com", Port: 9090},
		},
	}

	activeProfile := config.GetActiveProfile()
	if activeProfile == nil {
		t.Fatal("GetActiveProfile() returned nil")
	}

	if activeProfile.Name != "Production" {
		t.Errorf("Expected active profile 'Production', got '%s'", activeProfile.Name)
	}

	if activeProfile.Server != "prod.example.com" {
		t.Errorf("Expected server 'prod.example.com', got '%s'", activeProfile.Server)
	}

	if activeProfile.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", activeProfile.Port)
	}
}

// TestGetActiveProfile_NotFound tests GetActiveProfile when active profile doesn't exist
func TestGetActiveProfile_NotFound(t *testing.T) {
	config := &Config{
		ActiveProfile: "NonExistent",
		Profiles: []Profile{
			{Name: "Development", Server: "localhost", Port: 9191},
			{Name: "Production", Server: "prod.example.com", Port: 8080},
		},
	}

	activeProfile := config.GetActiveProfile()
	if activeProfile == nil {
		t.Fatal("GetActiveProfile() should return first profile when active not found")
	}

	// Should return first profile as fallback
	if activeProfile.Name != "Development" {
		t.Errorf("Expected fallback to first profile 'Development', got '%s'", activeProfile.Name)
	}
}

// TestGetActiveProfile_EmptyProfiles tests GetActiveProfile with no profiles
func TestGetActiveProfile_EmptyProfiles(t *testing.T) {
	config := &Config{
		ActiveProfile: "Any",
		Profiles:      []Profile{},
	}

	activeProfile := config.GetActiveProfile()
	if activeProfile != nil {
		t.Error("GetActiveProfile() should return nil when no profiles exist")
	}
}

// TestMultipleProfiles tests configuration with multiple profiles
func TestMultipleProfiles(t *testing.T) {
	tempDir := t.TempDir()

	originalConfigDirFunc := ConfigDirFunc
	ConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { ConfigDirFunc = originalConfigDirFunc }()

	config := &Config{
		ActiveProfile: "Dev",
		Profiles: []Profile{
			{
				Name:      "Dev",
				Server:    "localhost",
				Port:      9191,
				Theme:     "dark",
				Lang:      "en",
				ShowTypes: true,
			},
			{
				Name:      "Staging",
				Server:    "staging.example.com",
				Port:      8080,
				Theme:     "light",
				Lang:      "es",
				ShowTypes: false,
			},
			{
				Name:      "Production",
				Server:    "prod.example.com",
				Port:      443,
				Theme:     "dark",
				Lang:      "en",
				ShowTypes: false,
			},
		},
	}

	// Save configuration
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Load configuration
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if len(loadedConfig.Profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(loadedConfig.Profiles))
	}

	activeProfile := loadedConfig.GetActiveProfile()
	if activeProfile.Name != "Dev" {
		t.Errorf("Expected active profile 'Dev', got '%s'", activeProfile.Name)
	}
}

// BenchmarkLoadConfig benchmarks loading configuration
func BenchmarkLoadConfig(b *testing.B) {
	tempDir := b.TempDir()

	originalConfigDirFunc := ConfigDirFunc
	ConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { ConfigDirFunc = originalConfigDirFunc }()

	// Create initial config
	testConfig := &Config{
		ActiveProfile: "Default",
		Profiles: []Profile{
			{
				Name:      "Default",
				Server:    "localhost",
				Port:      9191,
				Theme:     "dark",
				Lang:      "en",
				ShowTypes: true,
			},
		},
	}
	SaveConfig(testConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig()
		if err != nil {
			b.Fatalf("LoadConfig() failed: %v", err)
		}
	}
}

// BenchmarkSaveConfig benchmarks saving configuration
func BenchmarkSaveConfig(b *testing.B) {
	tempDir := b.TempDir()

	originalConfigDirFunc := ConfigDirFunc
	ConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { ConfigDirFunc = originalConfigDirFunc }()

	testConfig := &Config{
		ActiveProfile: "Default",
		Profiles: []Profile{
			{
				Name:      "Default",
				Server:    "localhost",
				Port:      9191,
				Theme:     "dark",
				Lang:      "en",
				ShowTypes: true,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := SaveConfig(testConfig)
		if err != nil {
			b.Fatalf("SaveConfig() failed: %v", err)
		}
	}
}
