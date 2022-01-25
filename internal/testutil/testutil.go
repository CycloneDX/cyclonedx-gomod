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

package testutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

	cmd := exec.Command("tar", "xzf", archivePath, "-C", tmpDir)
	out, err := cmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("extraction error: %s\n", string(out))
	}

	return tmpDir
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
	defer bomFile.Close()

	encoder := cdx.NewBOMEncoder(bomFile, fileFormat)
	encoder.SetPretty(true)
	err = encoder.Encode(bom)
	require.NoError(t, err)

	valCmd := exec.Command("cyclonedx", "validate", "--input-file", bomFile.Name(), "--input-format", inputFormat, "--fail-on-errors")
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
