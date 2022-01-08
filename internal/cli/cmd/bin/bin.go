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

package bin

import (
	"context"
	"flag"
	"fmt"

	cliUtil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate/bin"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod bin", flag.ExitOnError)

	var options Options
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "bin",
		ShortHelp:  "Generate SBOMs for binaries",
		ShortUsage: "cyclonedx-gomod bin [FLAGS...] BINARY_PATH",
		LongHelp: `Generate SBOMs for binaries.

Although the binary is never executed by cyclonedx-gomod, it must be executable.
This is a requirement by the "go version -m" command that is used to provide this functionality.

When license detection is enabled, all modules (including the main module) 
will be downloaded to the module cache using "go mod download".
For the download of the main module to work, its version has to be provided
via the -version flag.

Please note that data embedded in binaries shouldn't be trusted,
unless there's solid evidence that the binaries haven't been modified
since they've been built.

Example:
  $ cyclonedx-gomod bin -json -output acme-app-v1.0.0.bom.json -version v1.0.0 ./acme-app`,
		FlagSet: fs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments (expected 1, got %d)", len(args))
			}
			if len(args) == 1 {
				options.BinaryPath = args[0]
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

	generator, err := bin.NewGenerator(options.BinaryPath,
		bin.WithLogger(logger),
		bin.WithStdlib(options.IncludeStd),
		bin.WithLicenseResolution(options.ResolveLicenses),
		bin.WithVersionOverride(options.Version))
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
	err = cliUtil.AddCommonMetadata(logger, bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to add common metadata")
	}
	if options.AssertLicenses {
		sbom.AssertLicenses(bom)
	}

	return cliUtil.WriteBOM(bom, options.OutputOptions)
}
