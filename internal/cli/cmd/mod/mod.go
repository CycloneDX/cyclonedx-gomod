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
	"errors"
	"flag"
	"fmt"
	"io"

	cdx "github.com/CycloneDX/cyclonedx-go"
	cliUtil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/sbom"
	modConv "github.com/CycloneDX/cyclonedx-gomod/pkg/sbom/convert/module"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
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
  $ cyclonedx-gomod mod -reproducible -test -output bom.xml ./cyclonedx-go`,
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

			options.LogOptions.ConfigureLogger()

			return Exec(options)
		},
	}
}

func Exec(options Options) error {
	err := options.Validate()
	if err != nil {
		return err
	}

	// Cheap trick to make Go download all required modules in the module graph
	// without modifying go.sum (as `go mod download` would do).
	err = gocmd.ModWhy(options.ModuleDir, []string{"github.com/CycloneDX/cyclonedx-gomod"}, io.Discard)
	if err != nil {
		return fmt.Errorf("failed to download modules: %w", err)
	}

	// Try to collect modules from vendor/ directory first and if that fails, use `go list`.
	modules, err := gomod.GetVendoredModules(options.ModuleDir, options.IncludeTest)
	if err != nil {
		if errors.Is(err, gomod.ErrNotVendoring) {
			modules, err = gomod.LoadModules(options.ModuleDir, options.IncludeTest)
			if err != nil {
				return fmt.Errorf("failed to collect modules: %w", err)
			}
		} else {
			return fmt.Errorf("failed to collect vendored modules: %w", err)
		}
	}

	if options.IncludeStd {
		stdlibModule, err := gomod.LoadStdlibModule()
		if err != nil {
			return fmt.Errorf("failed to load stdlib module: %w", err)
		}

		modules[0].Dependencies = append(modules[0].Dependencies, stdlibModule)
		modules = append(modules, *stdlibModule)
	}

	err = gomod.ApplyModuleGraph(options.ModuleDir, modules)
	if err != nil {
		return fmt.Errorf("failed to apply module graph: %w", err)
	}

	// Determine version of main module
	modules[0].Version, err = gomod.GetModuleVersion(modules[0].Dir)
	if err != nil {
		log.Warn().Err(err).Msg("failed to determine version of main module")
	}

	// Convert main module
	mainComponent, err := modConv.ToComponent(modules[0],
		modConv.WithComponentType(cdx.ComponentType(options.ComponentType)),
		modConv.WithLicenses(options.ResolveLicenses),
	)
	if err != nil {
		return fmt.Errorf("failed to convert main module: %w", err)
	}

	// Convert the other modules
	components, err := modConv.ToComponents(modules[1:],
		modConv.WithLicenses(options.ResolveLicenses),
		modConv.WithModuleHashes(),
	)
	if err != nil {
		return fmt.Errorf("failed to convert modules: %w", err)
	}

	bom := cdx.NewBOM()
	err = cliUtil.SetSerialNumber(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to set serial number: %w", err)
	}

	// Assemble metadata
	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}
	err = cliUtil.AddCommonMetadata(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to add common metadata: %w", err)
	}

	bom.Components = &components
	dependencyGraph := sbom.BuildDependencyGraph(modules)
	bom.Dependencies = &dependencyGraph

	if options.AssertLicenses {
		sbom.AssertLicenses(bom)
	}

	return cliUtil.WriteBOM(bom, options.OutputOptions)
}
