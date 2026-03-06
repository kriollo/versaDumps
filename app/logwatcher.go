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

// LogWatcher monitors log files and directories for changes.
type LogWatcher struct {
	appCtx  context.Context // Wails context for logging and event emission
	cancel  context.CancelFunc
	watcher *fsnotify.Watcher
	files   map[string]*LogFile // path -> LogFile
	folders []LogFolder
	running bool
	mu      sync.RWMutex
	wg      sync.WaitGroup
}

// LogFile represents a monitored log file (no persistent file handle).
type LogFile struct {
	Path         string
	LastPosition int64
	LastModTime  time.Time
	LastSize     int64
	mu           sync.Mutex
}

// LogEntry represents a single log line with metadata.
type LogEntry struct {
	FilePath  string    `json:"filePath"`
	FileName  string    `json:"fileName"`
	Line      string    `json:"line"`
	Level     string    `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	LineNum   int       `json:"lineNum"`
}

// NewLogWatcher creates a new LogWatcher instance.
func NewLogWatcher(ctx context.Context) (*LogWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %v", err)
	}
	return &LogWatcher{
		appCtx:  ctx,
		watcher: watcher,
		files:   make(map[string]*LogFile),
		folders: []LogFolder{},
	}, nil
}

// Start begins monitoring the configured folders.
func (lw *LogWatcher) Start(folders []LogFolder) error {
	lw.mu.Lock()

	if lw.running {
		lw.mu.Unlock()
		return fmt.Errorf("log watcher is already running")
	}

	if len(folders) == 0 {
		lw.mu.Unlock()
		runtime.LogInfof(lw.appCtx, "No log folders configured, log watcher will not start")
		return nil
	}

	lw.folders = folders

	enabledCount := 0
	for _, folder := range folders {
		if !folder.Enabled {
			runtime.LogInfof(lw.appCtx, "Skipping disabled folder: %s", folder.Path)
			continue
		}
		runtime.LogInfof(lw.appCtx, "Adding folder: %s (extensions: %v)", folder.Path, folder.Extensions)
		if err := lw.addFolder(folder); err != nil {
			runtime.LogErrorf(lw.appCtx, "Error adding folder %s: %v", folder.Path, err)
			continue
		}
		enabledCount++
	}

	if enabledCount == 0 {
		lw.mu.Unlock()
		runtime.LogWarningf(lw.appCtx, "No enabled log folders were successfully added")
		return fmt.Errorf("no enabled log folders available")
	}

	// Snapshot files for initial read (captured while holding the lock).
	initialFiles := make([]*LogFile, 0, len(lw.files))
	for _, f := range lw.files {
		initialFiles = append(initialFiles, f)
	}

	ctx, cancel := context.WithCancel(lw.appCtx)
	lw.cancel = cancel
	lw.running = true

	// Start event loop goroutine.
	lw.wg.Add(1)
	go lw.eventLoop(ctx)

	// Release the lock before doing file I/O.
	// readNewLines acquires lw.mu.RLock internally; holding lw.mu.Lock() here
	// would deadlock because sync.RWMutex is not reentrant.
	lw.mu.Unlock()

	// Read existing file content in a separate goroutine so the caller is not
	// blocked. The goroutine respects ctx cancellation so Stop() is fast.
	lw.wg.Add(1)
	go func() {
		defer lw.wg.Done()
		for _, f := range initialFiles {
			select {
			case <-ctx.Done():
				return
			default:
				lw.readNewLines(f)
			}
		}
	}()

	runtime.LogInfof(lw.appCtx, "Log watcher started: %d enabled folders, %d files",
		enabledCount, len(lw.files))
	return nil
}

// Stop stops the log watcher and releases all resources.
func (lw *LogWatcher) Stop() {
	lw.mu.Lock()
	wasRunning := lw.running
	lw.running = false
	cancel := lw.cancel
	lw.cancel = nil
	lw.mu.Unlock()

	// Cancel the context to signal all goroutines.
	if cancel != nil {
		cancel()
	}

	if wasRunning {
		// Wait for event loop and initial-read goroutines to finish.
		lw.wg.Wait()
		runtime.LogInfof(lw.appCtx, "Log watcher stopped")
	}

	// Always release OS resources (inotify watches), even if Start was never called.
	lw.mu.Lock()
	lw.files = make(map[string]*LogFile)
	if lw.watcher != nil {
		lw.watcher.Close()
		lw.watcher = nil
	}
	lw.mu.Unlock()
}

// GetStatus returns the current watcher status for the frontend.
func (lw *LogWatcher) GetStatus() map[string]interface{} {
	lw.mu.RLock()
	defer lw.mu.RUnlock()
	return map[string]interface{}{
		"running":     lw.running,
		"folderCount": len(lw.folders),
		"fileCount":   len(lw.files),
	}
}

// addFolder registers a directory and all its matching files with the watcher.
// Must be called while lw.mu is held (write lock).
func (lw *LogWatcher) addFolder(folder LogFolder) error {
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

	if err := lw.watcher.Add(folder.Path); err != nil {
		return fmt.Errorf("failed to watch folder %s: %v", folder.Path, err)
	}
	runtime.LogInfof(lw.appCtx, "Watching folder: %s", folder.Path)

	files, err := lw.findMatchingFiles(folder)
	if err != nil {
		return fmt.Errorf("error scanning files in %s: %v", folder.Path, err)
	}
	runtime.LogInfof(lw.appCtx, "Found %d matching files in %s", len(files), folder.Path)

	for _, filePath := range files {
		if err := lw.registerFile(filePath); err != nil {
			runtime.LogErrorf(lw.appCtx, "Error registering file %s: %v", filePath, err)
		}
	}
	return nil
}

// findMatchingFiles returns all files in folder that match the configured extensions.
func (lw *LogWatcher) findMatchingFiles(folder LogFolder) ([]string, error) {
	var matches []string
	err := filepath.Walk(folder.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		for _, ext := range folder.Extensions {
			pattern := "^" + strings.ReplaceAll(strings.ReplaceAll(ext, ".", "\\."), "*", ".*") + "$"
			if matched, _ := regexp.MatchString(pattern, filepath.Base(path)); matched {
				matches = append(matches, path)
				break
			}
		}
		return nil
	})
	return matches, err
}

// registerFile adds a file to the tracking map (no file handle kept open).
// Must be called while lw.mu write lock is held.
func (lw *LogWatcher) registerFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	lw.files[filePath] = &LogFile{
		Path:        filePath,
		LastModTime: info.ModTime(),
		LastSize:    info.Size(),
		// LastPosition starts at 0 so existing content is read on startup.
	}
	runtime.LogInfof(lw.appCtx, "Registered file: %s", filePath)
	return nil
}

// eventLoop is the main goroutine that processes fsnotify events.
func (lw *LogWatcher) eventLoop(ctx context.Context) {
	defer lw.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-lw.watcher.Events:
			if !ok {
				return
			}
			lw.handleEvent(event)
		case err, ok := <-lw.watcher.Errors:
			if !ok {
				return
			}
			runtime.LogErrorf(lw.appCtx, "Watcher error: %v", err)
		}
	}
}

// handleEvent dispatches a fsnotify event to the appropriate handler.
func (lw *LogWatcher) handleEvent(event fsnotify.Event) {
	// Use RLock since we are only reading the files map.
	lw.mu.RLock()
	logFile, exists := lw.files[event.Name]
	lw.mu.RUnlock()

	if event.Op&fsnotify.Create == fsnotify.Create {
		if !exists {
			lw.handleNewFile(event.Name)
		}
		return
	}

	if !exists {
		return
	}

	if event.Op&fsnotify.Write == fsnotify.Write {
		lw.readNewLines(logFile)
	}

	if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
		lw.removeFile(event.Name)
	}
}

// handleNewFile checks whether a newly created file matches monitored patterns
// and, if so, registers and reads it.
func (lw *LogWatcher) handleNewFile(filePath string) {
	lw.mu.RLock()
	folders := make([]LogFolder, len(lw.folders))
	copy(folders, lw.folders)
	lw.mu.RUnlock()

	for _, folder := range folders {
		if !folder.Enabled {
			continue
		}
		if !strings.HasPrefix(filePath, folder.Path) {
			continue
		}
		for _, ext := range folder.Extensions {
			pattern := "^" + strings.ReplaceAll(strings.ReplaceAll(ext, ".", "\\."), "*", ".*") + "$"
			if matched, _ := regexp.MatchString(pattern, filepath.Base(filePath)); matched {
				// Short delay to ensure the file is ready to read.
				time.Sleep(100 * time.Millisecond)

				lw.mu.Lock()
				err := lw.registerFile(filePath)
				var newFile *LogFile
				if err == nil {
					newFile = lw.files[filePath]
				}
				lw.mu.Unlock()

				if newFile != nil {
					lw.readNewLines(newFile)
				} else if err != nil {
					runtime.LogErrorf(lw.appCtx, "Error registering new file %s: %v", filePath, err)
				}
				return
			}
		}
	}
}

// removeFile removes a file from tracking.
func (lw *LogWatcher) removeFile(filePath string) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	if _, exists := lw.files[filePath]; exists {
		delete(lw.files, filePath)
		runtime.LogInfof(lw.appCtx, "Stopped monitoring: %s", filePath)
	}
}

// readNewLines opens the file, reads any new lines since the last position,
// applies filters, and emits each matching line as a "logLine" event.
func (lw *LogWatcher) readNewLines(logFile *LogFile) {
	logFile.mu.Lock()
	defer logFile.mu.Unlock()

	info, err := os.Stat(logFile.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			runtime.LogErrorf(lw.appCtx, "Stat error for %s: %v", logFile.Path, err)
		}
		return
	}

	currentSize := info.Size()

	// Detect log rotation (file shrunk or was recreated).
	if currentSize < logFile.LastPosition {
		runtime.LogInfof(lw.appCtx, "Log rotation detected for %s", filepath.Base(logFile.Path))
		logFile.LastPosition = 0
		logFile.LastSize = 0
	}

	if currentSize == logFile.LastPosition {
		return // No new content.
	}

	file, err := os.OpenFile(logFile.Path, os.O_RDONLY, 0)
	if err != nil {
		runtime.LogErrorf(lw.appCtx, "Open error for %s: %v", logFile.Path, err)
		return
	}
	defer file.Close()

	if _, err = file.Seek(logFile.LastPosition, io.SeekStart); err != nil {
		runtime.LogErrorf(lw.appCtx, "Seek error for %s: %v", logFile.Path, err)
		return
	}

	// Find the folder config for this file (read lock, short scope).
	lw.mu.RLock()
	var folder *LogFolder
	for i := range lw.folders {
		if strings.HasPrefix(logFile.Path, lw.folders[i].Path) {
			cp := lw.folders[i]
			folder = &cp
			break
		}
	}
	lw.mu.RUnlock()

	const maxScanTokenSize = 1024 * 1024 // 1 MB per line
	const maxLinesPerRead = 1000

	buf := make([]byte, 0, 64*1024)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(buf, maxScanTokenSize)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum > maxLinesPerRead {
			runtime.LogInfof(lw.appCtx, "Max lines/read reached for %s", filepath.Base(logFile.Path))
			break
		}

		line := scanner.Text()
		level := detectLogLevel(line)

		if folder != nil && len(folder.Filters) > 0 && !matchesFilter(level, folder.Filters) {
			continue
		}

		runtime.EventsEmit(lw.appCtx, "logLine", LogEntry{
			FilePath:  logFile.Path,
			FileName:  filepath.Base(logFile.Path),
			Line:      line,
			Level:     level,
			Timestamp: time.Now(),
			LineNum:   lineNum,
		})
	}

	if err := scanner.Err(); err != nil {
		runtime.LogErrorf(lw.appCtx, "Scanner error for %s: %v", logFile.Path, err)
	}

	// Update position to where the scanner left off.
	if pos, err := file.Seek(0, io.SeekCurrent); err == nil {
		logFile.LastPosition = pos
		logFile.LastModTime = info.ModTime()
		logFile.LastSize = currentSize
	}
}

// detectLogLevel infers a log level from the content of a line.
func detectLogLevel(line string) string {
	lower := strings.ToLower(line)
	switch {
	case containsAny(lower, "error", "err", "fatal", "critical", "exception"):
		return "error"
	case containsAny(lower, "warning", "warn"):
		return "warning"
	case containsAny(lower, "debug", "trace"):
		return "debug"
	case containsAny(lower, "success", "ok", "passed"):
		return "success"
	case containsAny(lower, "info", "information"):
		return "info"
	default:
		return "info"
	}
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

// matchesFilter reports whether level is in the filter list.
func matchesFilter(level string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, f := range filters {
		if strings.EqualFold(level, f) {
			return true
		}
	}
	return false
}
