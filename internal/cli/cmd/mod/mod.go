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
	"encoding/base64"
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

func Exec(modOptions ModOptions) error {
	err := modOptions.Validate()
	if err != nil {
		return err
	}

	// Cheap trick to make Go download all required modules in the module graph
	// without modifying go.sum (as `go mod download` would do).
	err = gocmd.ModWhy(modOptions.ModuleDir, []string{"github.com/CycloneDX/cyclonedx-gomod"}, io.Discard)
	if err != nil {
		return fmt.Errorf("downloading modules failed: %w", err)
	}

	modules, err := gomod.GetModules(modOptions.ModuleDir, modOptions.IncludeTest)
	if err != nil {
		return fmt.Errorf("failed to enumerate modules: %w", err)
	}

	modules[0].Version, err = gomod.GetModuleVersion(modules[0].Dir)
	if err != nil {
		log.Warn().Err(err).Msg("failed to determine version of main module")
	}

	sbom.NormalizeVersions(modules, modOptions.NoVersionPrefix)

	mainComponent, err := modconv.ToComponent(modules[0],
		modconv.WithComponentType(cdx.ComponentType(modOptions.ComponentType)),
		withLicenses(modOptions.ResolveLicenses),
		modconv.WithScope(""), // Main component can't have a scope
	)
	if err != nil {
		return fmt.Errorf("failed to convert main module: %w", err)
	}

	components, err := modconv.ToComponents(modules[1:],
		withModuleHashes(),
		withLicenses(modOptions.ResolveLicenses),
	)
	if err != nil {
		return fmt.Errorf("failed to convert modules: %w", err)
	}

	dependencyGraph := sbom.BuildDependencyGraph(modules)

	bom := cdx.NewBOM()

	err = cliutil.SetSerialNumber(bom, modOptions.SBOMOptions)
	if err != nil {
		return err
	}

	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}

	err = cliutil.AddCommonMetadata(bom, modOptions.SBOMOptions)
	if err != nil {
		return err
	}

	bom.Components = &components
	bom.Dependencies = &dependencyGraph

	if modOptions.IncludeStd {
		err = addStdComponent(bom)
		if err != nil {
			return err
		}
	}

	return cliutil.WriteBOM(bom, modOptions.OutputOptions)
}

func withLicenses(enabled bool) modconv.Option {
	return func(m gomod.Module, c *cdx.Component) error {
		if enabled {
			return modconv.WithLicenses()(m, c)
		}
		return nil
	}
}

func withModuleHashes() modconv.Option {
	return func(m gomod.Module, c *cdx.Component) error {
		if m.Main {
			// We currently don't have an accurate way of hashing the main module, as it may contain
			// files that are .gitignore'd and thus not part of the hashes in Go's sumdb.
			log.Debug().Str("module", m.Coordinates()).Msg("not calculating hash for main module")
			return nil
		}

		if m.Vendored {
			// Go's vendoring mechanism doesn't copy all files that make up a module to the vendor dir.
			// Hashing vendored modules thus won't result in the expected hash, probably causing more
			// confusion than anything else.
			log.Debug().Str("module", m.Coordinates()).Msg("not calculating hash for vendored module")
			return nil
		}

		log.Debug().Str("module", m.Coordinates()).Msg("calculating module hash")
		h1, err := m.Hash()
		if err != nil {
			return fmt.Errorf("failed to calculate module hash: %w", err)
		}

		h1Bytes, err := base64.StdEncoding.DecodeString(h1[3:])
		if err != nil {
			return fmt.Errorf("failed to base64 decode module hash: %w", err)
		}

		c.Hashes = &[]cdx.Hash{
			{Algorithm: cdx.HashAlgoSHA256, Value: fmt.Sprintf("%x", h1Bytes)},
		}

		return nil
	}
}

func addStdComponent(bom *cdx.BOM) error {
	stdComponent, err := sbom.BuildStdComponent()
	if err != nil {
		return fmt.Errorf("failed to build std component: %w", err)
	}

	*bom.Components = append(*bom.Components, *stdComponent)

	// Add std to dependency graph
	stdDependency := cdx.Dependency{Ref: stdComponent.BOMRef}
	*bom.Dependencies = append(*bom.Dependencies, stdDependency)

	// Add std as dependency of main module
	for i, dependency := range *bom.Dependencies {
		if dependency.Ref == bom.Metadata.Component.BOMRef {
			if dependency.Dependencies == nil {
				(*bom.Dependencies)[i].Dependencies = &[]cdx.Dependency{stdDependency}
			} else {
				*dependency.Dependencies = append(*dependency.Dependencies, stdDependency)
			}
			break
		}
	}

	return nil
}
