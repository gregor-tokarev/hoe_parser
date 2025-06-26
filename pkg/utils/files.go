package utils

import (
	"os"
	"path/filepath"
)

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// CreateDirIfNotExists creates a directory if it doesn't exist
func CreateDirIfNotExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// GetFileExtension returns the file extension
func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// GetFilenameWithoutExt returns filename without extension
func GetFilenameWithoutExt(filename string) string {
	name := filepath.Base(filename)
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}
