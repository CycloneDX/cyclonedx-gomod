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
	"flag"
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
)

type OptionsValidationError struct {
	Errors []error
}

func (e OptionsValidationError) Error() string {
	err := "invalid options:\n"
	for _, e := range e.Errors {
		err += fmt.Sprintf(" - %s\n", e)
	}
	return err
}

type Options interface {
	RegisterFlags(fs *flag.FlagSet)
	Validate() error
}

type OutputOptions struct {
	FilePath string
	UseJSON  bool
}

func (o *OutputOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.UseJSON, "json", false, "Output in JSON")
	fs.StringVar(&o.FilePath, "output", "-", "Output file path (or - for STDOUT)")
}

func (o OutputOptions) Validate() error {
	return nil
}

type SBOMOptions struct {
	ComponentType   string
	IncludeStd      bool
	IncludeTest     bool
	NoSerialNumber  bool
	NoVersionPrefix bool
	Reproducible    bool
	SerialNumber    string
}

func (s *SBOMOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&s.ComponentType, "type", "application", "Type of the main component")
	fs.BoolVar(&s.IncludeStd, "std", false, "Include Go standard library as component and dependency of the module")
	fs.BoolVar(&s.IncludeTest, "test", false, "Include test dependencies")
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
		return &OptionsValidationError{Errors: errs}
	}

	return nil
}
