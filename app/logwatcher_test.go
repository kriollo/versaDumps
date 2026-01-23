package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewLogWatcher tests creating a new LogWatcher
func TestNewLogWatcher(t *testing.T) {
	ctx := context.Background()

	watcher, err := NewLogWatcher(ctx)
	if err != nil {
		t.Fatalf("NewLogWatcher() failed: %v", err)
	}

	if watcher == nil {
		t.Fatal("NewLogWatcher() returned nil")
	}

	if watcher.ctx != ctx {
		t.Error("Context not set correctly")
	}

	if watcher.files == nil {
		t.Error("files map not initialized")
	}

	if watcher.folders == nil {
		t.Error("folders slice not initialized")
	}

	if watcher.running {
		t.Error("watcher should not be running initially")
	}

	// Clean up
	watcher.Stop()
}

// TestLogWatcher_StartStop tests starting and stopping the log watcher
func TestLogWatcher_StartStop(t *testing.T) {
	ctx := context.Background()

	watcher, err := NewLogWatcher(ctx)
	if err != nil {
		t.Fatalf("NewLogWatcher() failed: %v", err)
	}

	// Create temporary directory with log files
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	// Create a test log file
	f, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	f.WriteString("Test log line\n")
	f.Close()

	// Create log folder config
	folders := []LogFolder{
		{
			Path:       tempDir,
			Extensions: []string{"*.log"},
			Filters:    []string{"error"},
			Enabled:    true,
			Format:     "text",
		},
	}

	// Start watcher
	err = watcher.Start(folders)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	if !watcher.running {
		t.Error("Watcher should be running after Start()")
	}

	// Give it a moment to initialize
	time.Sleep(100 * time.Millisecond)

	// Stop watcher
	watcher.Stop()

	if watcher.running {
		t.Error("Watcher should not be running after Stop()")
	}
}

// TestLogWatcher_StartWithNoFolders tests starting with empty folder list
func TestLogWatcher_StartWithNoFolders(t *testing.T) {
	ctx := context.Background()

	watcher, err := NewLogWatcher(ctx)
	if err != nil {
		t.Fatalf("NewLogWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	// Start with empty folders
	err = watcher.Start([]LogFolder{})
	if err == nil {
		t.Error("Start() should return error when no folders provided")
	}

	if watcher.running {
		t.Error("Watcher should not be running when no folders provided")
	}
}

// TestLogWatcher_StartAlreadyRunning tests starting an already running watcher
func TestLogWatcher_StartAlreadyRunning(t *testing.T) {
	ctx := context.Background()

	watcher, err := NewLogWatcher(ctx)
	if err != nil {
		t.Fatalf("NewLogWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	// Create temporary directory with log files
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	f, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	f.Close()

	folders := []LogFolder{
		{
			Path:       tempDir,
			Extensions: []string{"*.log"},
			Enabled:    true,
			Format:     "text",
		},
	}

	// Start watcher first time
	err = watcher.Start(folders)
	if err != nil {
		t.Fatalf("First Start() failed: %v", err)
	}

	// Try to start again
	err = watcher.Start(folders)
	if err == nil {
		t.Error("Second Start() should return error when already running")
	}
}

// TestLogWatcher_FindMatchingFiles tests finding files matching patterns
func TestLogWatcher_FindMatchingFiles(t *testing.T) {
	ctx := context.Background()

	watcher, err := NewLogWatcher(ctx)
	if err != nil {
		t.Fatalf("NewLogWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	// Create temporary directory with various files
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"app.log",
		"error.log",
		"debug.txt",
		"info.txt",
		"readme.md",
		"data.json",
	}

	for _, filename := range testFiles {
		f, err := os.Create(filepath.Join(tempDir, filename))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		f.Close()
	}

	// Test pattern matching
	testCases := []struct {
		name          string
		extensions    []string
		expectedCount int
		expectedFiles []string
	}{
		{
			name:          "Match .log files",
			extensions:    []string{"*.log"},
			expectedCount: 2,
			expectedFiles: []string{"app.log", "error.log"},
		},
		{
			name:          "Match .txt files",
			extensions:    []string{"*.txt"},
			expectedCount: 2,
			expectedFiles: []string{"debug.txt", "info.txt"},
		},
		{
			name:          "Match multiple patterns",
			extensions:    []string{"*.log", "*.txt"},
			expectedCount: 4,
			expectedFiles: []string{"app.log", "error.log", "debug.txt", "info.txt"},
		},
		{
			name:          "Match all files",
			extensions:    []string{"*"},
			expectedCount: 6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folder := LogFolder{
				Path:       tempDir,
				Extensions: tc.extensions,
				Enabled:    true,
				Format:     "text",
			}

			files, err := watcher.findMatchingFiles(folder)
			if err != nil {
				t.Fatalf("findMatchingFiles() failed: %v", err)
			}

			if len(files) != tc.expectedCount {
				t.Errorf("Expected %d files, got %d", tc.expectedCount, len(files))
			}

			// Verify expected files are present
			if tc.expectedFiles != nil {
				fileMap := make(map[string]bool)
				for _, f := range files {
					fileMap[filepath.Base(f)] = true
				}

				for _, expectedFile := range tc.expectedFiles {
					if !fileMap[expectedFile] {
						t.Errorf("Expected file %s not found in results", expectedFile)
					}
				}
			}
		})
	}
}

// TestDetectLogLevel tests log level detection
func TestDetectLogLevel(t *testing.T) {
	testCases := []struct {
		line          string
		expectedLevel string
	}{
		{"[ERROR] Something went wrong", "error"},
		{"ERROR: Database connection failed", "error"},
		{"Fatal error occurred", "error"},
		{"[WARNING] Disk space running low", "warning"},
		{"WARN: Deprecated function used", "warning"},
		{"[INFO] Application started successfully", "info"},
		{"Information: User logged in", "info"},
		{"[DEBUG] Processing request", "debug"},
		{"Trace: Method called with params", "debug"},
		{"[SUCCESS] Operation completed", "success"},
		{"OK: All tests passed", "success"},
		{"Normal log line without level", "info"},
	}

	for _, tc := range testCases {
		t.Run(tc.line, func(t *testing.T) {
			level := detectLogLevel(tc.line)
			if level != tc.expectedLevel {
				t.Errorf("For line '%s', expected level '%s', got '%s'",
					tc.line, tc.expectedLevel, level)
			}
		})
	}
}

// TestMatchesFilter tests filter matching
func TestMatchesFilter(t *testing.T) {
	testCases := []struct {
		name     string
		level    string
		filters  []string
		expected bool
	}{
		{
			name:     "Empty filters - match all",
			level:    "error",
			filters:  []string{},
			expected: true,
		},
		{
			name:     "Match error",
			level:    "error",
			filters:  []string{"error", "warning"},
			expected: true,
		},
		{
			name:     "Match warning",
			level:    "warning",
			filters:  []string{"error", "warning"},
			expected: true,
		},
		{
			name:     "No match",
			level:    "info",
			filters:  []string{"error", "warning"},
			expected: false,
		},
		{
			name:     "Case insensitive match",
			level:    "ERROR",
			filters:  []string{"error"},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matchesFilter(tc.level, tc.filters)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestLogWatcher_RegisterFile tests file registration
func TestLogWatcher_RegisterFile(t *testing.T) {
	ctx := context.Background()

	watcher, err := NewLogWatcher(ctx)
	if err != nil {
		t.Fatalf("NewLogWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	// Create temporary log file
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	f, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	f.WriteString("Line 1\nLine 2\nLine 3\n")
	f.Close()

	// Register file
	err = watcher.registerFile(logFile)
	if err != nil {
		t.Fatalf("registerFile() failed: %v", err)
	}

	// Verify file is registered
	if _, exists := watcher.files[logFile]; !exists {
		t.Error("File not found in watcher.files map")
	}

	logFileData := watcher.files[logFile]
	if logFileData.Path != logFile {
		t.Errorf("Path mismatch: expected %s, got %s", logFile, logFileData.Path)
	}

	// Verify initial position is at end of file (tail -f behavior)
	info, _ := os.Stat(logFile)
	if logFileData.LastPosition != info.Size() {
		t.Errorf("Initial position should be at end of file. Expected %d, got %d",
			info.Size(), logFileData.LastPosition)
	}
}

// TestLogWatcher_AddFolder tests adding a folder
func TestLogWatcher_AddFolder(t *testing.T) {
	ctx := context.Background()

	watcher, err := NewLogWatcher(ctx)
	if err != nil {
		t.Fatalf("NewLogWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	// Create temporary directory with log files
	tempDir := t.TempDir()

	// Create test log files
	for i := 1; i <= 3; i++ {
		f, err := os.Create(filepath.Join(tempDir, "test"+string(rune('0'+i))+".log"))
		if err != nil {
			t.Fatalf("Failed to create test log file: %v", err)
		}
		f.Close()
	}

	folder := LogFolder{
		Path:       tempDir,
		Extensions: []string{"*.log"},
		Enabled:    true,
		Format:     "text",
	}

	err = watcher.addFolder(folder)
	if err != nil {
		t.Fatalf("addFolder() failed: %v", err)
	}

	// Verify files are registered
	if len(watcher.files) != 3 {
		t.Errorf("Expected 3 registered files, got %d", len(watcher.files))
	}
}

// TestLogWatcher_AddFolder_NonExistent tests adding a non-existent folder
func TestLogWatcher_AddFolder_NonExistent(t *testing.T) {
	ctx := context.Background()

	watcher, err := NewLogWatcher(ctx)
	if err != nil {
		t.Fatalf("NewLogWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	folder := LogFolder{
		Path:       "/non/existent/path",
		Extensions: []string{"*.log"},
		Enabled:    true,
		Format:     "text",
	}

	err = watcher.addFolder(folder)
	if err == nil {
		t.Error("addFolder() should return error for non-existent folder")
	}
}

// BenchmarkDetectLogLevel benchmarks log level detection
func BenchmarkDetectLogLevel(b *testing.B) {
	lines := []string{
		"[ERROR] Something went wrong",
		"[WARNING] Disk space running low",
		"[INFO] Application started",
		"[DEBUG] Processing request",
		"Normal log line",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		line := lines[i%len(lines)]
		detectLogLevel(line)
	}
}

// BenchmarkMatchesFilter benchmarks filter matching
func BenchmarkMatchesFilter(b *testing.B) {
	filters := []string{"error", "warning", "critical"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchesFilter("error", filters)
	}
}
