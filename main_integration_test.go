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
// Copyright (c) Niklas Düster. All Rights Reserved.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	itSnapshotter = cupaloy.NewDefaultConfig().
			WithOptions(cupaloy.SnapshotSubdirectory("./testdata/integration/snapshots"))

	// Prefix for temporary files and directories created during ITs
	tmpPrefix = version.Name + "_"

	// Serial number to use in order to keep generated SBOMs reproducible
	zeroUUID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
)

// Integration test with a "simple" module with only a few dependencies,
// no replacements and no vendoring.
func TestIntegrationSimple(t *testing.T) {
	runSnapshotIT(t, Options{
		ComponentType:   cdx.ComponentTypeLibrary,
		ModulePath:      "./testdata/integration/simple",
		ResolveLicenses: true,
		Reproducible:    true,
		SerialNumber:    &zeroUUID,
	})
}

// Integration test with a module that uses replacement with a local module.
func TestIntegrationLocal(t *testing.T) {
	runSnapshotIT(t, Options{
		ComponentType:   cdx.ComponentTypeLibrary,
		ModulePath:      "./testdata/integration/local",
		ResolveLicenses: true,
		Reproducible:    true,
		SerialNumber:    &zeroUUID,
	})
}

// Integration test with a module that doesn't have any dependencies.
func TestIntegrationNoDependencies(t *testing.T) {
	runSnapshotIT(t, Options{
		ComponentType:   cdx.ComponentTypeLibrary,
		ModulePath:      "./testdata/integration/no-dependencies",
		ResolveLicenses: true,
		Reproducible:    true,
		SerialNumber:    &zeroUUID,
	})
}

// Integration test with a module that doesn't have any module dependencies,
// but includes the Go standard library as component.
func TestIntegrationNoDependenciesWithStd(t *testing.T) {
	runSnapshotIT(t, Options{
		ComponentType:   cdx.ComponentTypeLibrary,
		IncludeStd:      true,
		ModulePath:      "./testdata/integration/no-dependencies",
		ResolveLicenses: true,
		Reproducible:    true,
		SerialNumber:    &zeroUUID,
	})
}

// Integration test with a "simple" module with only a few dependencies,
// no replacements and no vendoring.
func TestIntegrationVendored(t *testing.T) {
	runSnapshotIT(t, Options{
		ComponentType:   cdx.ComponentTypeLibrary,
		ModulePath:      "./testdata/integration/vendored",
		ResolveLicenses: true,
		Reproducible:    true,
		SerialNumber:    &zeroUUID,
	})
}

func runSnapshotIT(t *testing.T, options Options) {
	skipIfShort(t)

	bomFileExtension := ".xml"
	if options.UseJSON {
		bomFileExtension = ".json"
	}

	// Create a temporary file to write the SBOM to
	bomFile, err := os.CreateTemp("", tmpPrefix+t.Name()+"_*.bom"+bomFileExtension)
	require.NoError(t, err)
	defer os.Remove(bomFile.Name())
	require.NoError(t, bomFile.Close())

	// Generate the SBOM
	options.OutputPath = bomFile.Name()
	err = executeCommand(options)
	require.NoError(t, err)

	// Sanity check: Make sure the SBOM is valid
	assertValidSBOM(t, bomFile.Name())

	// Read SBOM and compare with snapshot
	bomFileContent, err := os.ReadFile(bomFile.Name())
	require.NoError(t, err)
	itSnapshotter.SnapshotT(t, string(bomFileContent))
}

func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}

func assertValidSBOM(t *testing.T, bomFilePath string) {
	inputFormat := "xml_v1_2"
	if strings.HasSuffix(bomFilePath, ".json") {
		inputFormat = "json_v1_2"
	}
	valCmd := exec.Command("cyclonedx", "validate", "--input-file", bomFilePath, "--input-format", inputFormat, "--fail-on-errors")
	valOut, err := valCmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("validation error: %s\n", string(valOut))
	}
}
