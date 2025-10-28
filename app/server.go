package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StartServer starts the HTTP server in a new goroutine.
func StartServer(ctx context.Context, host string, port int, app *App) {
	go func() {
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

			body, err := io.ReadAll(r.Body)
			if err != nil {
				runtime.LogErrorf(ctx, "Error reading request body: %v", err)
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}

			// DEBUG: Log del payload RAW recibido
			runtime.LogInfof(ctx, "════════════════════════════════════════")
			runtime.LogInfof(ctx, "RAW PAYLOAD RECEIVED:")
			runtime.LogInfof(ctx, "%s", string(body))
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
			Addr:    serverAddr,
			Handler: mux,
		}

		runtime.LogInfof(ctx, "About to call ListenAndServe...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			runtime.LogErrorf(ctx, "HTTP server failed to start: %v", err)
			runtime.LogErrorf(ctx, "Server address was: %s", serverAddr)
		} else {
			runtime.LogInfof(ctx, "HTTP server started successfully on %s", serverAddr)
		}
	}()
}
