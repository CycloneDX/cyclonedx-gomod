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

package module

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/package-url/packageurl-go"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
)

type stubLicenseDetector struct {
	Err      error
	Licenses []cdx.License
}

func (d stubLicenseDetector) Detect(_, _, _ string) ([]cdx.License, error) {
	if d.Err != nil {
		return nil, d.Err
	}

	return d.Licenses, nil
}

func TestWithLicenses(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		component := cdx.Component{}
		detector := &stubLicenseDetector{
			Licenses: []cdx.License{
				{
					ID: "Apache-2.0",
				},
			},
		}

		err := WithLicenses(detector)(zerolog.Nop(), gomod.Module{Dir: t.TempDir()}, &component)
		require.NoError(t, err)
		require.NotNil(t, component.Evidence)
		require.NotNil(t, component.Evidence.Licenses)
		require.Len(t, *component.Evidence.Licenses, 1)
	})

	t.Run("NoLicenseFound", func(t *testing.T) {
		component := cdx.Component{}
		detector := &stubLicenseDetector{
			Licenses: []cdx.License{},
		}

		err := WithLicenses(detector)(zerolog.Nop(), gomod.Module{Dir: t.TempDir()}, &component)
		require.NoError(t, err)
		require.Nil(t, component.Evidence)
	})

	t.Run("ModuleNotInCache", func(t *testing.T) {
		component := cdx.Component{}
		detector := &stubLicenseDetector{}

		err := WithLicenses(detector)(zerolog.Nop(), gomod.Module{Dir: ""}, &component)
		require.NoError(t, err)
		require.Nil(t, component.Evidence)
	})

	t.Run("OtherError", func(t *testing.T) {
		component := cdx.Component{}
		detector := &stubLicenseDetector{
			Err: errors.New("test"),
		}

		err := WithLicenses(detector)(zerolog.Nop(), gomod.Module{Dir: t.TempDir()}, &component)
		require.Error(t, err)
		require.Nil(t, component.Evidence)
	})

	t.Run("Disabled", func(t *testing.T) {
		component := cdx.Component{}

		err := WithLicenses(nil)(zerolog.Nop(), gomod.Module{Dir: t.TempDir()}, &component)
		require.NoError(t, err)
		require.Nil(t, component.Evidence)
	})
}

func TestWithModuleHashes(t *testing.T) {
	// Download a specific version of a module
	cmd := exec.Command("go", "mod", "download", "github.com/google/uuid@v1.2.0")
	cmd.Dir = t.TempDir() // Just has to be outside of this module's directory to prevent modification of go.mod
	require.NoError(t, cmd.Run())

	// Locate the module on the file system
	modCacheDir, err := exec.Command("go", "env", "GOMODCACHE").Output()
	require.NoError(t, err)
	modDir := filepath.Join(string(bytes.TrimSpace(modCacheDir)), "github.com", "google", "uuid@v1.2.0")

	// Construct module instance
	module := gomod.Module{
		Path:    "github.com/google/uuid",
		Version: "v1.2.0",
		Dir:     modDir,
	}

	// Construct component which the hashes will be applied to
	component := new(cdx.Component)

	// Calculate hashes
	err = WithModuleHashes()(zerolog.Nop(), module, component)
	require.NoError(t, err)
	require.NotNil(t, component.Hashes)

	// Check for expected hash
	hashes := *component.Hashes
	assert.Len(t, hashes, 1)
	assert.Equal(t, cdx.HashAlgoSHA256, hashes[0].Algorithm)
	assert.Equal(t, "a8962d5e72515a6a5eee6ff75e5ca1aec2eb11446a1d1336931ce8c57ab2503b", hashes[0].Value)
}

func TestWithComponentType(t *testing.T) {
	module := gomod.Module{}
	component := cdx.Component{}

	err := WithComponentType(cdx.ComponentTypeContainer)(zerolog.Nop(), module, &component)
	require.NoError(t, err)
	require.Equal(t, cdx.ComponentTypeContainer, component.Type)
}

func TestWithScope(t *testing.T) {
	module := gomod.Module{}
	component := cdx.Component{}

	err := WithScope(cdx.ScopeExcluded)(zerolog.Nop(), module, &component)
	require.NoError(t, err)
	require.Equal(t, cdx.ScopeExcluded, component.Scope)
}

func TestWithTestScope(t *testing.T) {
	t.Run("TestOnly", func(t *testing.T) {
		module := gomod.Module{
			TestOnly: true,
		}
		component := cdx.Component{}

		err := WithTestScope(cdx.ScopeExcluded)(zerolog.Nop(), module, &component)
		require.NoError(t, err)
		require.Equal(t, cdx.ScopeExcluded, component.Scope)
	})

	t.Run("Not TestOnly", func(t *testing.T) {
		module := gomod.Module{
			TestOnly: false,
		}
		component := cdx.Component{}

		err := WithTestScope(cdx.ScopeExcluded)(zerolog.Nop(), module, &component)
		require.NoError(t, err)
		require.Equal(t, cdx.Scope(""), component.Scope)
	})
}

func TestToComponent(t *testing.T) {
	// To get value from "go env -json", cannot just use t.GetEnv() might return ""
	envMap, _ := gocmd.GetEnv(zerolog.Nop())
	goos := envMap["GOOS"]
	goarch := envMap["GOARCH"]

	t.Run("Success", func(t *testing.T) {
		module := gomod.Module{
			Path:    "path",
			Version: "version",
		}

		component, err := ToComponent(zerolog.Nop(), module)
		require.NoError(t, err)
		require.NotNil(t, component)

		qualifiers := packageurl.Qualifiers{
			{Key: "type", Value: "module"},
			{Key: "goos", Value: goos},
			{Key: "goarch", Value: goarch},
		}
		require.NoError(t, qualifiers.Normalize())

		require.Equal(t, "pkg:golang/path@version?type=module", component.BOMRef)
		require.Equal(t, cdx.ComponentTypeLibrary, component.Type)
		require.Equal(t, "path", component.Name)
		require.Equal(t, "version", component.Version)
		require.Equal(t, "pkg:golang/path@version?"+qualifiers.String(), component.PackageURL)
		require.Equal(t, cdx.ScopeRequired, component.Scope)
	})

	t.Run("With TestOnly", func(t *testing.T) {
		module := gomod.Module{
			Path:     "path",
			Version:  "version",
			TestOnly: true,
		}

		component, err := ToComponent(zerolog.Nop(), module)
		require.NoError(t, err)
		require.NotNil(t, component)

		qualifiers := packageurl.Qualifiers{
			{Key: "type", Value: "module"},
			{Key: "goos", Value: goos},
			{Key: "goarch", Value: goarch},
		}
		require.NoError(t, qualifiers.Normalize())

		require.Equal(t, "pkg:golang/path@version?type=module", component.BOMRef)
		require.Equal(t, cdx.ComponentTypeLibrary, component.Type)
		require.Equal(t, "path", component.Name)
		require.Equal(t, "version", component.Version)
		require.Equal(t, "pkg:golang/path@version?"+qualifiers.String(), component.PackageURL)
		require.Equal(t, cdx.ScopeOptional, component.Scope)
	})

	t.Run("With Replace", func(t *testing.T) {
		module := gomod.Module{
			Path:    "path",
			Version: "version",
			Replace: &gomod.Module{
				Path:    "pathReplace",
				Version: "versionReplace",
			},
		}

		component, err := ToComponent(zerolog.Nop(), module)
		require.NoError(t, err)
		require.NotNil(t, component)

		qualifiers := packageurl.Qualifiers{
			{Key: "type", Value: "module"},
			{Key: "goos", Value: goos},
			{Key: "goarch", Value: goarch},
		}
		require.NoError(t, qualifiers.Normalize())

		require.Equal(t, "pkg:golang/pathReplace@versionReplace?type=module", component.BOMRef)
		require.Equal(t, cdx.ComponentTypeLibrary, component.Type)
		require.Equal(t, "pathReplace", component.Name)
		require.Equal(t, "versionReplace", component.Version)
		require.Equal(t, "pkg:golang/pathReplace@versionReplace?"+qualifiers.String(), component.PackageURL)
		require.Equal(t, cdx.ScopeRequired, component.Scope)
	})

	t.Run("WithSum", func(t *testing.T) {
		module := gomod.Module{
			Path:    "path",
			Version: "version",
			Sum:     "h1:qJYtXnJRWmpe7m/3XlyhrsLrEURqHRM2kxzoxXqyUDs=",
		}

		component, err := ToComponent(zerolog.Nop(), module)
		require.NoError(t, err)
		require.NotNil(t, component)

		require.NotNil(t, component.Hashes)
		require.Equal(t, cdx.HashAlgoSHA256, (*component.Hashes)[0].Algorithm)
		require.Equal(t, "a8962d5e72515a6a5eee6ff75e5ca1aec2eb11446a1d1336931ce8c57ab2503b", (*component.Hashes)[0].Value)
	})
}

func TestResolveVCSURL(t *testing.T) {
	t.Run("GitHub", func(t *testing.T) {
		require.Equal(t, "https://github.com/CycloneDX/cyclonedx-go", resolveVCSURL("github.com/CycloneDX/cyclonedx-go"))
	})

	t.Run("GitHub with major version", func(t *testing.T) {
		assert.Equal(t, "https://github.com/CycloneDX/cyclonedx-go", resolveVCSURL("github.com/CycloneDX/cyclonedx-go/v2"))
		assert.Equal(t, "https://github.com/CycloneDX/cyclonedx-go", resolveVCSURL("github.com/CycloneDX/cyclonedx-go/v222"))
	})

	t.Run("gopkg.in variant 1", func(t *testing.T) {
		require.Equal(t, "https://github.com/go-playground/assert", resolveVCSURL("gopkg.in/go-playground/assert.v1"))
	})

	t.Run("gopkg.in variant 2", func(t *testing.T) {
		require.Equal(t, "https://github.com/go-check/check", resolveVCSURL("gopkg.in/check.v1"))
	})
}
