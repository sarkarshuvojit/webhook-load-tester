package webhook_tester

import (
	"os"
	"path/filepath"
)

// createFileWithParentDirs creates a file at the given path.
// If the parent directories do not exist, they will be created.
func createFileWithParentDirs(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return os.Create(path)
}
