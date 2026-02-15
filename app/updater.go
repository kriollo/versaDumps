package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Version actual de la aplicación
const CurrentVersion = "3.1.0"

// GitHubRelease estructura para parsear la respuesta de GitHub API
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
	HTMLURL     string    `json:"html_url"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int    `json:"size"`
}

type UpdateInfo struct {
	Available      bool   `json:"available"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	DownloadURL    string `json:"downloadUrl"`
	ReleaseURL     string `json:"releaseUrl"`
	Size           int    `json:"size"`
	CurrentVersion string `json:"currentVersion"`
}

type UpdateManager struct {
	owner      string
	repo       string
	httpClient *http.Client
}

func NewUpdateManager() *UpdateManager {
	return &UpdateManager{
		owner: "kriollo",
		repo:  "versaDumps",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckForUpdates verifica si hay una nueva versión disponible
func (um *UpdateManager) CheckForUpdates() (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", um.owner, um.repo)
	fmt.Printf("CheckForUpdates: Checking URL: %s\n", url)
	fmt.Printf("CheckForUpdates: Current version: %s\n", CurrentVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("CheckForUpdates: Error creating request: %v\n", err)
		return nil, err
	}

	// GitHub API requiere User-Agent
	req.Header.Set("User-Agent", "VersaDumps-Updater")
	// Agregar un token de autenticación si está disponible (para evitar rate limiting)
	// req.Header.Set("Authorization", "token YOUR_TOKEN_HERE")

	resp, err := um.httpClient.Do(req)
	if err != nil {
		fmt.Printf("CheckForUpdates: Error making request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("CheckForUpdates: GitHub API returned status: %d\n", resp.StatusCode)
		// Leer el cuerpo de la respuesta para más detalles
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("CheckForUpdates: Response body: %s\n", string(body))

		// Si hay rate limiting o cualquier error, retornar "no hay actualización"
		// en lugar de una actualización falsa
		if resp.StatusCode == 403 {
			fmt.Printf("CheckForUpdates: Rate limiting detected, returning no update available\n")
			return &UpdateInfo{
				Available:      false,          // Changed from true to false
				Version:        CurrentVersion, // Use current version instead of fake version
				Description:    "Unable to check for updates due to API rate limiting",
				DownloadURL:    "",
				ReleaseURL:     "https://github.com/kriollo/versaDumps/releases",
				Size:           0,
				CurrentVersion: CurrentVersion,
			}, nil
		}

		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("CheckForUpdates: Error decoding response: %v\n", err)
		return nil, err
	}

	// Limpiar la versión (quitar 'v' si existe)
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(CurrentVersion, "v")

	fmt.Printf("CheckForUpdates: Latest version: %s, Current version: %s\n", latestVersion, currentVersion)

	// Comparar versiones
	comparison := compareVersions(latestVersion, currentVersion)
	fmt.Printf("CheckForUpdates: Version comparison result: %d (1=newer, 0=same, -1=older)\n", comparison)

	if comparison <= 0 {
		fmt.Printf("CheckForUpdates: No update available\n")
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: CurrentVersion,
			Version:        CurrentVersion, // Use current version, not latest when no update available
		}, nil
	}

	// Buscar el asset correcto para el SO actual
	downloadURL, size := um.getDownloadURL(release.Assets)

	fmt.Printf("CheckForUpdates: Update available! Version: %s, Download URL: %s\n", latestVersion, downloadURL)

	return &UpdateInfo{
		Available:      true,
		Version:        latestVersion,
		Description:    release.Body,
		DownloadURL:    downloadURL,
		ReleaseURL:     release.HTMLURL,
		Size:           size,
		CurrentVersion: CurrentVersion,
	}, nil
}

// getDownloadURL obtiene la URL de descarga correcta según el SO
func (um *UpdateManager) getDownloadURL(assets []Asset) (string, int) {
	osArch := runtime.GOOS + "-" + runtime.GOARCH

	var patterns []string
	switch runtime.GOOS {
	case "windows":
		// Preferir el instalador sobre el portable
		patterns = []string{"setup-amd64.exe", "windows-amd64", "windows"}
	case "darwin":
		patterns = []string{"macos-amd64", "darwin-amd64", "darwin", "macos"}
	case "linux":
		// On Linux try to prefer distro packages when possible (rpm/deb), then archives, then generic linux artifacts
		pkgType := detectPackageType()
		if pkgType == "rpm" {
			// Prefer RPM packages
			for _, asset := range assets {
				if strings.HasSuffix(strings.ToLower(asset.Name), ".rpm") {
					return asset.BrowserDownloadURL, asset.Size
				}
			}
		} else if pkgType == "deb" {
			for _, asset := range assets {
				if strings.HasSuffix(strings.ToLower(asset.Name), ".deb") {
					return asset.BrowserDownloadURL, asset.Size
				}
			}
		}

		// Fallback to archives or linux-specific assets
		patterns = []string{"linux-amd64", "linux", ".tar.gz", ".tgz"}
	default:
		patterns = []string{osArch}
	}

	for _, pattern := range patterns {
		for _, asset := range assets {
			if strings.Contains(strings.ToLower(asset.Name), pattern) {
				return asset.BrowserDownloadURL, asset.Size
			}
		}
	}

	// Si no encuentra uno específico, devolver el primero
	if len(assets) > 0 {
		return assets[0].BrowserDownloadURL, assets[0].Size
	}

	return "", 0
}

// detectPackageType intenta detectar si el sistema es RPM o DEB basado en /etc/os-release
func detectPackageType() string {
	// Default: no specific package type
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	s := strings.ToLower(string(data))
	if strings.Contains(s, "fedora") || strings.Contains(s, "centos") || strings.Contains(s, "rhel") || strings.Contains(s, "redhat") || strings.Contains(s, "rocky") || strings.Contains(s, "almalinux") {
		return "rpm"
	}
	if strings.Contains(s, "debian") || strings.Contains(s, "ubuntu") || strings.Contains(s, "linuxmint") || strings.Contains(s, "pop") {
		return "deb"
	}
	return ""
}

// DownloadUpdate descarga la actualización
func (um *UpdateManager) DownloadUpdate(downloadURL string, onProgress func(downloaded, total int64)) (string, error) {
	resp, err := um.httpClient.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Crear directorio temporal para la descarga
	tempDir := os.TempDir()
	fileName := filepath.Base(downloadURL)
	filePath := filepath.Join(tempDir, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Crear un reader con callback de progreso
	reader := &progressReader{
		Reader:     resp.Body,
		onProgress: onProgress,
		total:      resp.ContentLength,
	}

	_, err = io.Copy(out, reader)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// InstallUpdate está implementado en archivos específicos por plataforma:
// - updater_windows.go para Windows
// - updater_unix.go para Linux/macOS

// progressReader implementa io.Reader con callback de progreso
type progressReader struct {
	io.Reader
	onProgress func(downloaded, total int64)
	total      int64
	downloaded int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.downloaded += int64(n)
	if pr.onProgress != nil {
		pr.onProgress(pr.downloaded, pr.total)
	}
	return n, err
}

// compareVersions compara dos versiones en formato semver
// Retorna: -1 si v1 < v2, 0 si v1 == v2, 1 si v1 > v2
func compareVersions(v1, v2 string) int {
	// Log the comparison for debugging
	fmt.Printf("compareVersions: Comparing v1='%s' with v2='%s'\n", v1, v2)

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Ensure we have at least 3 parts for major.minor.patch
	maxLen := 3
	if len(parts1) > maxLen {
		maxLen = len(parts1)
	}
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &p1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &p2)
		}

		fmt.Printf("compareVersions: Part %d: p1=%d, p2=%d\n", i, p1, p2)

		if p1 < p2 {
			fmt.Printf("compareVersions: Result: -1 (v1 < v2)\n")
			return -1
		}
		if p1 > p2 {
			fmt.Printf("compareVersions: Result: 1 (v1 > v2)\n")
			return 1
		}
	}

	fmt.Printf("compareVersions: Result: 0 (v1 == v2)\n")
	return 0
}
