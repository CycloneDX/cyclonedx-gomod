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

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// ModOptions provides options for the `mod` command.
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
	fs := flag.NewFlagSet("cyclonedx-gomod mod", flag.ExitOnError)

	var options ModOptions
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "mod",
		ShortHelp:  "Generate SBOM for a module",
		ShortUsage: "cyclonedx-gomod mod [FLAGS...] [PATH]",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 1 {
				return flag.ErrHelp
			}
			if len(args) == 0 {
				options.ModuleDir = "."
			} else {
				options.ModuleDir = args[0]
			}

			return execModCmd(options)
		},
	}
}

func execModCmd(options ModOptions) error {
	if err := options.Validate(); err != nil {
		return err
	}

	var serial *uuid.UUID
	if !options.NoSerialNumber && options.SerialNumber != "" {
		serialUUID := uuid.MustParse(options.SerialNumber)
		serial = &serialUUID
	}

	bom, err := sbom.Generate(options.ModuleDir, sbom.GenerateOptions{
		ComponentType:   cdx.ComponentType(options.ComponentType),
		IncludeStdLib:   options.IncludeStd,
		IncludeTest:     options.IncludeTest,
		NoSerialNumber:  options.NoSerialNumber,
		NoVersionPrefix: options.NoVersionPrefix,
		Reproducible:    options.Reproducible,
		ResolveLicenses: options.ResolveLicenses,
		SerialNumber:    serial,
	})
	if err != nil {
		return fmt.Errorf("failed to generate sbom: %w", err)
	}

	return WriteBOM(bom, options.OutputOptions)
}
