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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"

	"github.com/google/uuid"
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

// ConfigureLogger configures the global logger according to LogOptions.
func (l LogOptions) ConfigureLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: os.Getenv("CI") != "",
	})

	if l.Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
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
	UseJSON        bool
}

func (o *OutputOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.UseJSON, "json", false, "Output in JSON")
	fs.StringVar(&o.OutputFilePath, "output", "-", "Output file path (or - for STDOUT)")
}

func (o OutputOptions) Validate() error {
	return nil // Nothing to validate
}

// SBOMOptions provides options for customizing the SBOM.
type SBOMOptions struct {
	IncludeStd      bool
	NoSerialNumber  bool
	ResolveLicenses bool
	SerialNumber    string

	// Make SBOM reproducible by omitting dynamic content.
	// Only used internally for testing.
	Reproducible bool
}

func (s *SBOMOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&s.IncludeStd, "std", false, "Include Go standard library as component and dependency of the module")
	fs.BoolVar(&s.NoSerialNumber, "noserial", false, "Omit serial number")
	// Reproducible is used for testing only and intentionally omitted here
	fs.BoolVar(&s.ResolveLicenses, "licenses", false, "Perform license detection")
	fs.StringVar(&s.SerialNumber, "serial", "", "Serial number")
}

func (s SBOMOptions) Validate() error {
	errs := make([]error, 0)

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
