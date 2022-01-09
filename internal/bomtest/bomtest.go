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

// Package bomtest provides utility functionality for testing generated BOMs.
package bomtest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RequireSnapshot compares the BOM with a previous snapshot.
func RequireSnapshot(t *testing.T, s *cupaloy.Config, bom *cdx.BOM) {
	buf := new(bytes.Buffer)
	requireEncode(t, bom, buf)
	s.SnapshotT(t, buf.String())
}

// RequireValid validates the BOM using the CycloneDX CLI.
func RequireValid(t *testing.T, bom *cdx.BOM) {
	bomFile, err := os.CreateTemp("", fmt.Sprintf("cyclonedx-gomod_%s_*.bom.xml", strings.ReplaceAll(t.Name(), "/", "_")))
	require.NoError(t, err)
	defer func() {
		bomFile.Close()
		os.Remove(bomFile.Name())
	}()

	requireEncode(t, bom, bomFile)

	cmd := exec.Command("cyclonedx", "validate", "--input-file", bomFile.Name(), "--input-format", "xml_v1_3", "--fail-on-errors")
	valOut, err := cmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("validation error: %s\n", string(valOut))
	}
}

func requireEncode(t *testing.T, bom *cdx.BOM, writer io.Writer) {
	encoder := cdx.NewBOMEncoder(writer, cdx.BOMFileFormatXML)
	encoder.SetPretty(true)

	err := encoder.Encode(bom)
	require.NoError(t, err)
}
