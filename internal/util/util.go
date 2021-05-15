package util

import (
	"go/build"
	"os"
	"path/filepath"
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

// StringSliceIndex determines the index of a given string in a given string slice.
func StringSliceIndex(haystack []string, needle string) int {
	for i := range haystack {
		if haystack[i] == needle {
			return i
		}
	}
	return -1
}
