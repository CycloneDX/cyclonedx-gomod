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

// Package testutil provides utility functions for tests.
package testutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Redacted marks an intentionally removed field.
const Redacted = "REDACTED"

// SilentLogger discards all inputs.
var SilentLogger = zerolog.Nop()

// ExtractFixtureArchive extracts a test fixture's TAR archive to a temporary directory and returns its path.
func ExtractFixtureArchive(t *testing.T, archivePath string) string {
	tmpDir := t.TempDir()

	cmd := exec.Command("tar", "xzf", archivePath, "-C", tmpDir) // #nosec G204
	out, err := cmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("extraction error: %s\n", string(out))
	}

	return tmpDir
}

// RequireMatchingPropertyToBeRedacted ensures that a given property is present and its value matched the provided regex.
//
// If a property matches, its value is then replaced by the string "REDACTED".
// This is intended to be used for properties that hold dynamic values that are expected to differ
// from system to system. Only use if absolutely necessary!
func RequireMatchingPropertyToBeRedacted(t *testing.T, properties []cdx.Property, name, valueRegex string) {
	for i, property := range properties {
		if property.Name == name {
			require.Regexp(t, valueRegex, property.Value)
			properties[i].Value = Redacted
			return
		}
	}

	t.Fatalf("property %s does not exist", name)
}

// RequireStdlibComponentToBeRedacted ensures that a stdlib component is present and redacts its version.
//
// Version will be redacted from packages as well, if applicable.
// If files are expected, their correlating components will be removed and replaced by an empty slice.
func RequireStdlibComponentToBeRedacted(t *testing.T, bom *cdx.BOM, expectPackages, expectFiles bool) {
	var (
		version   string
		oldBOMRef string
		newBOMRef string
		newPURL   string
	)

	for i, component := range *bom.Components {
		if component.Name == "std" {
			require.Regexp(t, `^go1\.`, component.Version)

			version = component.Version
			oldBOMRef = component.BOMRef
			newBOMRef = strings.ReplaceAll((*bom.Components)[i].BOMRef, version, Redacted)
			newPURL = strings.ReplaceAll((*bom.Components)[i].PackageURL, version, Redacted)

			(*bom.Components)[i].Version = Redacted
			(*bom.Components)[i].BOMRef = newBOMRef
			(*bom.Components)[i].PackageURL = newPURL

			// Redact all packages and files, as they may differ from one go version to another.
			if component.Components != nil { // Redact version from packages as well
				for _, component2 := range *(*bom.Components)[i].Components {
					require.Equal(t, version, component2.Version)

					if expectFiles {
						require.NotNil(t, component2.Components, "stdlib is missing files")
					}
				}

				// Use an empty slice instead of nil, in order for this modification
				// to be somewhat visible in the snapshot file.
				(*bom.Components)[i].Components = &[]cdx.Component{}
			} else if expectPackages {
				t.Fatalf("stdlib is missing packages")
			}
			break
		}
	}
	if newPURL == "" && newBOMRef == "" {
		t.Fatalf("stdlib component not found")
	}

	for i, dependency := range *bom.Dependencies {
		if dependency.Ref == oldBOMRef { // Dependant
			(*bom.Dependencies)[i].Ref = newBOMRef
		} else if dependency.Dependencies != nil { // Dependencies
			for j, dependency2 := range *(*bom.Dependencies)[i].Dependencies {
				if dependency2 == oldBOMRef {
					(*(*bom.Dependencies)[i].Dependencies)[j] = newBOMRef
				}
			}
		}
	}
}

// RequireMatchingSBOMSnapshot encodes a BOM and compares it to the snapshot of a test case.
func RequireMatchingSBOMSnapshot(t *testing.T, snapShooter *cupaloy.Config, bom *cdx.BOM, fileFormat cdx.BOMFileFormat) {
	buf := new(bytes.Buffer)

	encoder := cdx.NewBOMEncoder(buf, fileFormat)
	encoder.SetPretty(true)
	err := encoder.Encode(bom)
	require.NoError(t, err)

	snapShooter.SnapshotT(t, buf.String())
}

// RequireValidSBOM encodes the BOM and validates it using the CycloneDX CLI.
func RequireValidSBOM(t *testing.T, bom *cdx.BOM, fileFormat cdx.BOMFileFormat) {
	var inputFormat string
	switch fileFormat {
	case cdx.BOMFileFormatJSON:
		inputFormat = "json"
	case cdx.BOMFileFormatXML:
		inputFormat = "xml"
	}

	bomFile, err := os.Create(filepath.Join(t.TempDir(), fmt.Sprintf("bom.%s", inputFormat)))
	require.NoError(t, err)
	defer func() {
		if err := bomFile.Close(); err != nil && err.Error() != "file already closed" {
			fmt.Printf("failed to close bom file: %v\n", err)
		}
	}()

	encoder := cdx.NewBOMEncoder(bomFile, fileFormat)
	encoder.SetPretty(true)
	err = encoder.Encode(bom)
	require.NoError(t, err)
	require.NoError(t, bomFile.Close())

	valCmd := exec.Command("cyclonedx", "validate", "--input-file", bomFile.Name(), "--input-format", inputFormat, "--input-version", "v1_6", "--fail-on-errors") //nolint:gosec // #nosec G204
	valOut, err := valCmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("validation error: %s\n", string(valOut))
		t.FailNow()
	}
}

// SkipIfShort skips the test if `go test` was launched using the -short flag.
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}
