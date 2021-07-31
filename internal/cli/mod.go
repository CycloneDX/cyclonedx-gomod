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

package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type ModOptions struct {
	OutputOptions
	SBOMOptions

	ModuleDir       string
	ResolveLicenses bool
}

func (m *ModOptions) RegisterFlags(fs *flag.FlagSet) {
	m.OutputOptions.RegisterFlags(fs)
	m.SBOMOptions.RegisterFlags(fs)

	fs.BoolVar(&m.ResolveLicenses, "licenses", false, "Resolve module licenses")
}

func (m ModOptions) Validate() error {
	errs := make([]error, 0)

	if err := m.OutputOptions.Validate(); err != nil {
		var verr *OptionsValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}
	if err := m.SBOMOptions.Validate(); err != nil {
		var verr *OptionsValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}

	if len(errs) > 0 {
		return &OptionsValidationError{Errors: errs}
	}

	return nil
}

func newModCmd() *ffcli.Command {
	var modOptions ModOptions

	fs := flag.NewFlagSet("cyclonedx-gomod mod", flag.ExitOnError)
	modOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "mod",
		ShortHelp:  "Generate SBOM for a module",
		ShortUsage: "cyclonedx-gomod mod [FLAGS...] PATH",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 1 {
				return flag.ErrHelp
			}
			if len(args) == 0 {
				modOptions.ModuleDir = "."
			} else {
				modOptions.ModuleDir = args[0]
			}

			return execModCmd(modOptions)
		},
	}
}

func execModCmd(modOptions ModOptions) error {
	if err := modOptions.Validate(); err != nil {
		return err
	}

	var serial *uuid.UUID
	if modOptions.SerialNumber != "" {
		serialUUID := uuid.MustParse(modOptions.SerialNumber)
		serial = &serialUUID
	}

	bom, err := sbom.Generate(modOptions.ModuleDir, sbom.GenerateOptions{
		ComponentType:   cdx.ComponentType(modOptions.ComponentType),
		IncludeStdLib:   modOptions.IncludeStd,
		IncludeTest:     modOptions.IncludeTest,
		NoSerialNumber:  modOptions.NoSerialNumber,
		NoVersionPrefix: modOptions.NoVersionPrefix,
		Reproducible:    modOptions.Reproducible,
		ResolveLicenses: modOptions.ResolveLicenses,
		SerialNumber:    serial,
	})
	if err != nil {
		return fmt.Errorf("failed to generate sbom: %w", err)
	}

	var outputFormat cdx.BOMFileFormat
	if modOptions.UseJSON {
		outputFormat = cdx.BOMFileFormatJSON
	} else {
		outputFormat = cdx.BOMFileFormatXML
	}

	var outputWriter io.Writer
	if modOptions.FilePath == "" || modOptions.FilePath == "-" {
		outputWriter = os.Stdout
	} else {
		outputFile, err := os.Create(modOptions.FilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", modOptions.FilePath, err)
		}
		defer outputFile.Close()
		outputWriter = outputFile
	}

	encoder := cdx.NewBOMEncoder(outputWriter, outputFormat)
	encoder.SetPretty(true)

	if err = encoder.Encode(bom); err != nil {
		return fmt.Errorf("failed to encode sbom: %w", err)
	}

	return nil
}
