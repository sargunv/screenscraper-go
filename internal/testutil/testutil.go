package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// ROMsPath returns the absolute path to the testroms directory.
// It works regardless of which package calls it by finding the workspace root
// (by looking for go.mod) and then navigating to testroms/.
func ROMsPath(t *testing.T) string {
	t.Helper()

	// Get the directory of the calling test file
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("Failed to get test file path")
	}
	testDir := filepath.Dir(filename)

	// Find workspace root by looking for go.mod
	dir := testDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Failed to find workspace root (go.mod not found)")
		}
		dir = parent
	}

	// Navigate to testroms from workspace root
	testromsPath := filepath.Join(dir, "testroms")

	// Resolve to absolute path
	absPath, err := filepath.Abs(testromsPath)
	if err != nil {
		t.Fatalf("Failed to resolve testroms path: %v", err)
	}

	// Verify it exists
	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("testroms directory not found at %s: %v", absPath, err)
	}

	return absPath
}
