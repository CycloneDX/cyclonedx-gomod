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

package util

import (
	"fmt"
	"io"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
)

func WriteBOM(bom *cdx.BOM, options options.OutputOptions) error {
	var outputFormat cdx.BOMFileFormat
	if options.UseJSON {
		outputFormat = cdx.BOMFileFormatJSON
	} else {
		outputFormat = cdx.BOMFileFormatXML
	}

	var outputWriter io.Writer
	if options.OutputFilePath == "" || options.OutputFilePath == "-" {
		outputWriter = os.Stdout
	} else {
		outputFile, err := os.Create(options.OutputFilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", options.OutputFilePath, err)
		}
		defer outputFile.Close()
		outputWriter = outputFile
	}

	encoder := cdx.NewBOMEncoder(outputWriter, outputFormat)
	encoder.SetPretty(true)

	if err := encoder.Encode(bom); err != nil {
		return fmt.Errorf("failed to encode sbom: %w", err)
	}

	return nil
}
