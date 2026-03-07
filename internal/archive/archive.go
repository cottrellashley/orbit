package archive

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Move moves a directory to the archive path with a timestamp suffix.
// Returns the new path.
func Move(sourcePath, archivePath string) (string, error) {
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("source directory does not exist: %s", sourcePath)
	}

	if err := os.MkdirAll(archivePath, 0755); err != nil {
		return "", fmt.Errorf("cannot create archive directory %s: %w", archivePath, err)
	}

	baseName := filepath.Base(sourcePath)
	timestamp := time.Now().Format("2006-01-02-1504")
	destName := fmt.Sprintf("%s-%s", baseName, timestamp)
	destPath := filepath.Join(archivePath, destName)

	// Ensure unique name if collision
	if _, err := os.Stat(destPath); err == nil {
		destName = fmt.Sprintf("%s-%s-%d", baseName, timestamp, time.Now().Unix())
		destPath = filepath.Join(archivePath, destName)
	}

	if err := os.Rename(sourcePath, destPath); err != nil {
		return "", fmt.Errorf("cannot move %s to %s: %w", sourcePath, destPath, err)
	}

	return destPath, nil
}
