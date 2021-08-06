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

package mod

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithModuleHashes(t *testing.T) {
	// Download a specific version of a module
	cmd := exec.Command("go", "get", "github.com/google/uuid@v1.2.0")
	cmd.Dir = os.TempDir() // Just has to be outside of this module's directory to prevent modification of go.mod
	require.NoError(t, cmd.Run())

	// Locate the module on the file system
	modDir := filepath.Join(util.GetModuleCacheDir(), "github.com", "google", "uuid@v1.2.0")

	// Construct module instance
	module := gomod.Module{
		Path:    "github.com/google/uuid",
		Version: "v1.2.0",
		Dir:     modDir,
	}

	// Construct component which the hashes will be applied to
	component := new(cdx.Component)

	// Calculate hashes
	err := withModuleHashes()(module, component)
	require.NoError(t, err)
	require.NotNil(t, component.Hashes)

	// Check for expected hash
	hashes := *component.Hashes
	assert.Len(t, hashes, 1)
	assert.Equal(t, cdx.HashAlgoSHA256, hashes[0].Algorithm)
	assert.Equal(t, "a8962d5e72515a6a5eee6ff75e5ca1aec2eb11446a1d1336931ce8c57ab2503b", hashes[0].Value)
}
