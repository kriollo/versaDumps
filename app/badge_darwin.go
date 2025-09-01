//go:build darwin

package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

// SetTaskbarBadge sets the dock badge on macOS using AppleScript
func SetTaskbarBadge(ctx context.Context, count int) {
	log.Printf("[Badge] Setting macOS dock badge to: %d", count)
	
	var script string
	if count <= 0 {
		// Clear the badge
		script = `tell application "System Events" to tell process "app" to set badge of dock tile to ""`
	} else {
		// Set the badge number
		badgeText := fmt.Sprintf("%d", count)
		if count > 99 {
			badgeText = "99+"
		}
		script = fmt.Sprintf(`tell application "System Events" to tell process "app" to set badge of dock tile to "%s"`, badgeText)
	}
	
	// Execute AppleScript
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		// Fallback: Try using the app's bundle identifier if available
		log.Printf("[Badge] Failed to set badge via AppleScript: %v", err)
	}
}
