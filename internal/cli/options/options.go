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

	cdx "github.com/CycloneDX/cyclonedx-go"
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
	return nil
}

// SBOMOptions provides options for customizing the SBOM.
type SBOMOptions struct {
	ComponentType   string
	IncludeStd      bool
	NoSerialNumber  bool
	NoVersionPrefix bool
	Reproducible    bool
	SerialNumber    string
}

func (s *SBOMOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&s.ComponentType, "type", "application", "Type of the main component")
	fs.BoolVar(&s.IncludeStd, "std", false, "Include Go standard library as component and dependency of the module")
	fs.BoolVar(&s.NoSerialNumber, "noserial", false, "Omit serial number")
	fs.BoolVar(&s.NoVersionPrefix, "novprefix", false, "Omit \"v\" prefix from versions")
	fs.BoolVar(&s.Reproducible, "reproducible", false, "Make the SBOM reproducible by omitting dynamic content")
	fs.StringVar(&s.SerialNumber, "serial", "", "Serial number")
}

var allowedComponentTypes = []cdx.ComponentType{
	cdx.ComponentTypeApplication,
	cdx.ComponentTypeContainer,
	cdx.ComponentTypeDevice,
	cdx.ComponentTypeFile,
	cdx.ComponentTypeFirmware,
	cdx.ComponentTypeFramework,
	cdx.ComponentTypeLibrary,
	cdx.ComponentTypeOS,
}

func (s SBOMOptions) Validate() error {
	errs := make([]error, 0)

	isAllowedComponentType := false
	for i := range allowedComponentTypes {
		if allowedComponentTypes[i] == cdx.ComponentType(s.ComponentType) {
			isAllowedComponentType = true
			break
		}
	}
	if !isAllowedComponentType {
		errs = append(errs, fmt.Errorf("invalid component type: \"%s\"", s.ComponentType))
	}

	// Serial numbers must be valid UUIDs
	if !s.NoSerialNumber && s.SerialNumber != "" {
		if _, err := uuid.Parse(s.SerialNumber); err != nil {
			errs = append(errs, fmt.Errorf("invalid serial number: %w", err))
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}

	return nil
}
