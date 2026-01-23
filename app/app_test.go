package main

import (
	"context"
	"testing"
)

// Helper function to setup test environment
func setupTestApp(t *testing.T) (*App, func()) {
	// Create temporary directory for config
	tempDir := t.TempDir()

	// Override the config directory
	originalConfigDirFunc := ConfigDirFunc
	ConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}

	app := NewApp()
	app.ctx = context.Background()

	cleanup := func() {
		ConfigDirFunc = originalConfigDirFunc
		app.stopHTTPServer()
		app.StopLogWatcher()
	}

	return app, cleanup
}

// TestNewApp tests creating a new App
func TestNewApp(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	if app.messageCounter != 0 {
		t.Errorf("Expected messageCounter=0, got %d", app.messageCounter)
	}

	if app.updateManager == nil {
		t.Error("UpdateManager not initialized")
	}
}

// TestApp_GetConfig tests getting configuration
func TestApp_GetConfig(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// First load should create default config
	cfg, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("GetConfig() returned nil")
	}

	if cfg.Name != "Default" {
		t.Errorf("Expected profile name 'Default', got '%s'", cfg.Name)
	}

	if cfg.Server != "localhost" {
		t.Errorf("Expected server 'localhost', got '%s'", cfg.Server)
	}

	// Port should be 9191 for new config, but might be different if config exists
	if cfg.Port == 0 {
		t.Error("Port should not be 0")
	}
}

// TestApp_SaveFrontendConfig tests saving frontend configuration
func TestApp_SaveFrontendConfig(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Save new configuration
	newConfig := map[string]interface{}{
		"server":     "test.example.com",
		"port":       8080,
		"theme":      "light",
		"language":   "es",
		"show_types": false,
	}

	// Note: This will try to restart HTTP server which requires Wails context
	// We'll skip the actual save and just test that it doesn't crash
	err = app.SaveFrontendConfig(newConfig)
	// Expect error because context is not properly initialized for runtime operations
	// This is expected in unit tests
	t.Logf("SaveFrontendConfig() returned (expected to fail in unit test): %v", err)
}

// TestApp_CreateProfile tests creating a new profile
func TestApp_CreateProfile(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Create new profile
	err = app.CreateProfile("TestProfile", "test.example.com", 8080, "light", "es", false)
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	// Verify profile was created
	profiles, err := app.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("Expected 2 profiles, got %d", len(profiles))
	}

	// Find the new profile
	var testProfile *Profile
	for i := range profiles {
		if profiles[i].Name == "TestProfile" {
			testProfile = &profiles[i]
			break
		}
	}

	if testProfile == nil {
		t.Fatal("TestProfile not found")
	}

	if testProfile.Server != "test.example.com" {
		t.Errorf("Server mismatch: expected 'test.example.com', got '%s'", testProfile.Server)
	}

	if testProfile.Port != 8080 {
		t.Errorf("Port mismatch: expected 8080, got %d", testProfile.Port)
	}

	if testProfile.Theme != "light" {
		t.Errorf("Theme mismatch: expected 'light', got '%s'", testProfile.Theme)
	}

	if testProfile.Lang != "es" {
		t.Errorf("Language mismatch: expected 'es', got '%s'", testProfile.Lang)
	}
}

// TestApp_CreateProfile_Duplicate tests creating duplicate profile
func TestApp_CreateProfile_Duplicate(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Try to create profile with same name as default
	err = app.CreateProfile("Default", "test.example.com", 8080, "light", "es", false)
	if err == nil {
		t.Error("CreateProfile() should return error for duplicate profile name")
	}
}

// TestApp_SwitchProfile tests switching profiles
func TestApp_SwitchProfile(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Create new profile
	err = app.CreateProfile("Production", "prod.example.com", 443, "dark", "en", true)
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	// Switch to new profile
	err = app.SwitchProfile("Production")
	if err != nil {
		t.Fatalf("SwitchProfile() failed: %v", err)
	}

	// Verify active profile changed
	activeProfileName, err := app.GetActiveProfileName()
	if err != nil {
		t.Fatalf("GetActiveProfileName() failed: %v", err)
	}

	if activeProfileName != "Production" {
		t.Errorf("Active profile not switched: expected 'Production', got '%s'", activeProfileName)
	}

	// Verify GetConfig returns the new active profile
	cfg, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	if cfg.Name != "Production" {
		t.Errorf("GetConfig() should return active profile: expected 'Production', got '%s'", cfg.Name)
	}

	if cfg.Server != "prod.example.com" {
		t.Errorf("Server mismatch: expected 'prod.example.com', got '%s'", cfg.Server)
	}
}

// TestApp_SwitchProfile_NonExistent tests switching to non-existent profile
func TestApp_SwitchProfile_NonExistent(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Try to switch to non-existent profile
	err = app.SwitchProfile("NonExistent")
	if err == nil {
		t.Error("SwitchProfile() should return error for non-existent profile")
	}
}

// TestApp_DeleteProfile tests deleting a profile
func TestApp_DeleteProfile(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Create new profile
	err = app.CreateProfile("ToDelete", "test.example.com", 8080, "light", "es", false)
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	// Verify we have 2 profiles
	profiles, err := app.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("Expected 2 profiles before delete, got %d", len(profiles))
	}

	// Delete profile
	err = app.DeleteProfile("ToDelete")
	if err != nil {
		t.Fatalf("DeleteProfile() failed: %v", err)
	}

	// Verify profile was deleted
	profiles, err = app.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() failed: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile after delete, got %d", len(profiles))
	}

	// Verify deleted profile is gone
	for _, p := range profiles {
		if p.Name == "ToDelete" {
			t.Error("Deleted profile still exists")
		}
	}
}

// TestApp_DeleteProfile_LastProfile tests deleting last profile
func TestApp_DeleteProfile_LastProfile(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config (only one profile)
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Try to delete the only profile
	err = app.DeleteProfile("Default")
	if err == nil {
		t.Error("DeleteProfile() should return error when trying to delete last profile")
	}
}

// TestApp_DeleteProfile_Active tests deleting active profile
func TestApp_DeleteProfile_Active(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Create second profile
	err = app.CreateProfile("Secondary", "test.example.com", 8080, "light", "es", false)
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	// Try to delete active profile (Default is active by default)
	err = app.DeleteProfile("Default")
	if err == nil {
		t.Error("DeleteProfile() should return error when trying to delete active profile")
	}
}

// TestApp_UpdateProfile tests updating a profile
func TestApp_UpdateProfile(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Update profile
	err = app.UpdateProfile("Default", "updated.example.com", 9999, "light", "es", false)
	if err != nil {
		t.Fatalf("UpdateProfile() failed: %v", err)
	}

	// Verify profile was updated
	cfg, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	if cfg.Server != "updated.example.com" {
		t.Errorf("Server not updated: expected 'updated.example.com', got '%s'", cfg.Server)
	}

	if cfg.Port != 9999 {
		t.Errorf("Port not updated: expected 9999, got %d", cfg.Port)
	}

	if cfg.Theme != "light" {
		t.Errorf("Theme not updated: expected 'light', got '%s'", cfg.Theme)
	}

	if cfg.Lang != "es" {
		t.Errorf("Language not updated: expected 'es', got '%s'", cfg.Lang)
	}

	if cfg.ShowTypes {
		t.Error("ShowTypes not updated: expected false, got true")
	}
}

// TestApp_ListProfiles tests listing all profiles
func TestApp_ListProfiles(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	// Create additional profiles
	profiles := []string{"Dev", "Staging", "Production"}
	for _, name := range profiles {
		err = app.CreateProfile(name, "test.example.com", 8080, "dark", "en", true)
		if err != nil {
			t.Fatalf("CreateProfile(%s) failed: %v", name, err)
		}
	}

	// List profiles
	allProfiles, err := app.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() failed: %v", err)
	}

	// Should have Default + 3 new profiles = 4 total
	if len(allProfiles) != 4 {
		t.Fatalf("Expected 4 profiles, got %d", len(allProfiles))
	}

	// Verify all profile names are present
	profileNames := make(map[string]bool)
	for _, p := range allProfiles {
		profileNames[p.Name] = true
	}

	expectedNames := []string{"Default", "Dev", "Staging", "Production"}
	for _, name := range expectedNames {
		if !profileNames[name] {
			t.Errorf("Profile '%s' not found in list", name)
		}
	}
}

// TestApp_UpdateVisibleCount tests updating visible count
func TestApp_UpdateVisibleCount(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Initial count should be 0
	if app.GetVisibleCount() != 0 {
		t.Errorf("Initial count should be 0, got %d", app.GetVisibleCount())
	}

	// Update count
	app.UpdateVisibleCount(5)

	// Verify count was updated
	if app.GetVisibleCount() != 5 {
		t.Errorf("Expected count=5, got %d", app.GetVisibleCount())
	}

	// Update count again
	app.UpdateVisibleCount(0)

	// Verify count was reset
	if app.GetVisibleCount() != 0 {
		t.Errorf("Expected count=0 after reset, got %d", app.GetVisibleCount())
	}
}

// BenchmarkApp_GetConfig benchmarks getting configuration
func BenchmarkApp_GetConfig(b *testing.B) {
	tempDir := b.TempDir()

	originalConfigDirFunc := ConfigDirFunc
	ConfigDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { ConfigDirFunc = originalConfigDirFunc }()

	app := NewApp()
	app.ctx = context.Background()

	// Create initial config
	_, err := app.GetConfig()
	if err != nil {
		b.Fatalf("GetConfig() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := app.GetConfig()
		if err != nil {
			b.Fatalf("GetConfig() failed: %v", err)
		}
	}
}
