package util

import (
	"go/build"
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

// GetGoPath determines the GOPATH location.
func GetGoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return gopath
}

// GetModuleCacheDir determines the location of Go's module cache.
func GetModuleCacheDir() string {
	modCacheDir := os.Getenv("GOMODCACHE")
	if modCacheDir == "" {
		modCacheDir = filepath.Join(GetGoPath(), "pkg", "mod")
	}
	return modCacheDir
}

// StartsWith checks if a given string is prefixed with another.
func StartsWith(str, with string) bool {
	return strings.Index(str, with) == 0
}

// StringSliceIndex determines the index of a given string in a given string slice.
func StringSliceIndex(haystack []string, needle string) int {
	for i := range haystack {
		if haystack[i] == needle {
			return i
		}
	}
	return -1
}
