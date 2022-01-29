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

package app

import (
	"errors"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CycloneDX/cyclonedx-gomod/internal/testutil"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect/standard"
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
			WithLicenseDetector(standard.NewDetector(zerolog.Nop())),
			WithLogger(testutil.SilentLogger),
			WithIncludeStdlib(true))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:CGO_ENABLED", `(0|1)`)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOARCH", runtime.GOARCH)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOOS", runtime.GOOS)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOVERSION", `^go1\.`)
		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleWithFiles", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludeFiles(true),
			WithIncludePackages(true),
			WithIncludeStdlib(true),
			WithLicenseDetector(standard.NewDetector(zerolog.Nop())),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:CGO_ENABLED", `(0|1)`)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOARCH", runtime.GOARCH)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOOS", runtime.GOOS)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOVERSION", `^go1\.`)
		testutil.RequireStdlibComponentToBeRedacted(t, bom, true, true)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleWithPackages", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludePackages(true),
			WithIncludeStdlib(true),
			WithLicenseDetector(standard.NewDetector(zerolog.Nop())),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:CGO_ENABLED", `(0|1)`)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOARCH", runtime.GOARCH)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOOS", runtime.GOOS)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOVERSION", `^go1\.`)
		testutil.RequireStdlibComponentToBeRedacted(t, bom, true, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleMultiCommandPURL", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-multi-command.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludeStdlib(true),
			WithLicenseDetector(standard.NewDetector(zerolog.Nop())),
			WithLogger(testutil.SilentLogger),
			WithMainDir("cmd/purl"))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:CGO_ENABLED", `(0|1)`)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOARCH", runtime.GOARCH)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOOS", runtime.GOOS)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOVERSION", `^go1\.`)
		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleMultiCommandUUID", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-multi-command.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludeStdlib(true),
			WithLicenseDetector(standard.NewDetector(zerolog.Nop())),
			WithLogger(testutil.SilentLogger),
			WithMainDir("cmd/uuid"))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:CGO_ENABLED", `(0|1)`)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOARCH", runtime.GOARCH)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOOS", runtime.GOOS)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOVERSION", `^go1\.`)
		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleVendor", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-vendor.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludeStdlib(true),
			WithLicenseDetector(standard.NewDetector(zerolog.Nop())),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:CGO_ENABLED", `(0|1)`)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOARCH", runtime.GOARCH)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOOS", runtime.GOOS)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOVERSION", `^go1\.`)
		testutil.RequireStdlibComponentToBeRedacted(t, bom, false, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleVendorWithFiles", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-vendor.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludeFiles(true),
			WithIncludePackages(true),
			WithIncludeStdlib(true),
			WithLicenseDetector(standard.NewDetector(zerolog.Nop())),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:CGO_ENABLED", `(0|1)`)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOARCH", runtime.GOARCH)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOOS", runtime.GOOS)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOVERSION", `^go1\.`)
		testutil.RequireStdlibComponentToBeRedacted(t, bom, true, true)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})

	t.Run("SimpleVendorWithPackages", func(t *testing.T) {
		fixturePath := testutil.ExtractFixtureArchive(t, "../testdata/simple-vendor.tar.gz")

		g, err := NewGenerator(fixturePath,
			WithIncludePackages(true),
			WithIncludeStdlib(true),
			WithLicenseDetector(standard.NewDetector(zerolog.Nop())),
			WithLogger(testutil.SilentLogger))
		require.NoError(t, err)

		bom, err := g.Generate()
		require.NoError(t, err)

		testutil.RequireValidSBOM(t, bom, cyclonedx.BOMFileFormatXML)

		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:CGO_ENABLED", `(0|1)`)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOARCH", runtime.GOARCH)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOOS", runtime.GOOS)
		testutil.RequireMatchingPropertyToBeRedacted(t, *bom.Metadata.Component.Properties, "cdx:gomod:build:env:GOVERSION", `^go1\.`)
		testutil.RequireStdlibComponentToBeRedacted(t, bom, true, false)
		testutil.RequireMatchingSBOMSnapshot(t, snapShooter, bom, cyclonedx.BOMFileFormatXML)
	})
}

func TestGenerator_CreateBuildProperties(t *testing.T) {
	g := generator{
		logger: zerolog.New(io.Discard),
	}

	origGoflags := os.Getenv("GOFLAGS")
	os.Setenv("GOFLAGS", "-tags=foo,bar")

	if origGoflags != "" {
		defer func() {
			os.Setenv("GOFLAGS", origGoflags)
		}()
	}

	properties, err := g.createBuildProperties()
	require.NoError(t, err)
	require.Len(t, properties, 6)

	expectedCgoEnabled := "1" // Cgo is enabled per default
	if cgo := os.Getenv("CGO_ENABLED"); cgo != "" {
		expectedCgoEnabled = cgo
	}

	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:CGO_ENABLED", Value: expectedCgoEnabled})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:GOARCH", Value: runtime.GOARCH})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:GOOS", Value: runtime.GOOS})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:GOVERSION", Value: runtime.Version()})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:tag", Value: "foo"})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:tag", Value: "bar"})
}
