package gomod

import (
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModule_Hash(t *testing.T) {
	// Download a specific version of a module
	cmd := exec.Command("go", "get", "github.com/google/uuid@v1.2.0")
	cmd.Dir = os.TempDir() // Just has to be outside of this module's directory to prevent modification of go.mod
	require.NoError(t, cmd.Run())

	// Locate the module on the file system
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	modDir := filepath.Join(gopath, "pkg", "mod", "github.com", "google", "uuid@v1.2.0")

	// Construct module instance
	module := Module{
		Path:    "github.com/google/uuid",
		Version: "v1.2.0",
		Dir:     modDir,
	}

	// Calculate a directory hash for the downloaded module
	hash, err := module.Hash()
	require.NoError(t, err)

	// The returned hash must match the one from sumdb
	// See https://sum.golang.org/lookup/github.com/google/uuid@v1.2.0
	require.Equal(t, "h1:qJYtXnJRWmpe7m/3XlyhrsLrEURqHRM2kxzoxXqyUDs=", hash)
}
