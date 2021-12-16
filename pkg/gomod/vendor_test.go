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
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVendoredModules(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	t.Run("Simple", func(t *testing.T) {
		goModVendorOutput := "# github.com/CycloneDX/cyclonedx-go1 v0.1.0"

		modules, err := parseVendoredModules(cwd, strings.NewReader(goModVendorOutput))
		require.NoError(t, err)
		require.Len(t, modules, 1)

		assert.Equal(t, "github.com/CycloneDX/cyclonedx-go1", modules[0].Path)
		assert.Equal(t, "v0.1.0", modules[0].Version)
		assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go1"), modules[0].Dir)
		assert.True(t, modules[0].Vendored)
	})

	t.Run("Replacement PathVersion to PathVersion", func(t *testing.T) {
		goModVendorOutput := "# github.com/CycloneDX/cyclonedx-go v0.1.0 => github.com/nscuro/cyclonedx-go v0.1.1"

		modules, err := parseVendoredModules(cwd, strings.NewReader(goModVendorOutput))
		require.NoError(t, err)
		require.Len(t, modules, 1)

		assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[0].Path)
		assert.Equal(t, "v0.1.0", modules[0].Version)
		assert.Empty(t, modules[0].Dir)
		assert.False(t, modules[0].Vendored)
		assert.NotNil(t, modules[0].Replace)
		assert.Equal(t, "github.com/nscuro/cyclonedx-go", modules[0].Replace.Path)
		assert.Equal(t, "v0.1.1", modules[0].Replace.Version)
		assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[0].Replace.Dir)
		assert.True(t, modules[0].Replace.Vendored)
	})

	t.Run("Replacement Path to PathVersion", func(t *testing.T) {
		goModVendorOutput := "# github.com/CycloneDX/cyclonedx-go => github.com/nscuro/cyclonedx-go v0.1.1"

		modules, err := parseVendoredModules(cwd, strings.NewReader(goModVendorOutput))
		require.NoError(t, err)
		require.Len(t, modules, 1)

		assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[0].Path)
		assert.Empty(t, modules[0].Version)
		assert.Empty(t, modules[0].Dir)
		assert.False(t, modules[0].Vendored)
		assert.NotNil(t, modules[0].Replace)
		assert.Equal(t, "github.com/nscuro/cyclonedx-go", modules[0].Replace.Path)
		assert.Equal(t, "v0.1.1", modules[0].Replace.Version)
		assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[0].Replace.Dir)
		assert.True(t, modules[0].Replace.Vendored)
	})

	t.Run("Replacement PathVersion to Path", func(t *testing.T) {
		goModVendorOutput := "# github.com/CycloneDX/cyclonedx-go v0.1.0 => ../cyclonedx-go"

		modules, err := parseVendoredModules(cwd, strings.NewReader(goModVendorOutput))
		require.NoError(t, err)
		require.Len(t, modules, 1)

		assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[0].Path)
		assert.Equal(t, "v0.1.0", modules[0].Version)
		assert.Empty(t, modules[0].Dir)
		assert.False(t, modules[0].Vendored)
		assert.NotNil(t, modules[0].Replace)
		assert.Equal(t, "../cyclonedx-go", modules[0].Replace.Path)
		assert.Empty(t, modules[0].Replace.Version)
		assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[0].Replace.Dir)
		assert.True(t, modules[0].Replace.Vendored)
	})

	t.Run("Replacement Path to Path", func(t *testing.T) {
		goModVendorOutput := "# github.com/CycloneDX/cyclonedx-go => ../cyclonedx-go"

		modules, err := parseVendoredModules(cwd, strings.NewReader(goModVendorOutput))
		require.NoError(t, err)
		require.Len(t, modules, 1)

		assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[0].Path)
		assert.Empty(t, modules[0].Version)
		assert.Empty(t, modules[0].Dir)
		assert.False(t, modules[0].Vendored)
		assert.NotNil(t, modules[0].Replace)
		assert.Equal(t, "../cyclonedx-go", modules[0].Replace.Path)
		assert.Empty(t, modules[0].Replace.Version)
		assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[0].Replace.Dir)
		assert.True(t, modules[0].Replace.Vendored)
	})

	t.Run("Duplicates", func(t *testing.T) {
		goModVendorOutput := `
# github.com/CycloneDX/cyclonedx-go v0.1.0 => github.com/nscuro/cyclonedx-go v0.1.1
# github.com/CycloneDX/cyclonedx-go => github.com/nscuro/cyclonedx-go v0.1.1
# github.com/CycloneDX/cyclonedx-go v0.1.0 => ../cyclonedx-go
# github.com/CycloneDX/cyclonedx-go => ../cyclonedx-go`

		modules, err := parseVendoredModules(cwd, strings.NewReader(goModVendorOutput))
		require.NoError(t, err)
		require.Len(t, modules, 1)

		assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[0].Path)
		assert.Equal(t, "v0.1.0", modules[0].Version)
		assert.Empty(t, modules[0].Dir)
		assert.False(t, modules[0].Vendored)
		assert.NotNil(t, modules[0].Replace)
		assert.Equal(t, "github.com/nscuro/cyclonedx-go", modules[0].Replace.Path)
		assert.Equal(t, "v0.1.1", modules[0].Replace.Version)
		assert.Equal(t, filepath.Join(cwd, "vendor", "github.com/CycloneDX/cyclonedx-go"), modules[0].Replace.Dir)
		assert.True(t, modules[0].Replace.Vendored)
	})
}
