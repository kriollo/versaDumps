package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LogWatcher monitors log files and directories for changes
type LogWatcher struct {
	ctx      context.Context
	watcher  *fsnotify.Watcher
	files    map[string]*LogFile // path -> LogFile
	folders  []LogFolder
	running  bool
	mu       sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup // Track active goroutines
}

// LogFile represents a monitored log file
type LogFile struct {
	Path         string
	File         *os.File
	LastPosition int64
	LastModTime  time.Time
	mu           sync.Mutex
}

// LogEntry represents a single log line with metadata
type LogEntry struct {
	FilePath  string    `json:"filePath"`
	FileName  string    `json:"fileName"`
	Line      string    `json:"line"`
	Level     string    `json:"level"` // error, warning, info, debug, etc.
	Timestamp time.Time `json:"timestamp"`
	LineNum   int       `json:"lineNum"`
}

// NewLogWatcher creates a new LogWatcher instance
func NewLogWatcher(ctx context.Context) (*LogWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &LogWatcher{
		ctx:      ctx,
		watcher:  watcher,
		files:    make(map[string]*LogFile),
		folders:  []LogFolder{},
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins monitoring the configured folders
func (lw *LogWatcher) Start(folders []LogFolder) error {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.running {
		return fmt.Errorf("log watcher is already running")
	}

	lw.folders = folders
	lw.running = true

	// Add all folders and their files to the watcher
	for _, folder := range folders {
		if !folder.Enabled {
			continue
		}

		if err := lw.addFolder(folder); err != nil {
			runtime.LogErrorf(lw.ctx, "Error adding folder %s: %v", folder.Path, err)
			continue
		}
	}

	// Start the event loop in a goroutine
	lw.wg.Add(1)
	go lw.eventLoop()

	runtime.LogInfof(lw.ctx, "Log watcher started, monitoring %d folders", len(folders))
	return nil
}

// Stop stops the log watcher
func (lw *LogWatcher) Stop() {
	lw.mu.Lock()
	if !lw.running {
		lw.mu.Unlock()
		return
	}
	lw.running = false
	lw.mu.Unlock()

	// Signal stop and wait for goroutines to finish
	close(lw.stopChan)
	lw.wg.Wait()

	// Now safely close resources
	lw.mu.Lock()
	defer lw.mu.Unlock()

	// Close all open files
	for path, logFile := range lw.files {
		if logFile.File != nil {
			logFile.File.Close()
			runtime.LogInfof(lw.ctx, "Closed file: %s", path)
		}
	}

	// Clear the files map
	lw.files = make(map[string]*LogFile)

	// Close the watcher
	if lw.watcher != nil {
		lw.watcher.Close()
	}

	runtime.LogInfof(lw.ctx, "Log watcher stopped and resources cleaned up")
}

// addFolder adds a folder and its matching files to the watcher
func (lw *LogWatcher) addFolder(folder LogFolder) error {
	// Check if folder exists
	info, err := os.Stat(folder.Path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", folder.Path)
	}

	// Add the folder to the watcher
	if err := lw.watcher.Add(folder.Path); err != nil {
		return err
	}

	// Find all matching files in the folder
	files, err := lw.findMatchingFiles(folder)
	if err != nil {
		return err
	}

	// Start tailing each file
	for _, filePath := range files {
		if err := lw.tailFile(filePath, folder); err != nil {
			runtime.LogErrorf(lw.ctx, "Error tailing file %s: %v", filePath, err)
		}
	}

	return nil
}

// findMatchingFiles returns all files in a folder that match the extensions
func (lw *LogWatcher) findMatchingFiles(folder LogFolder) ([]string, error) {
	var matchingFiles []string

	err := filepath.Walk(folder.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file matches any of the extensions
		for _, ext := range folder.Extensions {
			// Convert glob pattern to regex (simple implementation)
			pattern := strings.ReplaceAll(ext, ".", "\\.")
			pattern = strings.ReplaceAll(pattern, "*", ".*")
			pattern = "^" + pattern + "$"

			matched, _ := regexp.MatchString(pattern, filepath.Base(path))
			if matched {
				matchingFiles = append(matchingFiles, path)
				break
			}
		}

		return nil
	})

	return matchingFiles, err
}

// tailFile starts tailing a file from the end (like tail -f)
func (lw *LogWatcher) tailFile(filePath string, folder LogFolder) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	// Seek to end of file (tail -f behavior)
	// If you want to read existing content, use: file.Seek(0, io.SeekStart)
	position, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		file.Close()
		return err
	}

	logFile := &LogFile{
		Path:         filePath,
		File:         file,
		LastPosition: position,
		LastModTime:  info.ModTime(),
	}

	lw.files[filePath] = logFile

	runtime.LogInfof(lw.ctx, "Started tailing file: %s", filePath)
	return nil
}

// eventLoop processes file system events
func (lw *LogWatcher) eventLoop() {
	defer lw.wg.Done()

	for {
		select {
		case <-lw.stopChan:
			runtime.LogInfof(lw.ctx, "Event loop stopping...")
			return

		case event, ok := <-lw.watcher.Events:
			if !ok {
				runtime.LogInfof(lw.ctx, "Watcher events channel closed")
				return
			}

			lw.handleEvent(event)

		case err, ok := <-lw.watcher.Errors:
			if !ok {
				runtime.LogInfof(lw.ctx, "Watcher errors channel closed")
				return
			}
			runtime.LogErrorf(lw.ctx, "Watcher error: %v", err)
		}
	}
}

// handleEvent processes a file system event
func (lw *LogWatcher) handleEvent(event fsnotify.Event) {
	lw.mu.RLock()
	logFile, exists := lw.files[event.Name]
	lw.mu.RUnlock()

	if !exists {
		// New file created, check if it matches our patterns
		lw.handleNewFile(event.Name)
		return
	}

	// Handle modification of existing file
	if event.Op&fsnotify.Write == fsnotify.Write {
		lw.readNewLines(logFile)
	}

	// Handle file removal/rename
	if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
		lw.removeFile(event.Name)
	}
}

// handleNewFile checks if a new file matches our patterns and starts tailing it
func (lw *LogWatcher) handleNewFile(filePath string) {
	lw.mu.RLock()
	// Make a copy of folders to avoid holding the lock during file operations
	folders := make([]LogFolder, len(lw.folders))
	copy(folders, lw.folders)
	lw.mu.RUnlock()

	for _, folder := range folders {
		if !folder.Enabled {
			continue
		}

		// Check if file is in this folder
		if !strings.HasPrefix(filePath, folder.Path) {
			continue
		}

		// Check if file matches extensions
		for _, ext := range folder.Extensions {
			pattern := strings.ReplaceAll(ext, ".", "\\.")
			pattern = strings.ReplaceAll(pattern, "*", ".*")
			pattern = "^" + pattern + "$"

			matched, _ := regexp.MatchString(pattern, filepath.Base(filePath))
			if matched {
				// Lock only when actually modifying the files map
				lw.mu.Lock()
				lw.tailFile(filePath, folder)
				lw.mu.Unlock()
				return
			}
		}
	}
}

// removeFile closes and removes a file from tracking
func (lw *LogWatcher) removeFile(filePath string) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if logFile, exists := lw.files[filePath]; exists {
		if logFile.File != nil {
			logFile.File.Close()
		}
		delete(lw.files, filePath)
		runtime.LogInfof(lw.ctx, "Stopped tailing file: %s", filePath)
	}
}

// readNewLines reads new lines from a log file and emits them
func (lw *LogWatcher) readNewLines(logFile *LogFile) {
	logFile.mu.Lock()
	defer logFile.mu.Unlock()

	// Get current file info
	info, err := logFile.File.Stat()
	if err != nil {
		runtime.LogErrorf(lw.ctx, "Error getting file info: %v", err)
		return
	}

	currentSize := info.Size()

	// If file was truncated (log rotation), start from beginning
	if currentSize < logFile.LastPosition {
		logFile.File.Seek(0, io.SeekStart)
		logFile.LastPosition = 0
	}

	// Read new lines with buffer limit to prevent memory issues
	scanner := bufio.NewScanner(logFile.File)

	// Set maximum token size (1MB per line)
	const maxScanTokenSize = 1024 * 1024 // 1MB
	buf := make([]byte, 0, 64*1024)      // 64KB initial buffer
	scanner.Buffer(buf, maxScanTokenSize)

	lineNum := 0
	const maxLinesPerRead = 1000 // Limit lines per read to prevent blocking

	// Find the folder config for filters (make a copy to avoid holding lock)
	lw.mu.RLock()
	var folder *LogFolder
	for i := range lw.folders {
		if strings.HasPrefix(logFile.Path, lw.folders[i].Path) {
			// Make a copy of the folder to avoid race conditions
			folderCopy := lw.folders[i]
			folder = &folderCopy
			break
		}
	}
	lw.mu.RUnlock()

	for scanner.Scan() {
		lineNum++

		// Prevent reading too many lines in one go
		if lineNum > maxLinesPerRead {
			runtime.LogInfof(lw.ctx, "Max lines per read reached for %s, will continue on next event", logFile.Path)
			break
		}

		line := scanner.Text()

		// Detect log level
		level := detectLogLevel(line)

		// Apply filters if configured
		if folder != nil && len(folder.Filters) > 0 {
			if !matchesFilter(level, folder.Filters) {
				continue
			}
		}

		// Create log entry
		entry := LogEntry{
			FilePath:  logFile.Path,
			FileName:  filepath.Base(logFile.Path),
			Line:      line,
			Level:     level,
			Timestamp: time.Now(),
			LineNum:   lineNum,
		}

		// Emit to frontend
		runtime.EventsEmit(lw.ctx, "logLine", entry)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		// Check if it's a file locking error (common in Windows)
		errMsg := err.Error()
		if strings.Contains(errMsg, "locked a portion of the file") ||
			strings.Contains(errMsg, "being used by another process") ||
			strings.Contains(errMsg, "access is denied") {
			// Silently skip locked files - this is normal when the app is writing to logs
			runtime.LogDebugf(lw.ctx, "File %s is locked by another process, will retry later", filepath.Base(logFile.Path))
			return
		}
		// Log other errors
		runtime.LogErrorf(lw.ctx, "Error scanning file %s: %v", logFile.Path, err)
	}

	// Update position
	newPosition, _ := logFile.File.Seek(0, io.SeekCurrent)
	logFile.LastPosition = newPosition
	logFile.LastModTime = info.ModTime()
}

// detectLogLevel attempts to detect the log level from a line
func detectLogLevel(line string) string {
	lineLower := strings.ToLower(line)

	// Common log level patterns
	patterns := map[string][]string{
		"error":   {"error", "err", "fatal", "critical", "exception"},
		"warning": {"warning", "warn"},
		"info":    {"info", "information"},
		"debug":   {"debug", "trace"},
		"success": {"success", "ok", "passed"},
	}

	for level, keywords := range patterns {
		for _, keyword := range keywords {
			if strings.Contains(lineLower, keyword) {
				return level
			}
		}
	}

	return "info" // default
}

// matchesFilter checks if a log level matches any of the filters
func matchesFilter(level string, filters []string) bool {
	if len(filters) == 0 {
		return true // No filters = show all
	}

	for _, filter := range filters {
		if strings.EqualFold(level, filter) {
			return true
		}
	}

	return false
}
