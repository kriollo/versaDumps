//go:build linux || freebsd || openbsd || netbsd

package main

import "context"

// SetTaskbarBadge on Unix-like systems: best-effort. Many DEs don't support overlay icons.
func SetTaskbarBadge(ctx context.Context, count int) {
	// TODO: implement tray icon update or libappindicator integration if desired
}
