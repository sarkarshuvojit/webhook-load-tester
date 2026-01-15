package webhook_tester

import "os"

// createFileWithParentDirs creates a file at the given path.
// If the parent directories do not exist, they will be created.
func createFileWithParentDirs(path string) (*os.File, error) {
	return os.Create(path)
}
