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

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	snapshotter = cupaloy.NewDefaultConfig().
			WithOptions(cupaloy.SnapshotSubdirectory("./testdata/snapshots"))

	// Prefix for temporary files and directories created during ITs
	tmpPrefix = version.Name + "_"

	// Serial number to use in order to keep generated SBOMs reproducible
	zeroUUID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
)

func runSnapshotIT(t *testing.T, outputOptions *options.OutputOptions, execFunc func() error) {
	skipIfShort(t)

	bomFileExtension := ".xml"
	if outputOptions.UseJSON {
		bomFileExtension = ".json"
	}

	// Create a temporary file to write the SBOM to
	bomFile, err := os.CreateTemp("", tmpPrefix+t.Name()+"_*.bom"+bomFileExtension)
	require.NoError(t, err)
	defer os.Remove(bomFile.Name())
	require.NoError(t, bomFile.Close())

	// Generate the SBOM
	outputOptions.OutputFilePath = bomFile.Name()
	err = execFunc()
	require.NoError(t, err)

	// Sanity check: Make sure the SBOM is valid
	assertValidSBOM(t, bomFile.Name())

	// Read SBOM and compare with snapshot
	bomFileContent, err := os.ReadFile(bomFile.Name())
	require.NoError(t, err)
	snapshotter.SnapshotT(t, string(bomFileContent))
}

func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}

func assertValidSBOM(t *testing.T, bomFilePath string) {
	inputFormat := "xml_v1_3"
	if strings.HasSuffix(bomFilePath, ".json") {
		inputFormat = "json_v1_3"
	}
	valCmd := exec.Command("cyclonedx", "validate", "--input-file", bomFilePath, "--input-format", inputFormat, "--fail-on-errors")
	valOut, err := valCmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("validation error: %s\n", string(valOut))
	}
}

func extractFixture(t *testing.T, archivePath string) string {
	tmpDir, err := os.MkdirTemp("", tmpPrefix+t.Name()+"_*")
	require.NoError(t, err)

	cmd := exec.Command("tar", "xzf", archivePath, "-C", tmpDir)
	out, err := cmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("validation error: %s\n", string(out))
	}

	return tmpDir
}
