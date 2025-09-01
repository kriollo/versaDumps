//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// InstallUpdateWindows instala la actualización en Windows
func (um *UpdateManager) InstallUpdate(filePath string) error {
	// En Windows, ejecutar el instalador con elevación de privilegios
	if strings.HasSuffix(filePath, ".exe") {
		// Usar PowerShell Start-Process con -Verb RunAs para solicitar elevación
		// El instalador se ejecutará y la app actual se cerrará después
		psCmd := fmt.Sprintf(`Start-Process -FilePath "%s" -Verb RunAs`, filePath)
		cmd := exec.Command("powershell", "-Command", psCmd)
		
		// Iniciar el instalador
		if err := cmd.Start(); err != nil {
			return err
		}
		
		// Dar tiempo para que el instalador se inicie antes de cerrar la app
		go func() {
			time.Sleep(2 * time.Second)
			os.Exit(0) // Cerrar la aplicación actual
		}()
		
		return nil
	}
	// Si es un zip, descomprimir (necesitaría implementación adicional)
	return fmt.Errorf("archivo no soportado para instalación automática: %s", filePath)
}
