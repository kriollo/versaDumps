package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StartServer starts the HTTP server and returns the server instance for graceful shutdown.
// The caller is responsible for managing the server lifecycle.
func StartServer(ctx context.Context, host string, port int, app *App) *http.Server {
	runtime.LogInfof(ctx, "Attempting to start HTTP server...")
	runtime.LogInfof(ctx, "Host: %s, Port: %d", host, port)

	mux := http.NewServeMux()

	// Health endpoint to report server status
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		runtime.LogInfof(ctx, "Health endpoint accessed from %s", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Limit request body size to 10MB to prevent memory exhaustion
		const maxBodySize = 10 * 1024 * 1024 // 10MB
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			runtime.LogErrorf(ctx, "Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		// DEBUG: Log del payload RAW recibido (truncate if too large)
		debugPayload := string(body)
		if len(debugPayload) > 1000 {
			debugPayload = debugPayload[:1000] + "... (truncated)"
		}
		runtime.LogInfof(ctx, "════════════════════════════════════════")
		runtime.LogInfof(ctx, "RAW PAYLOAD RECEIVED:")
		runtime.LogInfof(ctx, "%s", debugPayload)
		runtime.LogInfof(ctx, "════════════════════════════════════════")

		var js interface{}
		if err := json.Unmarshal(body, &js); err != nil {
			runtime.LogErrorf(ctx, "Invalid JSON received: %v", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// Don't increment counter here, let frontend handle it via UpdateVisibleCount
		// This avoids double counting and ensures sync between frontend and backend

		// Emit event to the frontend with the raw JSON string
		runtime.EventsEmit(ctx, "newData", string(body))
		runtime.LogInfo(ctx, "Received and processed data successfully.")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Data received successfully"))
	})

	serverAddr := fmt.Sprintf("%s:%d", host, port)
	runtime.LogInfof(ctx, "Starting HTTP server on %s", serverAddr)
	runtime.LogInfof(ctx, "Server should be accessible at: http://%s:%d", host, port)

	server := &http.Server{
		Addr:              serverAddr,
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Start server in background goroutine
	go func() {
		runtime.LogInfof(ctx, "About to call ListenAndServe...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			runtime.LogErrorf(ctx, "HTTP server failed to start: %v", err)
			runtime.LogErrorf(ctx, "Server address was: %s", serverAddr)
		} else if err == http.ErrServerClosed {
			runtime.LogInfof(ctx, "HTTP server closed gracefully")
		}
	}()

	// Monitor context for graceful shutdown
	go func() {
		<-ctx.Done()
		runtime.LogInfof(ctx, "Context cancelled, shutting down HTTP server...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			runtime.LogErrorf(ctx, "Error shutting down server: %v", err)
		} else {
			runtime.LogInfof(ctx, "Server shutdown complete")
		}
	}()

	return server
}
