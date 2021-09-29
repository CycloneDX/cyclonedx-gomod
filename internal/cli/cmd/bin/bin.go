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
	"encoding/base64"
	"flag"
	"fmt"
	"path/filepath"

	cdx "github.com/CycloneDX/cyclonedx-go"
	cliUtil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modConv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
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

Although the binary is never executed, it must be executable.
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

			cliUtil.ConfigureLogger(options.LogOptions)

			return Exec(options)
		},
	}
}

func Exec(options Options) error {
	err := options.Validate()
	if err != nil {
		return err
	}

	modules, hashes, err := gomod.LoadModulesFromBinary(options.BinaryPath)
	if err != nil {
		return fmt.Errorf("failed to extract modules: %w", err)
	} else if len(modules) == 0 {
		return fmt.Errorf("failed to parse modules from %s", options.BinaryPath)
	}

	if options.Version != "" {
		modules[0].Version = options.Version
	}

	// If we want to resolve licenses, we have to download the modules first
	if options.ResolveLicenses {
		err = downloadModules(modules, hashes)
		if err != nil {
			return fmt.Errorf("failed to download modules: %w", err)
		}
	}

	// Make all modules a direct dependency of the main module
	for i := 1; i < len(modules); i++ {
		modules[0].Dependencies = append(modules[0].Dependencies, &modules[i])
	}

	// Convert main module
	mainComponent, err := modConv.ToComponent(modules[0],
		modConv.WithComponentType(cdx.ComponentTypeApplication),
		modConv.WithLicenses(options.ResolveLicenses),
	)
	if err != nil {
		return fmt.Errorf("failed to convert main module: %w", err)
	}

	// Convert the other modules
	components, err := modConv.ToComponents(modules[1:],
		modConv.WithLicenses(options.ResolveLicenses),
		withModuleHashes(hashes),
	)
	if err != nil {
		return fmt.Errorf("failed to convert modules: %w", err)
	}

	binaryProperties, err := createBinaryProperties(options.BinaryPath)
	if err != nil {
		return fmt.Errorf("failed to create binary properties")
	}

	bom := cdx.NewBOM()

	err = cliUtil.SetSerialNumber(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to set serial number: %w", err)
	}

	bom.Metadata = &cdx.Metadata{
		Component:  mainComponent,
		Properties: &binaryProperties,
	}
	err = cliUtil.AddCommonMetadata(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to add common metadata")
	}

	bom.Components = &components
	dependencyGraph := sbom.BuildDependencyGraph(modules)
	bom.Dependencies = &dependencyGraph
	bom.Compositions = createCompositions(*mainComponent, components)

	return cliUtil.WriteBOM(bom, options.OutputOptions)
}

func withModuleHashes(hashes map[string]string) modConv.Option {
	return func(m gomod.Module, c *cdx.Component) error {
		h1, ok := hashes[m.Coordinates()]
		if !ok {
			return nil
		}

		h1Bytes, err := base64.StdEncoding.DecodeString(h1[3:])
		if err != nil {
			return fmt.Errorf("failed to base64 decode h1 hash: %w", err)
		}

		c.Hashes = &[]cdx.Hash{
			{
				Algorithm: cdx.HashAlgoSHA256,
				Value:     fmt.Sprintf("%x", h1Bytes),
			},
		}

		return nil
	}
}

func createBinaryProperties(binaryPath string) ([]cdx.Property, error) {
	binaryHashes, err := sbom.CalculateFileHashes(binaryPath,
		cdx.HashAlgoMD5, cdx.HashAlgoSHA1, cdx.HashAlgoSHA256, cdx.HashAlgoSHA384, cdx.HashAlgoSHA512)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate binary hashes: %w", err)
	}

	properties := []cdx.Property{
		sbom.NewProperty("binary:name", filepath.Base(binaryPath)),
	}
	for _, hash := range binaryHashes {
		properties = append(properties, sbom.NewProperty(fmt.Sprintf("binary:hash:%s", hash.Algorithm), hash.Value))
	}

	return properties, nil
}

func createCompositions(mainComponent cdx.Component, components []cdx.Component) *[]cdx.Composition {
	compositions := make([]cdx.Composition, 0)

	// We know all components that the main component directly or indirectly depends on,
	// thus the dependencies of it are considered complete.
	compositions = append(compositions, cdx.Composition{
		Aggregate: cdx.CompositionAggregateComplete,
		Dependencies: &[]cdx.BOMReference{
			cdx.BOMReference(mainComponent.BOMRef),
		},
	})

	// The exact relationships between the dependencies are unknown
	dependencyRefs := make([]cdx.BOMReference, 0, len(components))
	for _, component := range components {
		dependencyRefs = append(dependencyRefs, cdx.BOMReference(component.BOMRef))
	}
	compositions = append(compositions, cdx.Composition{
		Aggregate:    cdx.CompositionAggregateUnknown,
		Dependencies: &dependencyRefs,
	})

	return &compositions
}

func downloadModules(modules []gomod.Module, hashes map[string]string) error {
	// When modules are replaced, only download the replacement.
	modulesToDownload := make([]gomod.Module, len(modules))
	for i, module := range modules {
		if module.Replace != nil {
			modulesToDownload[i] = *modules[i].Replace
		} else {
			modulesToDownload[i] = modules[i]
		}
	}

	downloads, err := gomod.Download(modulesToDownload)
	if err != nil {
		return err
	}

	for i, download := range downloads {
		if download.Error != "" {
			log.Warn().
				Str("module", download.Coordinates()).
				Str("reason", download.Error).
				Msg("module download failed")
			continue
		}

		module := matchModule(modules, download.Coordinates())
		if module == nil {
			log.Warn().
				Str("module", download.Coordinates()).
				Msg("downloaded module not found")
			continue
		}

		// Check that the hash of the downloaded module matches
		// the one found in the binary. We want to report the version
		// for the *exact* module version or nothing at all.
		hash, ok := hashes[download.Coordinates()]
		if ok {
			if hash != download.Sum {
				log.Warn().
					Str("binaryHash", hash).
					Str("downloadHash", download.Sum).
					Str("module", download.Coordinates()).
					Msg("module hash mismatch")
				continue
			}
		}

		log.Debug().
			Str("module", download.Coordinates()).
			Msg("module downloaded")

		module.Dir = downloads[i].Dir
	}

	return nil
}

func matchModule(modules []gomod.Module, coordinates string) *gomod.Module {
	for i, module := range modules {
		if module.Replace != nil && coordinates == module.Replace.Coordinates() {
			return modules[i].Replace
		}
		if coordinates == module.Coordinates() {
			return &modules[i]
		}
	}

	return nil
}
