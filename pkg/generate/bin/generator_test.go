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

package bin

import (
	"errors"
	"io"
	"runtime"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/CycloneDX/cyclonedx-gomod/internal/bomtest"
)

var (
	silentLogger = zerolog.New(io.Discard)
	snapShooter  = cupaloy.NewDefaultConfig().
			WithOptions(cupaloy.SnapshotSubdirectory("./testdata/snapshots"))
)

func TestNewGenerator(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		g, err := NewGenerator("")
		require.NoError(t, err)
		require.NotNil(t, g)
	})

	t.Run("OptionError", func(t *testing.T) {
		failOption := func(g *generator) error {
			return errors.New("test")
		}

		g, err := NewGenerator("", failOption)
		require.Nil(t, g)
		require.Error(t, err)
		require.Equal(t, "test", err.Error())
	})
}

func TestGenerator_Generate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		skipOnWindows(t)

		g, err := NewGenerator("./testdata/simple",
			WithLogger(silentLogger),
			WithIncludeStdlib(true))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		bomtest.RequireSnapshot(t, snapShooter, bom)
		bomtest.RequireValid(t, bom)
	})

	t.Run("SuccessGo1.18", func(t *testing.T) {
		skipOnWindows(t)

		g, err := NewGenerator("./testdata/simple1.18",
			WithLogger(silentLogger),
			WithIncludeStdlib(true))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		bomtest.RequireSnapshot(t, snapShooter, bom)
		bomtest.RequireValid(t, bom)
	})

	t.Run("BinaryDoesNotExist", func(t *testing.T) {
		g, err := NewGenerator("./testdata/doesNotExist", WithLogger(silentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.Nil(t, bom)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load build info")
	})
}

// Go does not recognize the test binaries as executable on Windows.
// Once we've migrated to Go 1.18's debug.BuildInfo, this can be removed.
func skipOnWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
}
