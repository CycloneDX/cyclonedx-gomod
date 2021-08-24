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
