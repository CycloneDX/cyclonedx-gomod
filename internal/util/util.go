package util

import (
	"os"
	"path/filepath"
	"strings"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsGoModule(path string) bool {
	return FileExists(filepath.Join(path, "go.mod"))
}

func StartsWith(str, with string) bool {
	return strings.Index(str, with) == 0
}
