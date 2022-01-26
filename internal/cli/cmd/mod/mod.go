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

package mod

import (
	"context"
	"flag"
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/peterbourgon/ff/v3/ffcli"

	cliUtil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate/mod"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod mod", flag.ExitOnError)

	var options Options
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "mod",
		ShortHelp:  "Generate SBOMs for modules",
		ShortUsage: "cyclonedx-gomod mod [FLAGS...] [MODULE_PATH]",
		LongHelp: `Generate SBOMs for modules.

Examples:
  $ cyclonedx-gomod mod -licenses -type library -json -output bom.json ./cyclonedx-go
  $ cyclonedx-gomod mod -test -output bom.xml ./cyclonedx-go`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments (expected 1, got %d)", len(args))
			}
			if len(args) == 0 {
				options.ModuleDir = "."
			} else {
				options.ModuleDir = args[0]
			}

			return Exec(options)
		},
	}
}

func Exec(options Options) error {
	err := options.Validate()
	if err != nil {
		return err
	}

	logger := options.Logger()

	generator, err := mod.NewGenerator(options.ModuleDir,
		mod.WithLogger(logger),
		mod.WithComponentType(cdx.ComponentType(options.ComponentType)),
		mod.WithIncludeStdlib(options.IncludeStd),
		mod.WithIncludeTestModules(options.IncludeTest),
		mod.WithLicenseDetection(options.ResolveLicenses))
	if err != nil {
		return err
	}

	bom, err := generator.Generate()
	if err != nil {
		return err
	}

	err = cliUtil.SetSerialNumber(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to set serial number: %w", err)
	}
	err = cliUtil.AddCommonMetadata(logger, bom)
	if err != nil {
		return fmt.Errorf("failed to add common metadata: %w", err)
	}
	if options.AssertLicenses {
		sbom.AssertLicenses(bom)
	}

	return cliUtil.WriteBOM(bom, options.OutputOptions)
}
