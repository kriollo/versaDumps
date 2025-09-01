//go:build !windows

package main

import (
	"fmt"
	"runtime"
)

// InstallUpdate instala la actualización en sistemas Unix
func (um *UpdateManager) InstallUpdate(filePath string) error {
	// En Unix, extraer el tar.gz y reemplazar el ejecutable
	// Esta es una implementación simplificada
	return fmt.Errorf("auto-actualización no implementada aún para %s", runtime.GOOS)
}
