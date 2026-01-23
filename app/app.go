package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	gosys "runtime"
	"strconv"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	messageCounter int
	updateManager  *UpdateManager
	httpServer     *http.Server
	serverCancel   context.CancelFunc
	logWatcher     *LogWatcher
	serverMu       sync.Mutex // Protect server start/stop operations
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		updateManager: NewUpdateManager(),
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load config (will create with defaults if it doesn't exist)
	cfg, err := LoadConfig()
	if err != nil {
		runtime.LogErrorf(ctx, "Failed to load or create config: %v", err)
		return
	}

	// Get config path for logging
	configPath, _ := getConfigPath()
	runtime.LogInfof(ctx, "Using config file: %s", configPath)

	// Get active profile
	activeProfile := cfg.GetActiveProfile()
	if activeProfile == nil {
		runtime.LogErrorf(ctx, "No active profile found")
		return
	}

	// Emit loaded config to frontend so it can initialize theme/language
	// Send the active profile as the config
	cfgBytes, _ := json.Marshal(activeProfile)
	runtime.EventsEmit(ctx, "configLoaded", string(cfgBytes))

	// Start the background HTTP server with active profile settings
	runtime.LogInfof(ctx, "════════════════════════════════════════")
	runtime.LogInfof(ctx, "STARTING HTTP SERVER WITH CONFIG:")
	runtime.LogInfof(ctx, "Active Profile: %s", activeProfile.Name)
	runtime.LogInfof(ctx, "Server: %s", activeProfile.Server)
	runtime.LogInfof(ctx, "Port: %d", activeProfile.Port)
	runtime.LogInfof(ctx, "════════════════════════════════════════")
	a.startHTTPServer(activeProfile.Server, activeProfile.Port)

	// Start log watcher if there are log folders configured
	if len(activeProfile.LogFolders) > 0 {
		if err := a.StartLogWatcher(); err != nil {
			runtime.LogErrorf(ctx, "Failed to start log watcher: %v", err)
		}
	}

	// Initialize window title with current counter
	if a.ctx != nil {
		runtime.WindowSetTitle(a.ctx, fmt.Sprintf("VersaDumps Visualizer (%d)", a.messageCounter))
	}
}

// GetVisibleCount returns the current visible message count
func (a *App) GetVisibleCount() int {
	return a.messageCounter
}

// GetConfig returns the active profile configuration (callable from frontend)
// Returns the active profile as a Profile struct for backward compatibility
func (a *App) GetConfig() (*Profile, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	activeProfile := cfg.GetActiveProfile()
	if activeProfile == nil {
		return nil, fmt.Errorf("no active profile found")
	}

	return activeProfile, nil
}

// SaveFrontendConfig accepts partial configuration from frontend and persists it to the active profile
func (a *App) SaveFrontendConfig(partial map[string]interface{}) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Get active profile
	activeProfile := cfg.GetActiveProfile()
	if activeProfile == nil {
		return fmt.Errorf("no active profile found")
	}

	// Find the index of the active profile to update it
	profileIndex := -1
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == activeProfile.Name {
			profileIndex = i
			break
		}
	}

	if profileIndex == -1 {
		return fmt.Errorf("active profile not found in profiles list")
	}

	// Handle server field
	if v, ok := partial["server"]; ok {
		if serverStr, ok := v.(string); ok && serverStr != "" {
			cfg.Profiles[profileIndex].Server = serverStr
		}
	}

	// Handle port field
	if v, ok := partial["port"]; ok {
		switch port := v.(type) {
		case float64:
			cfg.Profiles[profileIndex].Port = int(port)
		case int:
			cfg.Profiles[profileIndex].Port = port
		case string:
			// Try to parse string to int
			if portInt, err := strconv.Atoi(port); err == nil {
				cfg.Profiles[profileIndex].Port = portInt
			}
		}
	}

	// Handle theme field
	if v, ok := partial["theme"]; ok {
		if themeStr, ok := v.(string); ok {
			cfg.Profiles[profileIndex].Theme = themeStr
		}
	}

	// Handle language field
	if v, ok := partial["language"]; ok {
		if langStr, ok := v.(string); ok {
			cfg.Profiles[profileIndex].Lang = langStr
		}
	}

	// Handle show_types field
	if v, ok := partial["show_types"]; ok {
		switch showTypes := v.(type) {
		case bool:
			cfg.Profiles[profileIndex].ShowTypes = showTypes
		case string:
			cfg.Profiles[profileIndex].ShowTypes = (showTypes == "true")
		}
	}

	// Save the configuration
	err = SaveConfig(cfg)
	if err != nil {
		return err
	}

	// Restart the HTTP server to apply new settings (synchronously to avoid race conditions)
	runtime.LogInfof(a.ctx, "Configuration saved, restarting HTTP server...")
	if err := a.RestartHTTPServer(); err != nil {
		runtime.LogErrorf(a.ctx, "Error restarting HTTP server: %v", err)
	}

	return nil
}

// UpdateVisibleCount updates the internal counter and window title based on the
// number of messages currently visible in the UI. This should be called by the frontend
// whenever logs are added/removed/cleared so the title and any OS badges remain in sync.
func (a *App) UpdateVisibleCount(count int) {
	runtime.LogInfof(a.ctx, "UpdateVisibleCount called: old=%d, new=%d", a.messageCounter, count)
	a.messageCounter = count
	if a.ctx != nil {
		newTitle := fmt.Sprintf("VersaDumps Visualizer (%d)", a.messageCounter)
		runtime.WindowSetTitle(a.ctx, newTitle)
		runtime.LogInfof(a.ctx, "Setting window title to: %s", newTitle)
	}
	// Update platform-specific taskbar/tray badge if available
	SetTaskbarBadge(a.ctx, a.messageCounter)
}

// OpenInEditor attempts to open a file at a specific line in the user's editor.
// It prefers VS Code (code -g file:line) if available, otherwise falls back to
// the platform default opener (start/open/xdg-open).
func (a *App) OpenInEditor(path string, line int) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	// Prefer VS Code if available
	// code -g file:line
	codeCmd := "code"
	arg := fmt.Sprintf("%s:%d", path, line)
	if _, err := exec.LookPath(codeCmd); err == nil {
		cmd := exec.Command(codeCmd, "-g", arg)
		return cmd.Start()
	}

	// Fallbacks by OS
	switch gosys.GOOS {
	case "windows":
		// start requires cmd /c start, but exec.Command can call 'cmd' directly
		cmd := exec.Command("cmd", "/C", "start", "", path)
		return cmd.Start()
	case "darwin":
		cmd := exec.Command("open", path)
		return cmd.Start()
	default:
		// Linux and others
		cmd := exec.Command("xdg-open", path)
		return cmd.Start()
	}
}

// CheckForUpdates verifica si hay una nueva versión disponible
func (a *App) CheckForUpdates() (*UpdateInfo, error) {
	return a.updateManager.CheckForUpdates()
}

// TestUpdateCheck - método de prueba para verificar actualizaciones
func (a *App) TestUpdateCheck() (*UpdateInfo, error) {
	// Usar el UpdateManager para obtener la comparación real
	actualInfo, err := a.updateManager.CheckForUpdates()
	if err != nil {
		// Si hay error (como rate limiting), simular que no hay actualizaciones
		return &UpdateInfo{
			Available:      false,
			Version:        CurrentVersion,
			Description:    "Tu versión de VersaDumps está actualizada.",
			DownloadURL:    "",
			ReleaseURL:     "https://github.com/kriollo/versaDumps/releases",
			Size:           0,
			CurrentVersion: CurrentVersion,
		}, nil
	}

	return actualInfo, nil
}

// DownloadAndInstallUpdate descarga e instala la actualización
func (a *App) DownloadAndInstallUpdate(downloadURL string) error {
	// Mostrar diálogo de progreso
	runtime.EventsEmit(a.ctx, "updateDownloadProgress", map[string]interface{}{
		"status":   "starting",
		"progress": 0,
	})

	// Descargar con callback de progreso
	filePath, err := a.updateManager.DownloadUpdate(downloadURL, func(downloaded, total int64) {
		if total > 0 {
			progress := float64(downloaded) / float64(total) * 100
			runtime.EventsEmit(a.ctx, "updateDownloadProgress", map[string]interface{}{
				"status":     "downloading",
				"progress":   progress,
				"downloaded": downloaded,
				"total":      total,
			})
		}
	})

	if err != nil {
		runtime.EventsEmit(a.ctx, "updateDownloadProgress", map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return err
	}

	// Notificar que la descarga se completó
	runtime.EventsEmit(a.ctx, "updateDownloadProgress", map[string]interface{}{
		"status":   "installing",
		"progress": 100,
	})

	// Instalar la actualización
	if err := a.updateManager.InstallUpdate(filePath); err != nil {
		runtime.EventsEmit(a.ctx, "updateDownloadProgress", map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return err
	}

	// Notificar éxito
	runtime.EventsEmit(a.ctx, "updateDownloadProgress", map[string]interface{}{
		"status": "complete",
	})

	return nil
}

// GetCurrentVersion retorna la versión actual de la aplicación
func (a *App) GetCurrentVersion() string {
	return CurrentVersion
}

// startHTTPServer starts the HTTP server and stores its reference
func (a *App) startHTTPServer(host string, port int) {
	a.serverMu.Lock()
	defer a.serverMu.Unlock()

	// Stop existing server if running
	a.stopHTTPServerInternal()

	// Create context for the server
	ctx, cancel := context.WithCancel(a.ctx)
	a.serverCancel = cancel

	// Start server (StartServer now manages its own goroutine)
	a.httpServer = StartServer(ctx, host, port, a)

	runtime.LogInfof(a.ctx, "HTTP server started successfully")
}

// stopHTTPServer stops the current HTTP server
func (a *App) stopHTTPServer() {
	a.serverMu.Lock()
	defer a.serverMu.Unlock()
	a.stopHTTPServerInternal()
}

// stopHTTPServerInternal stops the server without locking (internal use only)
func (a *App) stopHTTPServerInternal() {
	if a.serverCancel != nil {
		runtime.LogInfof(a.ctx, "Stopping HTTP server...")
		a.serverCancel()
		// Wait a moment for graceful shutdown
		a.serverCancel = nil
		a.httpServer = nil
	}
}

// RestartHTTPServer restarts the HTTP server with new configuration
func (a *App) RestartHTTPServer() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	activeProfile := cfg.GetActiveProfile()
	if activeProfile != nil {
		runtime.LogInfof(a.ctx, "Restarting HTTP server with new config: %s:%d", activeProfile.Server, activeProfile.Port)
		a.startHTTPServer(activeProfile.Server, activeProfile.Port)
	}

	return nil
}

// ========================================
// Profile Management Functions
// ========================================

// ListProfiles returns all available profiles
func (a *App) ListProfiles() ([]Profile, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return cfg.Profiles, nil
}

// GetActiveProfileName returns the name of the active profile
func (a *App) GetActiveProfileName() (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}
	return cfg.ActiveProfile, nil
}

// CreateProfile creates a new profile with the given configuration
func (a *App) CreateProfile(name string, server string, port int, theme string, lang string, showTypes bool) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Check if profile with same name exists
	for _, p := range cfg.Profiles {
		if p.Name == name {
			return fmt.Errorf("profile with name '%s' already exists", name)
		}
	}

	// Create new profile
	newProfile := Profile{
		Name:       name,
		Server:     server,
		Port:       port,
		Theme:      theme,
		Lang:       lang,
		ShowTypes:  showTypes,
		LogFolders: []LogFolder{},
	}

	cfg.Profiles = append(cfg.Profiles, newProfile)

	return SaveConfig(cfg)
}

// DeleteProfile deletes a profile by name
func (a *App) DeleteProfile(name string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Don't allow deleting if it's the only profile
	if len(cfg.Profiles) <= 1 {
		return fmt.Errorf("cannot delete the last profile")
	}

	// Don't allow deleting active profile without switching first
	if cfg.ActiveProfile == name {
		return fmt.Errorf("cannot delete active profile, switch to another profile first")
	}

	// Find and remove profile
	newProfiles := []Profile{}
	found := false
	for _, p := range cfg.Profiles {
		if p.Name != name {
			newProfiles = append(newProfiles, p)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("profile '%s' not found", name)
	}

	cfg.Profiles = newProfiles
	return SaveConfig(cfg)
}

// SwitchProfile changes the active profile
func (a *App) SwitchProfile(name string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Check if profile exists
	var newProfile *Profile
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == name {
			newProfile = &cfg.Profiles[i]
			break
		}
	}

	if newProfile == nil {
		return fmt.Errorf("profile '%s' not found", name)
	}

	cfg.ActiveProfile = name

	if err := SaveConfig(cfg); err != nil {
		return err
	}

	// Emit the new active profile to frontend so it can update immediately
	cfgBytes, _ := json.Marshal(newProfile)
	runtime.EventsEmit(a.ctx, "profileSwitched", string(cfgBytes))

	// Restart HTTP server with new profile settings
	if err := a.RestartHTTPServer(); err != nil {
		return err
	}

	// Restart log watcher with new profile's log folders
	if len(newProfile.LogFolders) > 0 {
		if err := a.RestartLogWatcher(); err != nil {
			runtime.LogErrorf(a.ctx, "Error restarting log watcher after profile switch: %v", err)
		}
	} else {
		// Stop log watcher if new profile has no log folders
		a.StopLogWatcher()
	}

	return nil
}

// UpdateProfile updates an existing profile
func (a *App) UpdateProfile(name string, server string, port int, theme string, lang string, showTypes bool) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Find and update profile
	found := false
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == name {
			cfg.Profiles[i].Server = server
			cfg.Profiles[i].Port = port
			cfg.Profiles[i].Theme = theme
			cfg.Profiles[i].Lang = lang
			cfg.Profiles[i].ShowTypes = showTypes
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile '%s' not found", name)
	}

	if err := SaveConfig(cfg); err != nil {
		return err
	}

	// If this is the active profile, restart server
	if cfg.ActiveProfile == name {
		return a.RestartHTTPServer()
	}

	return nil
}

// AddLogFolder adds a log folder to a profile
func (a *App) AddLogFolder(profileName string, path string, extensions []string, filters []string, format string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Validate that the path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("the path '%s' does not exist", path)
		}
		return fmt.Errorf("error accessing path '%s': %v", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", path)
	}

	// Find profile
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == profileName {
			// Check if folder already exists
			for _, lf := range cfg.Profiles[i].LogFolders {
				if lf.Path == path {
					return fmt.Errorf("folder '%s' already exists in profile", path)
				}
			}

			// Default format to "text" if not specified
			if format == "" {
				format = "text"
			}

			// Add folder
			newFolder := LogFolder{
				Path:       path,
				Extensions: extensions,
				Filters:    filters,
				Enabled:    true,
				Format:     format,
			}
			cfg.Profiles[i].LogFolders = append(cfg.Profiles[i].LogFolders, newFolder)

			if err := SaveConfig(cfg); err != nil {
				return err
			}

			// Restart log watcher if this is the active profile
			if cfg.ActiveProfile == profileName {
				runtime.LogInfof(a.ctx, "Restarting log watcher after adding folder")
				if err := a.RestartLogWatcher(); err != nil {
					runtime.LogErrorf(a.ctx, "Error restarting log watcher: %v", err)
				}
			}

			return nil
		}
	}

	return fmt.Errorf("profile '%s' not found", profileName)
}

// RemoveLogFolder removes a log folder from a profile
func (a *App) RemoveLogFolder(profileName string, path string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Find profile
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == profileName {
			// Find and remove folder
			newFolders := []LogFolder{}
			for _, lf := range cfg.Profiles[i].LogFolders {
				if lf.Path != path {
					newFolders = append(newFolders, lf)
				}
			}
			cfg.Profiles[i].LogFolders = newFolders

			if err := SaveConfig(cfg); err != nil {
				return err
			}

			// Restart log watcher if this is the active profile
			if cfg.ActiveProfile == profileName {
				runtime.LogInfof(a.ctx, "Restarting log watcher after removing folder")
				if err := a.RestartLogWatcher(); err != nil {
					runtime.LogErrorf(a.ctx, "Error restarting log watcher: %v", err)
				}
			}

			return nil
		}
	}

	return fmt.Errorf("profile '%s' not found", profileName)
}

// ToggleLogFolder enables or disables a log folder
func (a *App) ToggleLogFolder(profileName string, path string, enabled bool) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Find profile and folder
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == profileName {
			for j := range cfg.Profiles[i].LogFolders {
				if cfg.Profiles[i].LogFolders[j].Path == path {
					cfg.Profiles[i].LogFolders[j].Enabled = enabled

					if err := SaveConfig(cfg); err != nil {
						return err
					}

					// Restart log watcher if this is the active profile
					if cfg.ActiveProfile == profileName {
						return a.RestartLogWatcher()
					}
					return nil
				}
			}
		}
	}

	return fmt.Errorf("folder not found")
}

// UpdateLogFolder updates the configuration of an existing log folder
func (a *App) UpdateLogFolder(profileName string, path string, extensions []string, filters []string, format string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Find profile and folder
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == profileName {
			for j := range cfg.Profiles[i].LogFolders {
				if cfg.Profiles[i].LogFolders[j].Path == path {
					// Default format to "text" if not specified
					if format == "" {
						format = "text"
					}

					// Update extensions, filters, and format
					cfg.Profiles[i].LogFolders[j].Extensions = extensions
					cfg.Profiles[i].LogFolders[j].Filters = filters
					cfg.Profiles[i].LogFolders[j].Format = format

					if err := SaveConfig(cfg); err != nil {
						return err
					}

					// Restart log watcher if this is the active profile
					if cfg.ActiveProfile == profileName {
						runtime.LogInfof(a.ctx, "Restarting log watcher after updating folder")
						if err := a.RestartLogWatcher(); err != nil {
							runtime.LogErrorf(a.ctx, "Error restarting log watcher: %v", err)
						}
					}

					return nil
				}
			}
			return fmt.Errorf("folder '%s' not found in profile", path)
		}
	}

	return fmt.Errorf("profile '%s' not found", profileName)
}

// ========================================
// Log Watcher Control Functions
// ========================================

// StartLogWatcher starts monitoring log folders for the active profile
func (a *App) StartLogWatcher() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	activeProfile := cfg.GetActiveProfile()
	if activeProfile == nil {
		return fmt.Errorf("no active profile")
	}

	// Stop existing watcher if running
	if a.logWatcher != nil {
		a.logWatcher.Stop()
	}

	// Create new watcher
	watcher, err := NewLogWatcher(a.ctx)
	if err != nil {
		return err
	}

	a.logWatcher = watcher

	// Start monitoring folders from active profile
	if len(activeProfile.LogFolders) > 0 {
		return a.logWatcher.Start(activeProfile.LogFolders)
	}

	return nil
}

// StopLogWatcher stops the log watcher
func (a *App) StopLogWatcher() {
	if a.logWatcher != nil {
		a.logWatcher.Stop()
		a.logWatcher = nil
	}
}

// RestartLogWatcher restarts the log watcher with current configuration
func (a *App) RestartLogWatcher() error {
	a.StopLogWatcher()
	return a.StartLogWatcher()
}

// GetLogFolders returns the log folders for the active profile
func (a *App) GetLogFolders() ([]LogFolder, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	activeProfile := cfg.GetActiveProfile()
	if activeProfile == nil {
		return nil, fmt.Errorf("no active profile")
	}

	return activeProfile.LogFolders, nil
}

// SelectFolder opens a folder selection dialog
func (a *App) SelectFolder() (string, error) {
	folder, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Log Folder",
	})
	return folder, err
}

// SaveWindowPosition saves the current window position and size to config
func (a *App) SaveWindowPosition() error {
	// Get current window position and size
	x, y := runtime.WindowGetPosition(a.ctx)
	width, height := runtime.WindowGetSize(a.ctx)

	// Load current config
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Update window position
	cfg.WindowPosition = &WindowPosition{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}

	// Save config
	return SaveConfig(cfg)
}

// GetWindowPosition returns the saved window position from config
func (a *App) GetWindowPosition() (*WindowPosition, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	return cfg.WindowPosition, nil
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Restore window position if saved
	if pos, err := a.GetWindowPosition(); err == nil && pos != nil {
		// Only restore if position seems valid (not off-screen)
		if pos.Width > 0 && pos.Height > 0 {
			runtime.WindowSetSize(ctx, pos.Width, pos.Height)
			runtime.WindowSetPosition(ctx, pos.X, pos.Y)
		}
	}
}

// beforeClose is called before the application terminates
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	runtime.LogInfof(ctx, "Application closing, cleaning up resources...")

	// Save window position before closing
	if err := a.SaveWindowPosition(); err != nil {
		runtime.LogErrorf(ctx, "Failed to save window position: %v", err)
	}

	// Stop log watcher if running
	if a.logWatcher != nil {
		runtime.LogInfof(ctx, "Stopping log watcher...")
		a.logWatcher.Stop()
	}

	// Stop HTTP server
	runtime.LogInfof(ctx, "Stopping HTTP server...")
	a.stopHTTPServer()

	runtime.LogInfof(ctx, "Cleanup complete")
	return false
}

// Implementations for SetTaskbarBadge are platform-specific and live in
// files guarded by build tags (badge_windows.go, badge_darwin.go, badge_unix.go).
