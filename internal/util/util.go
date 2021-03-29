package util

import (
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsGoModule(path string) bool {
	return FileExists(filepath.Join(path, "go.mod"))
}
