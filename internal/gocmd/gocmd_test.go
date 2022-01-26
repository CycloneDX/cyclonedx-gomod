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
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	version, err := GetVersion(zerolog.New(io.Discard))
	require.NoError(t, err)
	require.Equal(t, runtime.Version(), version)
}

func TestParseVersion(t *testing.T) {
	t.Run("Release", func(t *testing.T) {
		v, err := ParseVersion("foobar: go1.16.10")
		require.NoError(t, err)
		require.Equal(t, "go1.16.10", v)
	})

	t.Run("Devel", func(t *testing.T) {
		v, err := ParseVersion("foobar: devel go1.18-36be0be Thu Dec 2 16:48:07 2021 +0000")
		require.NoError(t, err)
		require.Equal(t, "go1.18-36be0be", v)
	})

	t.Run("Failure", func(t *testing.T) {
		v, err := ParseVersion("foobar: notgo1.16.10")
		require.Error(t, err)
		require.Equal(t, "", v)
	})
}

func TestGetEnv(t *testing.T) {
	env, err := GetEnv(zerolog.New(io.Discard))
	require.NoError(t, err)

	require.Contains(t, env, "CGO_ENABLED")
	require.Contains(t, env, "GOARCH")
	require.Contains(t, env, "GOFLAGS")
	require.Contains(t, env, "GOOS")
	require.Contains(t, env, "GOVERSION")
}

func TestListModule(t *testing.T) {
	buf := new(bytes.Buffer)
	err := ListModule(zerolog.New(io.Discard), "../../", buf)
	require.NoError(t, err)

	mod := make(map[string]interface{})
	require.NoError(t, json.NewDecoder(buf).Decode(&mod))

	require.Equal(t, "github.com/CycloneDX/cyclonedx-gomod", mod["Path"])
	assert.Equal(t, true, mod["Main"])
}

func TestListModules(t *testing.T) {
	buf := new(bytes.Buffer)
	err := ListModules(zerolog.New(io.Discard), "../../", buf)
	require.NoError(t, err)

	mod := make(map[string]interface{})
	require.NoError(t, json.NewDecoder(buf).Decode(&mod))

	// Smoke test - is this really the module list?
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-gomod", mod["Path"])
	assert.Equal(t, true, mod["Main"])
}

func TestGetModuleGraph(t *testing.T) {
	buf := new(bytes.Buffer)
	err := GetModuleGraph(zerolog.New(io.Discard), "../../", buf)
	require.NoError(t, err)

	assert.Equal(t, 0, strings.Index(buf.String(), "github.com/CycloneDX/cyclonedx-gomod"))
}

func TestModWhy(t *testing.T) {
	buf := new(bytes.Buffer)
	err := ModWhy(zerolog.New(io.Discard), "../../", []string{"github.com/CycloneDX/cyclonedx-go"}, buf)
	require.NoError(t, err)

	require.Equal(t, `# github.com/CycloneDX/cyclonedx-go
github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/mod
github.com/CycloneDX/cyclonedx-go
`, buf.String())
}
