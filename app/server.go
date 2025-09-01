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
		mux := http.NewServeMux()
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

		if err := http.ListenAndServe(serverAddr, mux); err != nil {
			runtime.LogErrorf(ctx, "HTTP server error: %v", err)
		}
	}()
}