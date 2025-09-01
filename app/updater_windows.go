//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// InstallUpdateWindows instala la actualización en Windows
func (um *UpdateManager) InstallUpdate(filePath string) error {
	// En Windows, ejecutar el instalador con elevación de privilegios
	if strings.HasSuffix(filePath, ".exe") {
		// Usar PowerShell Start-Process con -Verb RunAs para solicitar elevación
		// No usar /S para que el usuario vea el progreso de la instalación
		psCmd := fmt.Sprintf(`Start-Process -FilePath "%s" -Verb RunAs -Wait`, filePath)
		cmd := exec.Command("powershell", "-Command", psCmd)
		return cmd.Start()
	}
	// Si es un zip, descomprimir (necesitaría implementación adicional)
	return fmt.Errorf("archivo no soportado para instalación automática: %s", filePath)
}
