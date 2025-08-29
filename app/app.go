package main

import (
	"context"
	

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

	// Start the background HTTP server, passing the app instance
	StartServer(ctx, cfg.Server, cfg.Port, a)
}