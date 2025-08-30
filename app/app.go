package main

import (
	"context"
	"encoding/json"
	"fmt"

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
	return nil
}
