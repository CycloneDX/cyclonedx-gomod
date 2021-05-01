package gomod

import (
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModule_Coordinates(t *testing.T) {
	module := Module{
		Path:    "path",
		Version: "version",
	}
	assert.Equal(t, "path@version", module.Coordinates())

	module.Version = ""
	assert.Equal(t, "path", module.Coordinates())
}

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

func TestModule_PackageURL(t *testing.T) {
	module := Module{
		Path:    "github.com/CycloneDX/cyclonedx-go",
		Version: "v0.1.0",
	}

	assert.Equal(t, "pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0", module.PackageURL())
}

func TestParseModules(t *testing.T) {
	modulesJSON := `{
        "Path": "github.com/CycloneDX/cyclonedx-go",
        "Main": true,
        "Dir": "/path/to/cyclonedx-go",
        "GoMod": "/path/to/cyclonedx-go/go.mod",
        "GoVersion": "1.14"
}
{
        "Path": "github.com/davecgh/go-spew",
        "Version": "v1.1.1",
        "Time": "2018-02-21T23:26:28Z",
        "Indirect": true,
        "Dir": "/path/to/go/pkg/mod/github.com/davecgh/go-spew@v1.1.1",
        "GoMod": "/path/to/go/pkg/mod/cache/download/github.com/davecgh/go-spew/@v/v1.1.1.mod"
}`

	modules, err := parseModules(strings.NewReader(modulesJSON))
	require.NoError(t, err)

	assert.Len(t, modules, 2)
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[0].Path)
	assert.True(t, modules[0].Main)
	assert.Equal(t, "github.com/davecgh/go-spew", modules[1].Path)
	assert.Equal(t, "v1.1.1", modules[1].Version)
	assert.False(t, modules[1].Main)
}

func TestParseVendoredModules(t *testing.T) {
	goModVendorOutput := `# github.com/bradleyjkemp/cupaloy/v2 v2.6.0
## explicit
github.com/bradleyjkemp/cupaloy/v2
github.com/bradleyjkemp/cupaloy/v2/internal
# github.com/davecgh/go-spew v1.1.1
github.com/davecgh/go-spew/spew`

	cwd, err := os.Getwd()
	require.NoError(t, err)

	modules, err := parseVendoredModules(cwd, strings.NewReader(goModVendorOutput))
	require.NoError(t, err)

	assert.Len(t, modules, 2)
	assert.Equal(t, "github.com/bradleyjkemp/cupaloy/v2", modules[0].Path)
	assert.Equal(t, "v2.6.0", modules[0].Version)
	assert.Equal(t, filepath.Join(cwd, "github.com/bradleyjkemp/cupaloy/v2"), modules[0].Dir)
	assert.Equal(t, "github.com/davecgh/go-spew", modules[1].Path)
	assert.Equal(t, "v1.1.1", modules[1].Version)
	assert.Equal(t, filepath.Join(cwd, "github.com/davecgh/go-spew"), modules[1].Dir)
}
