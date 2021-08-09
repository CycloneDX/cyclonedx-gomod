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

package gocmd

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	version, err := GetVersion()
	require.NoError(t, err)
	require.Equal(t, runtime.Version(), version)
}

func TestGetModuleName(t *testing.T) {
	buf := new(bytes.Buffer)
	err := GetModule("../../", buf)
	require.NoError(t, err)

	mod := make(map[string]interface{})
	require.NoError(t, json.NewDecoder(buf).Decode(&mod))

	require.Equal(t, "github.com/CycloneDX/cyclonedx-gomod", mod["Path"])
	assert.Equal(t, true, mod["Main"])
}

func TestGetModuleList(t *testing.T) {
	buf := new(bytes.Buffer)
	err := ListModules("../../", buf)
	require.NoError(t, err)

	mod := make(map[string]interface{})
	require.NoError(t, json.NewDecoder(buf).Decode(&mod))

	// Smoke test - is this really the module list?
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-gomod", mod["Path"])
	assert.Equal(t, true, mod["Main"])
}

func TestGetModuleGraph(t *testing.T) {
	buf := new(bytes.Buffer)
	err := GetModuleGraph("../../", buf)
	require.NoError(t, err)

	assert.Equal(t, 0, strings.Index(buf.String(), "github.com/CycloneDX/cyclonedx-gomod"))
}

func TestModWhy(t *testing.T) {
	buf := new(bytes.Buffer)
	err := ModWhy("../../", []string{"github.com/CycloneDX/cyclonedx-go"}, buf)
	require.NoError(t, err)

	require.Equal(t, `# github.com/CycloneDX/cyclonedx-go
github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/bin
github.com/CycloneDX/cyclonedx-go
`, buf.String())
}
