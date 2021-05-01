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

// IsGoModule determines whether the directory at the given path is a Go module.
func IsGoModule(path string) bool {
	return FileExists(filepath.Join(path, "go.mod"))
}

// IsVendoring determines whether of not the module at the given path is vendoring its dependencies.
// Should be used in conjunction with IsGoModule.
func IsVendoring(path string) bool {
	return FileExists(filepath.Join(path, "vendor", "modules.txt"))
}

func StartsWith(str, with string) bool {
	return strings.Index(str, with) == 0
}
