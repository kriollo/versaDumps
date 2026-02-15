//go:build windows

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	// Si es un zip, intentar extraer y ejecutar el ejecutable portable dentro
	if strings.HasSuffix(strings.ToLower(filePath), ".zip") {
		// Abrir el zip
		zr, err := zip.OpenReader(filePath)
		if err != nil {
			return fmt.Errorf("error abriendo zip de actualización: %w", err)
		}
		defer zr.Close()

		// Crear directorio temporal donde extraer
		tmpDir, err := os.MkdirTemp("", "versaDumps-update-*")
		if err != nil {
			return fmt.Errorf("error creando directorio temporal: %w", err)
		}

		var foundExe string

		// Extraer archivos y buscar el primer .exe
		for _, f := range zr.File {
			// Ignorar directorios
			if f.FileInfo().IsDir() {
				continue
			}

			// Nombre destino: solo el base name para evitar paths relativos no deseados
			destName := filepath.Base(f.Name)
			destPath := filepath.Join(tmpDir, destName)

			rc, err := f.Open()
			if err != nil {
				// intentar continuar con otros archivos
				continue
			}

			outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				rc.Close()
				continue
			}

			_, _ = io.Copy(outFile, rc)
			rc.Close()
			outFile.Close()

			// Si es un .exe, marcarlo
			if strings.HasSuffix(strings.ToLower(destName), ".exe") {
				foundExe = destPath
				break
			}
		}

		if foundExe == "" {
			return fmt.Errorf("archivo zip de actualización no contiene ejecutable para instalación automática: %s", filePath)
		}

		// Ejecutar el exe con elevación
		psCmd := fmt.Sprintf(`Start-Process -FilePath "%s" -Verb RunAs`, foundExe)
		cmd := exec.Command("powershell", "-Command", psCmd)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("error lanzando instalador contenido en zip: %w", err)
		}

		// Dar tiempo para que el instalador se inicie antes de cerrar la app
		go func() {
			time.Sleep(2 * time.Second)
			os.Exit(0) // Cerrar la aplicación actual
		}()

		return nil
	}

	return fmt.Errorf("archivo no soportado para instalación automática: %s", filePath)
}
