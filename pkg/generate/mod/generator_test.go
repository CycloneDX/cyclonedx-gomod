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
	"errors"
	"path/filepath"
	"testing"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/require"

	"github.com/CycloneDX/cyclonedx-gomod/internal/testutil"
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
	testutil.SkipIfShort(t)

	snapShooter := cupaloy.NewDefaultConfig().
		WithOptions(cupaloy.SnapshotSubdirectory("./testdata/snapshots"))

	t.Run("Simple", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludeStdlib(true),
			WithLicenseDetection(true),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleMultiCommand", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-multi-command.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludeStdlib(true),
			WithLicenseDetection(true),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleLocal", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-local.tar.gz")

		g, err := NewGenerator(filepath.Join(fixturePath, "local"),
			WithIncludeStdlib(true),
			WithLicenseDetection(true),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	// Test with a "simple" module with only a few dependencies,
	// but as a subdirectory of a Git repository. The expectation is that the
	// (pseudo-) version is inherited from the repository of the parent dir.
	//
	// nested/
	// |-+ .git/
	// |-+ simple/
	//   |-+ go.mod
	//   |-+ go.sum
	//   |-+ main.go
	t.Run("SimpleNested", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-nested.tar.gz")

		g, err := NewGenerator(filepath.Join(fixturePath, "simple"),
			WithIncludeStdlib(true),
			WithLicenseDetection(true),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleNoDependencies", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-no-dependencies.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludeStdlib(true),
			WithLicenseDetection(true),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleVendor", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-vendor.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludeStdlib(true),
			WithLicenseDetection(true),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})
}
