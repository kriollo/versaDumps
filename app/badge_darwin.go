//go:build darwin

package main

import "context"

// SetTaskbarBadge on macOS can use NSApplication dockTile badge; implement in Objective-C or via cgo.
func SetTaskbarBadge(ctx context.Context, count int) {
	// TODO: implement macOS dock badge via cocoa APIs
}
