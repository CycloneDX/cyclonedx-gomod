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

package options

import (
	"flag"
	"fmt"
	"os"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"

	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
)

// ValidationError represents a validation error for options.
// It can contain multiple errors with details about which validation
// operations failed. The Errors slice should never be empty.
type ValidationError struct {
	Errors []error
}

func (e ValidationError) Error() string {
	err := "invalid options:\n"
	for _, e := range e.Errors {
		err += fmt.Sprintf(" - %s\n", e)
	}
	return err
}

// LogOptions provides options for log customization.
type LogOptions struct {
	Verbose bool
}

// Logger returns a zerolog.Logger configured according to LogOptions.
func (l LogOptions) Logger() zerolog.Logger {
	logger := log.Output(zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: os.Getenv("CI") != "",
	})

	if l.Verbose {
		logger = logger.Level(zerolog.DebugLevel)
	} else {
		logger = logger.Level(zerolog.InfoLevel)
	}

	return logger
}

func (l *LogOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&l.Verbose, "verbose", false, "Enable verbose output")
}

func (l LogOptions) Validate() error {
	return nil // Nothing to validate
}

// OutputOptions provides options for customizing the output.
type OutputOptions struct {
	OutputFilePath string
	OutputVersion  string
	UseJSON        bool
}

func (o *OutputOptions) RegisterFlags(fs *flag.FlagSet) {
	versionChoices := []string{
		cdx.SpecVersion1_6.String(),
		cdx.SpecVersion1_5.String(),
		cdx.SpecVersion1_4.String(),
		cdx.SpecVersion1_3.String(),
		cdx.SpecVersion1_2.String(),
		cdx.SpecVersion1_1.String(),
		cdx.SpecVersion1_0.String(),
	}

	fs.BoolVar(&o.UseJSON, "json", false, "Output in JSON")
	fs.StringVar(&o.OutputFilePath, "output", "-", "Output file path (or - for STDOUT)")
	fs.StringVar(&o.OutputVersion, "output-version", cdx.SpecVersion1_6.String(),
		fmt.Sprintf("Output spec verson (%s)", strings.Join(versionChoices, ", ")))
}

func (o OutputOptions) Validate() error {
	if _, err := util.ParseSpecVersion(o.OutputVersion); err != nil {
		return &ValidationError{Errors: []error{err}}
	}

	return nil
}

// SBOMOptions provides options for customizing the SBOM.
type SBOMOptions struct {
	AssertLicenses  bool
	IncludeStd      bool
	NoSerialNumber  bool
	ResolveLicenses bool
	SerialNumber    string
}

func (s *SBOMOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&s.AssertLicenses, "assert-licenses", false, "Assert detected licenses")
	fs.BoolVar(&s.IncludeStd, "std", false, "Include Go standard library as component and dependency of the module")
	fs.BoolVar(&s.NoSerialNumber, "noserial", false, "Omit serial number")
	fs.BoolVar(&s.ResolveLicenses, "licenses", false, "Perform license detection")
	fs.StringVar(&s.SerialNumber, "serial", "", "Serial number")
}

func (s SBOMOptions) Validate() error {
	errs := make([]error, 0)

	if s.AssertLicenses && !s.ResolveLicenses {
		errs = append(errs, fmt.Errorf("assertion of licenses has no effect without licenses detection"))
	}

	// Serial numbers must be valid UUIDs
	if !s.NoSerialNumber && s.SerialNumber != "" {
		if _, err := uuid.Parse(s.SerialNumber); err != nil {
			errs = append(errs, fmt.Errorf("serial number: %w", err))
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}

	return nil
}
