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
	"io"
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

// SilentLogger discards all inputs.
var SilentLogger = zerolog.New(io.Discard)

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
			properties[i].Value = "REDACTED"
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
		version string
		oldPURL string
		newPURL string
	)

	for i, component := range *bom.Components {
		if component.Name == "std" {
			require.Regexp(t, `^go1\.`, component.Version)

			version = component.Version
			oldPURL = component.PackageURL
			newPURL = strings.ReplaceAll((*bom.Components)[i].PackageURL, version, "REDACTED")

			(*bom.Components)[i].Version = "REDACTED"
			(*bom.Components)[i].BOMRef = newPURL
			(*bom.Components)[i].PackageURL = newPURL

			if component.Components != nil { // Redact version from packages as well
				for j, component2 := range *(*bom.Components)[i].Components {
					require.Equal(t, version, component2.Version)

					(*(*bom.Components)[i].Components)[j].Version = "REDACTED"
					(*(*bom.Components)[i].Components)[j].PackageURL = strings.ReplaceAll(component2.PackageURL, version, "REDACTED")

					// Redact all files, as they may differ from one go version to another.
					// It isn't worth the hassle to redact single fields for the time being.
					if component2.Components != nil {
						// Use an empty slice instead of null, in order for this modification
						// to be somewhat visible in the snapshot file.
						(*(*bom.Components)[i].Components)[j].Components = &[]cdx.Component{}
					} else if expectFiles {
						t.Fatalf("stdlib is missing files")
					}
				}
			} else if expectPackages {
				t.Fatalf("stdlib is missing packages")
			}

			break
		}
	}
	if newPURL == "" {
		t.Fatalf("stdlib component not found")
	}

	for i, dependency := range *bom.Dependencies {
		if dependency.Ref == oldPURL { // Dependant
			(*bom.Dependencies)[i].Ref = newPURL
		} else if dependency.Dependencies != nil { // Dependencies
			for j, dependency2 := range *(*bom.Dependencies)[i].Dependencies {
				if dependency2.Ref == oldPURL {
					(*(*bom.Dependencies)[i].Dependencies)[j].Ref = newPURL
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
	var (
		inputFormat   string
		fileExtension string
	)
	switch fileFormat {
	case cdx.BOMFileFormatJSON:
		fileExtension = "json"
		inputFormat = "json_v1_3"
	case cdx.BOMFileFormatXML:
		fileExtension = "xml"
		inputFormat = "xml_v1_3"
	}

	bomFile, err := os.Create(filepath.Join(t.TempDir(), fmt.Sprintf("bom.%s", fileExtension)))
	require.NoError(t, err)
	defer func() {
		if err := bomFile.Close(); err != nil {
			fmt.Printf("failed to close bom file: %v", err)
		}
	}()

	encoder := cdx.NewBOMEncoder(bomFile, fileFormat)
	encoder.SetPretty(true)
	err = encoder.Encode(bom)
	require.NoError(t, err)

	valCmd := exec.Command("cyclonedx", "validate", "--input-file", bomFile.Name(), "--input-format", inputFormat, "--fail-on-errors") // #nosec G204
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
