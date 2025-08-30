package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	gosys "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	messageCounter int
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		runtime.LogErrorf(ctx, "Failed to load config.yml: %v", err)
		runtime.Quit(ctx)
		return
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
func (a *App) SaveFrontendConfig(partial map[string]string) error {
	cfg, err := LoadConfig()
	if err != nil {
		// if config doesn't exist, start from defaults
		cfg = &Config{Server: "localhost", Port: 9191}
	}
	if v, ok := partial["theme"]; ok {
		cfg.Theme = v
	}
	if v, ok := partial["language"]; ok {
		cfg.Lang = v
	}
	if v, ok := partial["show_types"]; ok {
		// interpret 'true'/'false'
		if v == "true" {
			cfg.ShowTypes = true
		} else {
			cfg.ShowTypes = false
		}
	}
	return SaveConfig(cfg)
}

// UpdateVisibleCount updates the internal counter and window title based on the
// number of messages currently visible in the UI. This should be called by the frontend
// whenever logs are added/removed/cleared so the title and any OS badges remain in sync.
func (a *App) UpdateVisibleCount(count int) error {
	a.messageCounter = count
	if a.ctx != nil {
		runtime.WindowSetTitle(a.ctx, fmt.Sprintf("VersaDumps Visualizer (%d)", a.messageCounter))
	}
	// Update platform-specific taskbar/tray badge if available
	SetTaskbarBadge(a.ctx, a.messageCounter)
	return nil
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

// Implementations for SetTaskbarBadge are platform-specific and live in
// files guarded by build tags (badge_windows.go, badge_darwin.go, badge_unix.go).
