package webhook_tester

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFileWithParentDirs_CreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Path with non-existent parent directories
	outputPath := filepath.Join(tmpDir, "reports", "subdir", "output.txt")

	// Verify parent directory doesn't exist
	parentDir := filepath.Dir(outputPath)
	if _, err := os.Stat(parentDir); !os.IsNotExist(err) {
		t.Fatalf("Expected directory %s to not exist", parentDir)
	}

	// Should create parent dirs and file
	f, err := createFileWithParentDirs(outputPath)
	if err != nil {
		t.Fatalf("createFileWithParentDirs failed: %v", err)
	}
	defer f.Close()

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to be created", outputPath)
	}
}
