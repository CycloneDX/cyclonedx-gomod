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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
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
	modDir := filepath.Join(util.GetModuleCacheDir(), "github.com", "google", "uuid@v1.2.0")

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
	assert.Empty(t, modules[0].Version)
	assert.True(t, modules[0].Main)
	assert.False(t, modules[0].Vendored)

	assert.Equal(t, "github.com/davecgh/go-spew", modules[1].Path)
	assert.Equal(t, "v1.1.1", modules[1].Version)
	assert.False(t, modules[1].Main)
	assert.False(t, modules[1].Vendored)
}

func TestParseVendoredModules(t *testing.T) {
	goModVendorOutput := `# github.com/CycloneDX/cyclonedx-go v0.1.0

# github.com/CycloneDX/cyclonedx-go v0.1.0 => github.com/nscuro/cyclonedx-go v0.1.1

# github.com/CycloneDX/cyclonedx-go => github.com/nscuro/cyclonedx-go v0.1.1

# github.com/CycloneDX/cyclonedx-go v0.1.0 => ../cyclonedx-go

# github.com/CycloneDX/cyclonedx-go => ../cyclonedx-go
`

	cwd, err := os.Getwd()
	require.NoError(t, err)

	modules, err := parseVendoredModules(cwd, strings.NewReader(goModVendorOutput))
	require.NoError(t, err)

	assert.Len(t, modules, 5)

	// Normal module
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[0].Path)
	assert.Equal(t, "v0.1.0", modules[0].Version)
	assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[0].Dir)
	assert.True(t, modules[0].Vendored)

	// Module with replacement: "Path Version => Path Version"
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[1].Path)
	assert.Equal(t, "v0.1.0", modules[1].Version)
	assert.Empty(t, modules[1].Dir)
	assert.False(t, modules[1].Vendored)
	assert.NotNil(t, modules[1].Replace)
	assert.Equal(t, "github.com/nscuro/cyclonedx-go", modules[1].Replace.Path)
	assert.Equal(t, "v0.1.1", modules[1].Replace.Version)
	assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[1].Replace.Dir)
	assert.True(t, modules[1].Replace.Vendored)

	// Module with replacement: "Path => Path Version"
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[1].Path)
	assert.Empty(t, modules[2].Version)
	assert.Empty(t, modules[2].Dir)
	assert.False(t, modules[2].Vendored)
	assert.NotNil(t, modules[2].Replace)
	assert.Equal(t, "github.com/nscuro/cyclonedx-go", modules[2].Replace.Path)
	assert.Equal(t, "v0.1.1", modules[2].Replace.Version)
	assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[2].Replace.Dir)
	assert.True(t, modules[2].Replace.Vendored)

	// Module with replacement: "Path Version => Path"
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[3].Path)
	assert.Equal(t, "v0.1.0", modules[3].Version)
	assert.Empty(t, modules[3].Dir)
	assert.False(t, modules[3].Vendored)
	assert.NotNil(t, modules[3].Replace)
	assert.Equal(t, "../cyclonedx-go", modules[3].Replace.Path)
	assert.Empty(t, modules[3].Replace.Version)
	assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[3].Replace.Dir)
	assert.True(t, modules[3].Replace.Vendored)

	// Module with replacement: "Path => Path"
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[4].Path)
	assert.Empty(t, modules[4].Version)
	assert.Empty(t, modules[4].Dir)
	assert.False(t, modules[4].Vendored)
	assert.NotNil(t, modules[4].Replace)
	assert.Equal(t, "../cyclonedx-go", modules[4].Replace.Path)
	assert.Empty(t, modules[4].Replace.Version)
	assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[4].Replace.Dir)
	assert.True(t, modules[4].Replace.Vendored)
}

func TestParseModWhy(t *testing.T) {
	modWhyOutput := `
# github.com/stretchr/testify
github.com/CycloneDX/cyclonedx-gomod
github.com/CycloneDX/cyclonedx-gomod.test
github.com/stretchr/testify/assert

# github.com/CycloneDX/cyclonedx-go
(main module does not need module github.com/CycloneDX/cyclonedx-go)

# bazil.org/fuse
(main module does not need to vendor module bazil.org/fuse)
`

	modulePkgs := parseModWhy(strings.NewReader(modWhyOutput))
	require.Len(t, modulePkgs, 3)

	assert.Len(t, modulePkgs["github.com/stretchr/testify"], 3)
	assert.Len(t, modulePkgs["github.com/CycloneDX/cyclonedx-go"], 0)
	assert.Len(t, modulePkgs["bazil.org/fuse"], 0)
}

func TestParseModulesFromBinary(t *testing.T) {
	cmdOutput := `minikube: go1.16.4
path    k8s.io/minikube/cmd/minikube
mod     k8s.io/minikube (devel) 
dep     cloud.google.com/go     v0.84.0 h1:hVhK90DwCdOAYGME/FJd9vNIZye9HBR6Yy3fu4js3N8=
dep     github.com/briandowns/spinner   v1.11.1
=>      github.com/alonyb/spinner       v1.12.7 h1:FflTMA9I2xRd8OQ5swyZY6Q1DFeaicA/bWo6/oM82a8=
`

	modules, hashes := parseModulesFromBinary(strings.NewReader(cmdOutput))
	require.Len(t, modules, 3)
	require.Len(t, hashes, 2)

	// Main module
	require.Equal(t, "k8s.io/minikube", modules[0].Path)
	require.Equal(t, "(devel)", modules[0].Version)
	require.True(t, modules[0].Main)
	require.NotContains(t, hashes, modules[0].Coordinates())

	// Module w/o replacement
	require.Equal(t, "cloud.google.com/go", modules[1].Path)
	require.Equal(t, "v0.84.0", modules[1].Version)
	require.Contains(t, hashes, modules[1].Coordinates())
	require.Equal(t, "h1:hVhK90DwCdOAYGME/FJd9vNIZye9HBR6Yy3fu4js3N8=", hashes["cloud.google.com/go@v0.84.0"])

	// Module with replacement
	require.Equal(t, "github.com/briandowns/spinner", modules[2].Path)
	require.Equal(t, "v1.11.1", modules[2].Version)
	require.NotContains(t, hashes, modules[2].Coordinates())
	require.NotNil(t, modules[2].Replace)

	// Replacement
	require.Equal(t, "github.com/alonyb/spinner", modules[2].Replace.Path)
	require.Equal(t, "v1.12.7", modules[2].Replace.Version)
	require.Contains(t, hashes, modules[2].Replace.Coordinates())
	require.Equal(t, "h1:FflTMA9I2xRd8OQ5swyZY6Q1DFeaicA/bWo6/oM82a8=", hashes["github.com/alonyb/spinner@v1.12.7"])
}

func TestFindModule(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		modules := make([]Module, 0)
		require.Nil(t, findModule(modules, "path@version", true))
	})

	t.Run("Strict", func(t *testing.T) {
		modules := []Module{
			{
				Path:    "path",
				Version: "version1",
			},
			{
				Path:    "path",
				Version: "version2",
			},
		}

		require.Nil(t, findModule(modules, "path@version0", true))
		require.Nil(t, findModule(modules, "otherpath@version1", true))

		module := findModule(modules, "path@version2", true)
		require.NotNil(t, module)
		require.Equal(t, "path@version2", module.Coordinates())
	})

	t.Run("NonStrict", func(t *testing.T) {
		modules := []Module{
			{
				Path:    "path",
				Version: "version1",
			},
			{
				Path:    "path",
				Version: "version2",
			},
		}

		module := findModule(modules, "path@version0", false)
		require.NotNil(t, module)
		require.Equal(t, "path@version1", module.Coordinates())

		require.Nil(t, findModule(modules, "otherpath@version1", false))

		require.Equal(t, module, findModule(modules, "path@version2", false))
	})
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

	SortModules(modules)

	require.Equal(t, "v2.0.0", modules[0].Version) // main
	require.Equal(t, "v1.2.3", modules[1].Version)
	require.Equal(t, "v1.3.2", modules[2].Version)
}

func TestSortDependencies(t *testing.T) {
	modules := []*Module{
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

	SortDependencies(modules)

	require.Equal(t, "v1.2.3", modules[0].Version)
	require.Equal(t, "v1.3.2", modules[1].Version)
	require.Equal(t, "v2.0.0", modules[2].Version) // main
}
