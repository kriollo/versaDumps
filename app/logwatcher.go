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

// LogFile represents a monitored log file (NO FILE HANDLE - only metadata)
type LogFile struct {
	Path         string
	LastPosition int64
	LastModTime  time.Time
	LastSize     int64
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

	if len(folders) == 0 {
		runtime.LogInfof(lw.ctx, "No log folders configured, log watcher will not start")
		return nil
	}

	lw.folders = folders
	lw.running = true

	runtime.LogInfof(lw.ctx, "Starting log watcher with %d folders", len(folders))

	// Add all folders and their files to the watcher
	enabledCount := 0
	for _, folder := range folders {
		if !folder.Enabled {
			runtime.LogInfof(lw.ctx, "Skipping disabled folder: %s", folder.Path)
			continue
		}

		runtime.LogInfof(lw.ctx, "Adding folder: %s (extensions: %v, format: %s)",
			folder.Path, folder.Extensions, folder.Format)

		if err := lw.addFolder(folder); err != nil {
			runtime.LogErrorf(lw.ctx, "Error adding folder %s: %v", folder.Path, err)
			continue
		}
		enabledCount++
	}

	if enabledCount == 0 {
		runtime.LogWarningf(lw.ctx, "No enabled log folders were successfully added")
		lw.running = false
		return fmt.Errorf("no enabled log folders available")
	}

	// Start the event loop in a goroutine
	lw.wg.Add(1)
	go lw.eventLoop()

	runtime.LogInfof(lw.ctx, "Log watcher started successfully, monitoring %d enabled folders with %d files",
		enabledCount, len(lw.files))
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

	// Clear resources
	lw.mu.Lock()
	defer lw.mu.Unlock()

	// No files to close - we don't keep them open!
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
		if os.IsNotExist(err) {
			return fmt.Errorf("folder does not exist: %s", folder.Path)
		}
		return fmt.Errorf("error accessing folder %s: %v", folder.Path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", folder.Path)
	}

	// Add the folder to the watcher
	if err := lw.watcher.Add(folder.Path); err != nil {
		return fmt.Errorf("failed to watch folder %s: %v", folder.Path, err)
	}

	runtime.LogInfof(lw.ctx, "Successfully added folder to watcher: %s", folder.Path)

	// Find all matching files in the folder
	files, err := lw.findMatchingFiles(folder)
	if err != nil {
		return fmt.Errorf("error finding files in %s: %v", folder.Path, err)
	}

	runtime.LogInfof(lw.ctx, "Found %d matching files in folder %s", len(files), folder.Path)

	// Register each file for monitoring (but don't open them)
	registeredCount := 0
	for _, filePath := range files {
		if err := lw.registerFile(filePath); err != nil {
			runtime.LogErrorf(lw.ctx, "Error registering file %s: %v", filePath, err)
			continue
		}
		registeredCount++
	}

	runtime.LogInfof(lw.ctx, "Successfully registered %d/%d files from folder %s",
		registeredCount, len(files), folder.Path)

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

// registerFile registers a file for monitoring WITHOUT opening it
func (lw *LogWatcher) registerFile(filePath string) error {
	// Get file info to store initial state
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Create LogFile metadata (NO FILE HANDLE)
	logFile := &LogFile{
		Path:         filePath,
		LastPosition: info.Size(), // Start at end (tail -f behavior)
		LastModTime:  info.ModTime(),
		LastSize:     info.Size(),
	}

	lw.files[filePath] = logFile

	runtime.LogInfof(lw.ctx, "Registered file for monitoring: %s", filePath)
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

	// Handle file creation
	if event.Op&fsnotify.Create == fsnotify.Create {
		if !exists {
			lw.handleNewFile(event.Name)
		}
		return
	}

	if !exists {
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

// handleNewFile checks if a new file matches our patterns and starts monitoring it
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
				// Small delay to ensure file is ready
				time.Sleep(100 * time.Millisecond)

				// Lock only when actually modifying the files map
				lw.mu.Lock()
				lw.registerFile(filePath)
				lw.mu.Unlock()
				return
			}
		}
	}
}

// removeFile removes a file from tracking (no file handles to close!)
func (lw *LogWatcher) removeFile(filePath string) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if _, exists := lw.files[filePath]; exists {
		delete(lw.files, filePath)
		runtime.LogInfof(lw.ctx, "Stopped monitoring file: %s", filePath)
	}
}

// readNewLines reads new lines from a log file by opening it temporarily
func (lw *LogWatcher) readNewLines(logFile *LogFile) {
	logFile.mu.Lock()
	defer logFile.mu.Unlock()

	// Check if file still exists
	info, err := os.Stat(logFile.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// File was deleted, will be handled by remove event
			return
		}
		runtime.LogErrorf(lw.ctx, "Error getting file info for %s: %v", logFile.Path, err)
		return
	}

	currentSize := info.Size()

	// Detect log rotation (file was truncated or recreated)
	if currentSize < logFile.LastPosition {
		runtime.LogInfof(lw.ctx, "Log rotation detected for %s, restarting from beginning", filepath.Base(logFile.Path))
		logFile.LastPosition = 0
		logFile.LastSize = 0
	}

	// If no new content, skip
	if currentSize == logFile.LastPosition {
		return
	}

	// Open file ONLY for reading (shared read access)
	file, err := os.OpenFile(logFile.Path, os.O_RDONLY, 0)
	if err != nil {
		runtime.LogErrorf(lw.ctx, "Error opening file %s: %v", logFile.Path, err)
		return
	}
	// IMPORTANT: Always close the file when done
	defer file.Close()

	// Seek to last known position
	_, err = file.Seek(logFile.LastPosition, io.SeekStart)
	if err != nil {
		runtime.LogErrorf(lw.ctx, "Error seeking file %s: %v", logFile.Path, err)
		return
	}

	// Read new lines with buffer limit
	scanner := bufio.NewScanner(file)

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
			runtime.LogInfof(lw.ctx, "Max lines per read reached for %s, will continue on next event", filepath.Base(logFile.Path))
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
		runtime.LogErrorf(lw.ctx, "Error scanning file %s: %v", logFile.Path, err)
	}

	// Update position to current file pointer
	newPosition, err := file.Seek(0, io.SeekCurrent)
	if err == nil {
		logFile.LastPosition = newPosition
		logFile.LastModTime = info.ModTime()
		logFile.LastSize = currentSize
	}

	// File is automatically closed by defer when function exits
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
