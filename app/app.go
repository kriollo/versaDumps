package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	gosys "runtime"
	"strconv"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	messageCounter int
	updateManager  *UpdateManager
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
		runtime.LogErrorf(ctx, "Failed to load or create config.yml: %v", err)
		// Use fallback defaults if config creation failed
		cfg = &Config{
			Server:    "localhost",
			Port:      9191,
			Theme:     "dark",
			Lang:      "en",
			ShowTypes: true,
		}
		runtime.LogWarningf(ctx, "Using default configuration due to error")
	} else if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		// Config was just created with defaults
		runtime.LogInfof(ctx, "Created config.yml with default values")
	}

	// Emit loaded config to frontend so it can initialize theme/language
	// Encode as JSON string
	cfgBytes, _ := json.Marshal(cfg)
	runtime.EventsEmit(ctx, "configLoaded", string(cfgBytes))

	// Start the background HTTP server, passing the app instance
	StartServer(ctx, cfg.Server, cfg.Port, a)

	// Initialize window title with current counter
	if a.ctx != nil {
		runtime.WindowSetTitle(a.ctx, fmt.Sprintf("VersaDumps Visualizer (%d)", a.messageCounter))
	}
}

// GetVisibleCount returns the current visible message count
func (a *App) GetVisibleCount() int {
	return a.messageCounter
}

// GetConfig returns the current configuration (callable from frontend)
func (a *App) GetConfig() (*Config, error) {
	return LoadConfig()
}

// SaveFrontendConfig accepts partial configuration from frontend and persists it
func (a *App) SaveFrontendConfig(partial map[string]interface{}) error {
	cfg, err := LoadConfig()
	if err != nil {
		// if config doesn't exist, start from defaults
		cfg = &Config{Server: "localhost", Port: 9191}
	}
	
	// Handle server field
	if v, ok := partial["server"]; ok {
		if serverStr, ok := v.(string); ok && serverStr != "" {
			cfg.Server = serverStr
		}
	}
	
	// Handle port field
	if v, ok := partial["port"]; ok {
		switch port := v.(type) {
		case float64:
			cfg.Port = int(port)
		case int:
			cfg.Port = port
		case string:
			// Try to parse string to int
			if portInt, err := strconv.Atoi(port); err == nil {
				cfg.Port = portInt
			}
		}
	}
	
	// Handle theme field
	if v, ok := partial["theme"]; ok {
		if themeStr, ok := v.(string); ok {
			cfg.Theme = themeStr
		}
	}
	
	// Handle language field
	if v, ok := partial["language"]; ok {
		if langStr, ok := v.(string); ok {
			cfg.Lang = langStr
		}
	}
	
	// Handle show_types field
	if v, ok := partial["show_types"]; ok {
		switch showTypes := v.(type) {
		case bool:
			cfg.ShowTypes = showTypes
		case string:
			cfg.ShowTypes = (showTypes == "true")
		}
	}
	
	return SaveConfig(cfg)
}

// UpdateVisibleCount updates the internal counter and window title based on the
// number of messages currently visible in the UI. This should be called by the frontend
// whenever logs are added/removed/cleared so the title and any OS badges remain in sync.
func (a *App) UpdateVisibleCount(count int) {
	a.messageCounter = count
	if a.ctx != nil {
		runtime.WindowSetTitle(a.ctx, fmt.Sprintf("VersaDumps Visualizer (%d)", a.messageCounter))
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

// DownloadAndInstallUpdate descarga e instala la actualización
func (a *App) DownloadAndInstallUpdate(downloadURL string) error {
	// Mostrar diálogo de progreso
	runtime.EventsEmit(a.ctx, "updateDownloadProgress", map[string]interface{}{
		"status": "starting",
		"progress": 0,
	})

	// Descargar con callback de progreso
	filePath, err := a.updateManager.DownloadUpdate(downloadURL, func(downloaded, total int64) {
		if total > 0 {
			progress := float64(downloaded) / float64(total) * 100
			runtime.EventsEmit(a.ctx, "updateDownloadProgress", map[string]interface{}{
				"status":   "downloading",
				"progress": progress,
				"downloaded": downloaded,
				"total":    total,
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

// Implementations for SetTaskbarBadge are platform-specific and live in
// files guarded by build tags (badge_windows.go, badge_darwin.go, badge_unix.go).
