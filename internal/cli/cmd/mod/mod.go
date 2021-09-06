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
	cliutil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modconv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod mod", flag.ExitOnError)

	var options ModOptions
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "mod",
		ShortHelp:  "Generate SBOM for a module",
		ShortUsage: "cyclonedx-gomod mod [FLAGS...] [PATH]",
		LongHelp: `Generate SBOM for a module.

Examples:
  $ cyclonedx-gomod mod -licenses -type library -json -output bom.json ./cyclonedx-go
  $ cyclonedx-gomod mod -reproducible -test -output bom.xml ./cyclonedx-go`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 1 {
				return flag.ErrHelp
			}
			if len(args) == 0 {
				options.ModuleDir = "."
			} else {
				options.ModuleDir = args[0]
			}

			cliutil.ConfigureLogger(options.LogOptions)

			return Exec(options)
		},
	}
}

func Exec(options ModOptions) error {
	err := options.Validate()
	if err != nil {
		return err
	}

	// Cheap trick to make Go download all required modules in the module graph
	// without modifying go.sum (as `go mod download` would do).
	err = gocmd.ModWhy(options.ModuleDir, []string{"github.com/CycloneDX/cyclonedx-gomod"}, io.Discard)
	if err != nil {
		return fmt.Errorf("downloading modules failed: %w", err)
	}

	modules, err := gomod.GetVendoredModules(options.ModuleDir, options.IncludeTest)
	if err != nil {
		if errors.Is(err, gomod.ErrNotVendoring) {
			modules, err = gomod.GetModules(options.ModuleDir, options.IncludeTest)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err = gomod.ApplyModuleGraph(options.ModuleDir, modules)
	if err != nil {
		return fmt.Errorf("failed to apply module graph: %w", err)
	}

	modules[0].Version, err = gomod.GetModuleVersion(modules[0].Dir)
	if err != nil {
		log.Warn().Err(err).Msg("failed to determine version of main module")
	}

	mainComponent, err := modconv.ToComponent(modules[0],
		modconv.WithComponentType(cdx.ComponentType(options.ComponentType)),
		modconv.WithLicenses(options.ResolveLicenses),
	)
	if err != nil {
		return fmt.Errorf("failed to convert main module: %w", err)
	}

	components, err := modconv.ToComponents(modules[1:],
		modconv.WithLicenses(options.ResolveLicenses),
		modconv.WithModuleHashes(),
	)
	if err != nil {
		return fmt.Errorf("failed to convert modules: %w", err)
	}

	dependencyGraph := sbom.BuildDependencyGraph(modules)

	bom := cdx.NewBOM()

	err = cliutil.SetSerialNumber(bom, options.SBOMOptions)
	if err != nil {
		return err
	}

	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}

	err = cliutil.AddCommonMetadata(bom, options.SBOMOptions)
	if err != nil {
		return err
	}

	bom.Components = &components
	bom.Dependencies = &dependencyGraph

	if options.IncludeStd {
		err = cliutil.AddStdComponent(bom)
		if err != nil {
			return err
		}
	}

	return cliutil.WriteBOM(bom, options.OutputOptions)
}
