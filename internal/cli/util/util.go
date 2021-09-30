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
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func AddCommonMetadata(bom *cdx.BOM, sbomOptions options.SBOMOptions) error {
	if sbomOptions.Reproducible {
		return nil
	}

	if bom.Metadata == nil {
		bom.Metadata = &cdx.Metadata{}
	}

	tool, err := sbom.BuildToolMetadata()
	if err != nil {
		return fmt.Errorf("failed to build tool metadata: %w", err)
	}

	bom.Metadata.Timestamp = time.Now().Format(time.RFC3339)
	bom.Metadata.Tools = &[]cdx.Tool{*tool}

	return nil
}

func AddStdComponent(bom *cdx.BOM, goVersion string) error {
	log.Debug().
		Msg("adding std component")

	stdComponent, err := sbom.BuildStdComponent(goVersion)
	if err != nil {
		return fmt.Errorf("failed to build std component: %w", err)
	}

	// Append std to components
	*bom.Components = append(*bom.Components, *stdComponent)

	// Add std to dependency graph
	stdDependency := cdx.Dependency{Ref: stdComponent.BOMRef}
	*bom.Dependencies = append(*bom.Dependencies, stdDependency)

	// Add std as dependency of main module
	for i, dependency := range *bom.Dependencies {
		if dependency.Ref == bom.Metadata.Component.BOMRef {
			if dependency.Dependencies == nil {
				(*bom.Dependencies)[i].Dependencies = &[]cdx.Dependency{stdDependency}
			} else {
				*dependency.Dependencies = append(*dependency.Dependencies, stdDependency)
			}
			break
		}
	}

	return nil
}

// SetSerialNumber sets the serial number of a given BOM according to the provided SBOMOptions.
func SetSerialNumber(bom *cdx.BOM, sbomOptions options.SBOMOptions) error {
	if sbomOptions.NoSerialNumber {
		return nil
	}

	if sbomOptions.SerialNumber == "" {
		bom.SerialNumber = uuid.New().URN()
	} else {
		serial, err := uuid.Parse(sbomOptions.SerialNumber)
		if err != nil {
			return err
		}
		bom.SerialNumber = serial.URN()
	}

	return nil
}

// WriteBOM writes the given bom according to the provided OutputOptions.
func WriteBOM(bom *cdx.BOM, outputOptions options.OutputOptions) error {
	log.Debug().
		Str("output", outputOptions.OutputFilePath).
		Bool("json", outputOptions.UseJSON).
		Msg("writing sbom")

	var outputFormat cdx.BOMFileFormat
	if outputOptions.UseJSON {
		outputFormat = cdx.BOMFileFormatJSON
	} else {
		outputFormat = cdx.BOMFileFormatXML
	}

	var outputWriter io.Writer
	if outputOptions.OutputFilePath == "" || outputOptions.OutputFilePath == "-" {
		outputWriter = os.Stdout
	} else {
		outputFile, err := os.Create(outputOptions.OutputFilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputOptions.OutputFilePath, err)
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
