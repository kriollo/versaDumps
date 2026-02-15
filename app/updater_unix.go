//go:build !windows

package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InstallUpdate instala la actualización en sistemas Unix
func (um *UpdateManager) InstallUpdate(filePath string) error {
	// Soporta paquetes tar.gz/.tgz que contienen el binario.
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("abrir paquete de actualización: %w", err)
	}
	defer f.Close()

	// Si es un RPM, intentar instalar vía gestor de paquetes (requiere privilegios)
	if strings.HasSuffix(strings.ToLower(filePath), ".rpm") {
		// Preferir pkexec si está disponible, sino usar sudo. Si ambos faltan, devolver instrucción.
		if _, err := exec.LookPath("pkexec"); err == nil {
			cmd := exec.Command("pkexec", "dnf", "install", "-y", filePath)
			output, runErr := cmd.CombinedOutput()
			if runErr != nil {
				return fmt.Errorf("instalación RPM falló: %v; salida: %s", runErr, string(output))
			}
			return nil
		}
		if _, err := exec.LookPath("sudo"); err == nil {
			cmd := exec.Command("sudo", "dnf", "install", "-y", filePath)
			output, runErr := cmd.CombinedOutput()
			if runErr != nil {
				return fmt.Errorf("instalación RPM con sudo falló: %v; salida: %s", runErr, string(output))
			}
			return nil
		}
		return fmt.Errorf("archivo .rpm detectado. Instala manualmente con 'sudo dnf install -y %s' o instala pkexec para permitir instalación desde la aplicación", filePath)
	}

	// Detectar gzip by extension
	ext := filepath.Ext(filePath)
	if ext != ".gz" && ext != ".tgz" {
		// Si no es gzip, asumimos binario directo: intentar reemplazarlo
		return replaceExecutable(filePath)
	}

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	tmpDir, err := os.MkdirTemp("", "versa-update-*")
	if err != nil {
		return fmt.Errorf("crear temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	var extractedExec string

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("leer tar: %w", err)
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		name := filepath.Base(h.Name)
		dest := filepath.Join(tmpDir, name)
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(h.Mode))
		if err != nil {
			return fmt.Errorf("crear archivo extraído: %w", err)
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return fmt.Errorf("escribir archivo extraído: %w", err)
		}
		out.Close()

		// Preferir fichero que tenga permisos ejecutables o que coincida con el nombre del ejecutable actual
		if extractedExec == "" {
			if h.Mode&0111 != 0 {
				extractedExec = dest
			}
		}
		// Also prefer same basename as current executable
		if extractedExec == "" {
			if name == filepath.Base(os.Args[0]) {
				extractedExec = dest
			}
		}
	}

	if extractedExec == "" {
		// pick any file in tmpDir as fallback
		entries, err := os.ReadDir(tmpDir)
		if err != nil || len(entries) == 0 {
			return fmt.Errorf("no se encontró ejecutable en el paquete")
		}
		extractedExec = filepath.Join(tmpDir, entries[0].Name())
	}

	return replaceExecutable(extractedExec)
}

// replaceExecutable attempts to replace the currently running executable with srcPath.
func replaceExecutable(srcPath string) error {
	currExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("obtener ejecutable actual: %w", err)
	}

	destDir := filepath.Dir(currExe)
	tempDest := filepath.Join(destDir, filepath.Base(srcPath)+".tmp")

	// Open source
	in, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("abrir ejecutable extraído: %w", err)
	}
	defer in.Close()

	// Create temp file in same dir (best chance de rename)
	out, err := os.OpenFile(tempDest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("no se puede crear archivo temporal en %s: %w", destDir, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tempDest)
		return fmt.Errorf("copiar ejecutable: %w", err)
	}
	out.Close()

	// Try to rename over the current executable
	if err := os.Rename(tempDest, currExe); err != nil {
		os.Remove(tempDest)
		// Si falla por permisos, devolver instrucción clara
		return fmt.Errorf("no se pudo sobrescribir %s: %v. Si la instalación fue global, ejecuta como root o instala usando RPM/dnf. También puedes extraer %s manualmente.", currExe, err, srcPath)
	}

	// Ensure executable bit
	if err := os.Chmod(currExe, 0755); err != nil {
		// no crítico, sólo reportar
	}

	// Hecho. Nota: no hacemos restart automático aquí.
	return nil
}
