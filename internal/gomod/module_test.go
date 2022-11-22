// This file is part of CycloneDX GoMod
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

package gomod

import (
	"bytes"
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
	cmd := exec.Command("go", "mod", "download", "github.com/google/uuid@v1.2.0")
	cmd.Dir = t.TempDir() // Just has to be outside of this module's directory to prevent modification of go.mod
	require.NoError(t, cmd.Run())

	// Locate the module on the file system
	modCacheDir, err := exec.Command("go", "env", "GOMODCACHE").Output()
	require.NoError(t, err)
	modDir := filepath.Join(string(bytes.TrimSpace(modCacheDir)), "github.com", "google", "uuid@v1.2.0")

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

func TestModule_BOMRef(t *testing.T) {

	module := Module{
		Path:    "github.com/CycloneDX/cyclonedx-go",
		Version: "v0.1.0",
	}
	assert.Equal(t, "pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0?type=module", module.BOMRef())
}

func TestModule_PackageURL(t *testing.T) {

	module := Module{
		Path:    "github.com/CycloneDX/cyclonedx-go",
		Version: "v0.1.0",
	}
	goos := os.Getenv("GOOS")
	goarch := os.Getenv("GOARCH")

	assert.Equal(t, "pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0?type=module&goos="+goos+"&goarch="+goarch, module.PackageURL())
}

func TestIsModule(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		require.True(t, IsModule("../../"))
	})

	t.Run("Negative", func(t *testing.T) {
		require.False(t, IsModule(t.TempDir()))
	})
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
	assert.Empty(t, modules[0].Version)
	assert.True(t, modules[0].Main)
	assert.False(t, modules[0].Vendored)

	assert.Equal(t, "github.com/davecgh/go-spew", modules[1].Path)
	assert.Equal(t, "v1.1.1", modules[1].Version)
	assert.False(t, modules[1].Main)
	assert.False(t, modules[1].Vendored)
}

func TestSortModules(t *testing.T) {
	modules := []Module{
		{
			Path:    "path",
			Version: "v1.3.2",
		},
		{
			Path:    "path",
			Version: "v1.2.3",
		},
		{
			Path:    "path/v2",
			Version: "v2.0.0",
			Main:    true,
		},
	}

	sortModules(modules)

	require.Equal(t, "v2.0.0", modules[0].Version) // main
	require.Equal(t, "v1.2.3", modules[1].Version)
	require.Equal(t, "v1.3.2", modules[2].Version)
}
